import Foundation

enum APIConfig {
    /// Production API. Override in scheme env `BOXMIND_API_URL` for local dev.
    static var baseURL: URL {
        if let override = ProcessInfo.processInfo.environment["BOXMIND_API_URL"],
           let url = URL(string: override) {
            return url
        }
        return URL(string: "https://api.boxmind.link/api/v1")!
    }
}
