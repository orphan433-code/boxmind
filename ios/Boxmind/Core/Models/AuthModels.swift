import Foundation

struct LoginRequest: Encodable, Sendable {
    let email: String
}

struct VerifyLoginRequest: Encodable, Sendable {
    let email: String
    let code: String
}

struct LoginResponse: Decodable, Sendable {
    let message: String
}

struct AuthTokens: Decodable, Sendable {
    let accessToken: String

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
    }
}

struct VerifyLoginResponse: Decodable, Sendable {
    let tokens: AuthTokens
    let user: User
}

struct APIErrorResponse: Decodable, Sendable {
    let error: String
}
