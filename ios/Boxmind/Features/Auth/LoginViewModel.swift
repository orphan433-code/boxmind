import Foundation
import Observation

@Observable
@MainActor
final class LoginViewModel {
    enum Step {
        case email
        case code
    }

    var email = ""
    var code = ""
    var step: Step = .email
    var isLoading = false
    var errorMessage: String?

    func sendCode(session: SessionStore) async {
        errorMessage = nil
        isLoading = true
        defer { isLoading = false }

        let trimmedEmail = email.trimmingCharacters(in: .whitespacesAndNewlines)
        do {
            try await session.requestLogin(email: trimmedEmail)
            email = trimmedEmail
            step = .code
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func verifyCode(session: SessionStore) async {
        errorMessage = nil
        isLoading = true
        defer { isLoading = false }

        do {
            try await session.verifyLogin(
                email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                code: code.trimmingCharacters(in: .whitespacesAndNewlines)
            )
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func backToEmail() {
        step = .email
        code = ""
        errorMessage = nil
    }
}
