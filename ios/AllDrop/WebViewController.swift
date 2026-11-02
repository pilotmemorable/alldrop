import UIKit
import WebKit
import QuickLook

/// Shows the AllDrop interface served by the in-app engine on 127.0.0.1.
final class WebViewController: UIViewController, WKNavigationDelegate, WKUIDelegate, QLPreviewControllerDataSource {
    private var webView: WKWebView!
    private let statusLabel = UILabel()
    private var loadedPort = 0
    private var timer: Timer?
    private var previewURL: URL?

    override func loadView() {
        let config = WKWebViewConfiguration()
        let content = WKUserContentController()
        content.add(MessageProxy(owner: self), name: "alldrop")
        // Same bridge the Android app provides, so the web UI can open files natively.
        let bridge = """
        window.AllDropNative = {
          platform: 'ios',
          openFile: function (p) { window.webkit.messageHandlers.alldrop.postMessage({ cmd: 'open', path: p }); },
          openDownloads: function () { window.webkit.messageHandlers.alldrop.postMessage({ cmd: 'folder' }); }
        };
        """
        content.addUserScript(WKUserScript(source: bridge, injectionTime: .atDocumentStart, forMainFrameOnly: true))
        config.userContentController = content
        config.allowsInlineMediaPlayback = true
        config.websiteDataStore = .default()

        webView = WKWebView(frame: .zero, configuration: config)
        webView.navigationDelegate = self
        webView.uiDelegate = self
        webView.allowsBackForwardNavigationGestures = false
        webView.scrollView.contentInsetAdjustmentBehavior = .automatic
        webView.isOpaque = false
        webView.backgroundColor = UIColor(named: "LaunchBackground")
        webView.alpha = 0
        view = webView
    }

    override func viewDidLoad() {
        super.viewDidLoad()
        statusLabel.text = NSLocalizedString("Starting…", comment: "")
        statusLabel.textColor = .secondaryLabel
        statusLabel.textAlignment = .center
        statusLabel.translatesAutoresizingMaskIntoConstraints = false
        view.addSubview(statusLabel)
        NSLayoutConstraint.activate([
            statusLabel.centerXAnchor.constraint(equalTo: view.centerXAnchor),
            statusLabel.centerYAnchor.constraint(equalTo: view.centerYAnchor),
        ])
        // Watch the engine: it restarts when the app returns from the background.
        timer = Timer.scheduledTimer(withTimeInterval: 0.3, repeats: true) { [weak self] _ in
            self?.checkEngine()
        }
    }

    private func checkEngine() {
        let port = EngineHost.shared.port
        guard port > 0 else { return }
        if port != loadedPort {
            loadedPort = port
            let url = URL(string: "http://127.0.0.1:\(port)/?ot=\(EngineHost.shared.token)")!
            webView.load(URLRequest(url: url))
        }
    }

    func webView(_ webView: WKWebView, didFinish navigation: WKNavigation!) {
        statusLabel.isHidden = true
        UIView.animate(withDuration: 0.2) { webView.alpha = 1 }
    }

    func webView(_ webView: WKWebView, didFail navigation: WKNavigation!, withError error: Error) {
        loadedPort = 0 // retry on the next check
    }

    func webView(_ webView: WKWebView, didFailProvisionalNavigation navigation: WKNavigation!, withError error: Error) {
        loadedPort = 0
    }

    /// Only our own engine is shown inside the app; other links open in Safari.
    func webView(_ webView: WKWebView, decidePolicyFor navigationAction: WKNavigationAction,
                 decisionHandler: @escaping (WKNavigationActionPolicy) -> Void) {
        guard let url = navigationAction.request.url else { return decisionHandler(.cancel) }
        if url.host == "127.0.0.1" || url.scheme == "about" || url.scheme == "blob" {
            decisionHandler(.allow)
            return
        }
        if url.scheme == "http" || url.scheme == "https" {
            UIApplication.shared.open(url)
        }
        decisionHandler(.cancel)
    }

    // MARK: native actions requested by the page

    fileprivate func handle(message: Any) {
        guard let body = message as? [String: Any], let cmd = body["cmd"] as? String else { return }
        switch cmd {
        case "open":
            guard let path = body["path"] as? String else { return }
            let url = URL(fileURLWithPath: path)
            // Only files inside our own Documents folder.
            guard url.standardizedFileURL.path.hasPrefix(EngineHost.shared.downloadsDir.standardizedFileURL.path),
                  FileManager.default.fileExists(atPath: url.path) else { return }
            previewURL = url
            let ql = QLPreviewController()
            ql.dataSource = self
            present(ql, animated: true)
        case "folder":
            // Opens the Files app at AllDrop's folder.
            let path = EngineHost.shared.downloadsDir.path
            if let url = URL(string: "shareddocuments://" + (path.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? path)) {
                UIApplication.shared.open(url)
            }
        default:
            break
        }
    }

    func numberOfPreviewItems(in controller: QLPreviewController) -> Int {
        previewURL == nil ? 0 : 1
    }

    func previewController(_ controller: QLPreviewController, previewItemAt index: Int) -> QLPreviewItem {
        previewURL! as NSURL
    }
}

/// WKUserContentController retains its handlers; this proxy avoids a cycle.
private final class MessageProxy: NSObject, WKScriptMessageHandler {
    weak var owner: WebViewController?
    init(owner: WebViewController) { self.owner = owner }
    func userContentController(_ userContentController: WKUserContentController, didReceive message: WKScriptMessage) {
        owner?.handle(message: message.body)
    }
}
