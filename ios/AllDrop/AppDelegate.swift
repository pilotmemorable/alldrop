import UIKit

@main
final class AppDelegate: UIResponder, UIApplicationDelegate {
    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
        EngineHost.shared.start()
        return true
    }

    func application(_ application: UIApplication,
                     configurationForConnecting connectingSceneSession: UISceneSession,
                     options: UIScene.ConnectionOptions) -> UISceneConfiguration {
        UISceneConfiguration(name: "Default", sessionRole: connectingSceneSession.role)
    }
}

final class SceneDelegate: UIResponder, UIWindowSceneDelegate {
    var window: UIWindow?

    func scene(_ scene: UIScene, willConnectTo session: UISceneSession,
               options connectionOptions: UIScene.ConnectionOptions) {
        guard let windowScene = scene as? UIWindowScene else { return }
        let window = UIWindow(windowScene: windowScene)
        window.rootViewController = WebViewController()
        window.makeKeyAndVisible()
        self.window = window
        handle(connectionOptions.urlContexts)
    }

    /// "Open in AllDrop" / share sheet from Photos, Files, Mail ...
    func scene(_ scene: UIScene, openURLContexts URLContexts: Set<UIOpenURLContext>) {
        handle(URLContexts)
    }

    func sceneWillEnterForeground(_ scene: UIScene) {
        EngineHost.shared.start()
    }

    func sceneDidEnterBackground(_ scene: UIScene) {
        EngineHost.shared.enteredBackground()
    }

    private func handle(_ contexts: Set<UIOpenURLContext>) {
        let urls = contexts.map { $0.url }.filter { $0.isFileURL }
        if !urls.isEmpty { EngineHost.shared.stageIncoming(urls) }
    }
}
