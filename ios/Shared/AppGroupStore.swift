import Foundation

/// Shared storage between the main app and the Share Extension.
/// The app writes the session here after login; the extension reads it to call the API.
enum AppGroupStore {
    static let suiteName = "group.link.boxmind.app"

    private enum Keys {
        static let accessToken = "shared_access_token"
        static let apiBaseURL = "shared_api_base_url"
    }

    private static var defaults: UserDefaults? {
        UserDefaults(suiteName: suiteName)
    }

    static func saveSession(accessToken: String, apiBaseURL: String) {
        defaults?.set(accessToken, forKey: Keys.accessToken)
        defaults?.set(apiBaseURL, forKey: Keys.apiBaseURL)
    }

    static func clearSession() {
        defaults?.removeObject(forKey: Keys.accessToken)
        defaults?.removeObject(forKey: Keys.apiBaseURL)
    }

    static var accessToken: String? {
        defaults?.string(forKey: Keys.accessToken)
    }

    static var apiBaseURL: String? {
        defaults?.string(forKey: Keys.apiBaseURL)
    }
}
