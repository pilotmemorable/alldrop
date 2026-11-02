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
    var downloadsDir: URL {
        FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0]
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

    // MARK: files opened in / shared to AllDrop

    func stageIncoming(_ urls: [URL]) {
        var copies: [URL] = []
        let dir = shareDir.appendingPathComponent(UUID().uuidString, isDirectory: true)
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        for url in urls {
            let scoped = url.startAccessingSecurityScopedResource()
            defer { if scoped { url.stopAccessingSecurityScopedResource() } }
            let dest = dir.appendingPathComponent(url.lastPathComponent)
            do {
                if url.path.hasPrefix(shareDir.path) || url.path.hasPrefix(FileManager.default.temporaryDirectory.path) {
                    try FileManager.default.moveItem(at: url, to: dest)
                } else {
                    try FileManager.default.copyItem(at: url, to: dest)
                }
                copies.append(dest)
            } catch {
                NSLog("AllDrop could not read shared file: %@", error.localizedDescription)
            }
        }
        queue.async {
            self.pendingShares.append(contentsOf: copies)
            self.flushShares()
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
