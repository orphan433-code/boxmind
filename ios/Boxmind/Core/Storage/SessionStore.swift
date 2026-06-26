import Foundation
import Observation

@Observable
@MainActor
final class SessionStore {
    private enum Keys {
        static let accessToken = "access_token"
        static let userEmail = "user_email"
    }

    private(set) var accessToken: String?
    private(set) var userEmail: String?

    var isAuthenticated: Bool {
        accessToken != nil
    }

    private let api = APIClient()

    init() {
        accessToken = KeychainStore.load(account: Keys.accessToken)
        userEmail = KeychainStore.load(account: Keys.userEmail)
        syncSharedSession()
    }

    func requestLogin(email: String) async throws {
        try await api.requestLogin(email: email)
    }

    func verifyLogin(email: String, code: String) async throws {
        let result = try await api.verifyLogin(email: email, code: code)
        accessToken = result.tokens.accessToken
        userEmail = result.user.email
        try KeychainStore.save(result.tokens.accessToken, account: Keys.accessToken)
        try KeychainStore.save(result.user.email, account: Keys.userEmail)
        syncSharedSession()
    }

    func logout() {
        accessToken = nil
        userEmail = nil
        KeychainStore.delete(account: Keys.accessToken)
        KeychainStore.delete(account: Keys.userEmail)
        AppGroupStore.clearSession()
    }

    func makeAPIClient() -> APIClient {
        APIClient()
    }

    /// Mirrors the current token into the shared App Group so the Share Extension can post bookmarks.
    private func syncSharedSession() {
        if let accessToken {
            AppGroupStore.saveSession(
                accessToken: accessToken,
                apiBaseURL: APIConfig.baseURL.absoluteString
            )
        } else {
            AppGroupStore.clearSession()
        }
    }
}
