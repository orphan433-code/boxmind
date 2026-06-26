import Foundation

enum APIError: LocalizedError, Sendable {
    case invalidURL
    case unauthorized
    case server(message: String)
    case decoding
    case transport(Error)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Некорректный URL"
        case .unauthorized:
            return "Сессия истекла — войди снова"
        case .server(let message):
            return message
        case .decoding:
            return "Не удалось разобрать ответ сервера"
        case .transport(let error):
            return error.localizedDescription
        }
    }
}
