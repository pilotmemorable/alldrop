import Foundation
import UIKit
import Adkit

/// Runs the shared AllDrop engine (Go, built with gomobile) inside the app.
/// The engine serves the interface on 127.0.0.1 and talks to other devices
/// over TLS on the local network.
final class EngineHost: NSObject, AdkitListenerProtocol {
    static let shared = EngineHost()

    /// Proves to the engine that requests come from this app's web view,
    /// not from another app on the phone.
    let token: String = {
        var bytes = [UInt8](repeating: 0, count: 16)
        _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
        return bytes.map { String(format: "%02x", $0) }.joined()
    }()

    private(set) var port: Int = 0
    private(set) var busy = false
    private var pendingShares: [URL] = []
    private var backgroundTask: UIBackgroundTaskIdentifier = .invalid
    private let queue = DispatchQueue(label: "app.alldrop.engine")

    /// Received files go to Documents, which the Files app shows as "AllDrop".
    var defaultDownloadsDir: URL {
        FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0]
    }

    /// The folder the person chose in Settings (kept as a bookmark so it
    /// survives restarts), or Documents.
    var downloadsDir: URL { customDir ?? defaultDownloadsDir }

    private static let bookmarkKey = "saveDirBookmark"
    private var customDir: URL? = EngineHost.loadSaveDir()
    private var scopedURL: URL?

    private static func loadSaveDir() -> URL? {
        guard let data = UserDefaults.standard.data(forKey: bookmarkKey) else { return nil }
        var stale = false
        guard let url = try? URL(resolvingBookmarkData: data, options: [], relativeTo: nil, bookmarkDataIsStale: &stale) else {
            return nil
        }
        guard url.startAccessingSecurityScopedResource() else { return nil }
        if stale, let fresh = try? url.bookmarkData(options: [], includingResourceValuesForKeys: nil, relativeTo: nil) {
            UserDefaults.standard.set(fresh, forKey: bookmarkKey)
        }
        return url
    }

    /// Saves received files in `url` (a folder picked with the Files app).
    func setSaveDir(_ url: URL) {
        let opened = url.startAccessingSecurityScopedResource()
        guard let data = try? url.bookmarkData(options: [], includingResourceValuesForKeys: nil, relativeTo: nil) else {
            if opened { url.stopAccessingSecurityScopedResource() }
            return
        }
        UserDefaults.standard.set(data, forKey: Self.bookmarkKey)
        if let old = customDir, old != url { old.stopAccessingSecurityScopedResource() }
        customDir = url
        _ = AdkitSetDownloadDir(url.path)
    }

    /// Goes back to the app's own Documents folder.
    func resetSaveDir() {
        UserDefaults.standard.removeObject(forKey: Self.bookmarkKey)
        customDir?.stopAccessingSecurityScopedResource()
        customDir = nil
        _ = AdkitSetDownloadDir(defaultDownloadsDir.path)
    }

    private var dataDir: URL {
        let base = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask)[0]
        return base.appendingPathComponent("engine", isDirectory: true)
    }

    private var shareDir: URL {
        FileManager.default.temporaryDirectory.appendingPathComponent("share", isDirectory: true)
    }

    // MARK: lifecycle

    func start() {
        let name = UIDevice.current.name
        let platform = UIDevice.current.userInterfaceIdiom == .pad ? "ipad" : "ios"
        queue.async {
            if AdkitPort() > 0 {
                self.port = AdkitPort()
                return
            }
            try? FileManager.default.createDirectory(at: self.dataDir, withIntermediateDirectories: true)
            try? FileManager.default.createDirectory(at: self.downloadsDir, withIntermediateDirectories: true)
            let p = AdkitStart(self.dataDir.path, self.downloadsDir.path, name, platform,
                               self.token, Self.localIPv4().joined(separator: ","), self)
            if p > 0 {
                self.port = p
                self.flushShares()
            } else {
                NSLog("AllDrop engine failed: %@", AdkitLastError())
            }
        }
    }

    func stop() {
        queue.async {
            AdkitStop()
            self.port = 0
        }
    }

    /// iOS suspends apps in the background. Finish a running transfer if
    /// we can, then stop the engine; it restarts when the app comes back.
    func enteredBackground() {
        endBackgroundTask()
        backgroundTask = UIApplication.shared.beginBackgroundTask(withName: "AllDrop transfer") { [weak self] in
            self?.stop()
            self?.endBackgroundTask()
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 2) { [weak self] in
            guard let self = self else { return }
            if UIApplication.shared.applicationState == .background && !self.busy {
                self.stop()
                self.endBackgroundTask()
            }
        }
    }

    private func endBackgroundTask() {
        if backgroundTask != .invalid {
            UIApplication.shared.endBackgroundTask(backgroundTask)
            backgroundTask = .invalid
        }
    }

    // MARK: engine events

    func onEvent(_ kind: String?, a: String?, b: String?, c: String?) {
        DispatchQueue.main.async {
            switch kind ?? "" {
            case "OFFER":
                UINotificationFeedbackGenerator().notificationOccurred(.warning)
            case "ACTIVE":
                self.busy = (a == "1")
                UIApplication.shared.isIdleTimerDisabled = self.busy
                if !self.busy && UIApplication.shared.applicationState == .background {
                    self.stop()
                    self.endBackgroundTask()
                }
            case "SAVED":
                UINotificationFeedbackGenerator().notificationOccurred(.success)
            default:
                break
            }
        }
    }

    // MARK: files opened in / shared to / picked in AllDrop

    /// A fresh temporary folder for copies that will be sent.
    func newShareFolder() -> URL {
        let dir = shareDir.appendingPathComponent(UUID().uuidString, isDirectory: true)
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir
    }

    static func uniqueURL(in dir: URL, name: String) -> URL {
        var candidate = dir.appendingPathComponent(name)
        let base = (name as NSString).deletingPathExtension
        let ext = (name as NSString).pathExtension
        var n = 2
        while FileManager.default.fileExists(atPath: candidate.path) {
            candidate = dir.appendingPathComponent(ext.isEmpty ? "\(base) (\(n))" : "\(base) (\(n)).\(ext)")
            n += 1
        }
        return candidate
    }

    /// Files already copied into our temporary folder: queue them for sending.
    func stageReady(_ urls: [URL]) {
        guard !urls.isEmpty else { return }
        queue.async {
            self.pendingShares.append(contentsOf: urls)
            self.flushShares()
        }
    }

    /// Files or whole folders from other apps or the document picker (copied first,
    /// off the main thread so large folders do not freeze the screen).
    func stageIncoming(_ urls: [URL]) {
        let dir = newShareFolder()
        let tmp = FileManager.default.temporaryDirectory.path
        let share = shareDir.path
        DispatchQueue.global(qos: .userInitiated).async {
            var copies: [URL] = []
            for url in urls {
                let scoped = url.startAccessingSecurityScopedResource()
                defer { if scoped { url.stopAccessingSecurityScopedResource() } }
                let dest = EngineHost.uniqueURL(in: dir, name: url.lastPathComponent)
                do {
                    if url.path.hasPrefix(share) || url.path.hasPrefix(tmp) {
                        try FileManager.default.moveItem(at: url, to: dest)
                    } else {
                        try FileManager.default.copyItem(at: url, to: dest) // folders: copied with everything inside
                    }
                    copies.append(dest)
                } catch {
                    NSLog("AllDrop could not read shared file: %@", error.localizedDescription)
                }
            }
            self.stageReady(copies)
        }
    }

    private func flushShares() {
        guard port > 0, !pendingShares.isEmpty else { return }
        let joined = pendingShares.map { $0.path }.joined(separator: "\n")
        pendingShares.removeAll()
        _ = AdkitStage(joined, true)
    }

    // MARK: helpers

    static func localIPv4() -> [String] {
        var out: [String] = []
        var ifaddr: UnsafeMutablePointer<ifaddrs>?
        guard getifaddrs(&ifaddr) == 0, let first = ifaddr else { return out }
        defer { freeifaddrs(ifaddr) }
        var ptr: UnsafeMutablePointer<ifaddrs>? = first
        while let p = ptr {
            let flags = Int32(p.pointee.ifa_flags)
            if let sa = p.pointee.ifa_addr, sa.pointee.sa_family == UInt8(AF_INET),
               (flags & IFF_UP) != 0, (flags & IFF_LOOPBACK) == 0 {
                var host = [CChar](repeating: 0, count: Int(NI_MAXHOST))
                if getnameinfo(sa, socklen_t(sa.pointee.sa_len), &host, socklen_t(host.count), nil, 0, NI_NUMERICHOST) == 0 {
                    let ip = String(cString: host)
                    if !ip.hasPrefix("169.254.") { out.append(ip) }
                }
            }
            ptr = p.pointee.ifa_next
        }
        return out
    }
}
