import Foundation

/// Thin HTTP client for the Boxmind REST API.
struct APIClient: Sendable {
    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init(baseURL: URL = APIConfig.baseURL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let value = try container.decode(String.self)

            let formatters = [
                ISO8601DateFormatter(),
                {
                    let f = ISO8601DateFormatter()
                    f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
                    return f
                }()
            ]

            for formatter in formatters {
                if let date = formatter.date(from: value) {
                    return date
                }
            }

            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Invalid date: \(value)")
        }
        self.decoder = decoder

        let encoder = JSONEncoder()
        self.encoder = encoder
    }

    func requestLogin(email: String) async throws {
        let _: LoginResponse = try await post("/auth/login", body: LoginRequest(email: email))
    }

    func verifyLogin(email: String, code: String) async throws -> VerifyLoginResponse {
        try await post("/auth/verify", body: VerifyLoginRequest(email: email, code: code))
    }

    func listBookmarks(token: String) async throws -> [Bookmark] {
        try await get("/bookmarks", token: token)
    }

    func createBookmark(token: String, url: String) async throws -> Bookmark {
        try await post("/bookmarks", body: CreateBookmarkRequest(url: url), token: token)
    }

    func deleteBookmark(token: String, id: String) async throws {
        try await send(path: "/bookmarks/\(id)", method: "DELETE", token: token)
    }

    func listFolders(token: String) async throws -> [BookmarkFolder] {
        try await get("/folders", token: token)
    }

    func createFolder(token: String, name: String) async throws -> BookmarkFolder {
        try await post("/folders", body: CreateFolderRequest(name: name), token: token)
    }

    func deleteFolder(token: String, id: String) async throws {
        try await send(path: "/folders/\(id)", method: "DELETE", token: token)
    }

    func assignBookmarkFolder(token: String, bookmarkID: String, folderID: String?) async throws -> Bookmark {
        try await request(
            path: "/bookmarks/\(bookmarkID)/folder",
            method: "PUT",
            body: try JSONEncoder().encode(AssignBookmarkFolderRequest(folderID: folderID)),
            token: token
        )
    }

    // MARK: - HTTP helpers

    /// Joins `baseURL` (which may include a path like `/api/v1`) with an endpoint path.
    /// `URL(string:relativeTo:)` drops the base path when the endpoint starts with "/",
    /// so we append components manually to keep `/api/v1` intact.
    private func makeURL(_ path: String) -> URL? {
        let trimmed = path.hasPrefix("/") ? String(path.dropFirst()) : path
        guard var components = URLComponents(url: baseURL, resolvingAgainstBaseURL: false) else {
            return nil
        }
        var basePath = components.path
        if basePath.hasSuffix("/") {
            basePath.removeLast()
        }
        components.path = basePath + "/" + trimmed
        return components.url
    }

    private func get<T: Decodable>(_ path: String, token: String) async throws -> T {
        try await request(path: path, method: "GET", token: token)
    }

    private func post<T: Decodable, Body: Encodable>(
        _ path: String,
        body: Body,
        token: String? = nil
    ) async throws -> T {
        let data = try encoder.encode(body)
        return try await request(path: path, method: "POST", body: data, token: token)
    }

    private func send(path: String, method: String, token: String? = nil, body: Data? = nil) async throws {
        guard let url = makeURL(path) else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        if let token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        request.httpBody = body

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: request)
        } catch {
            throw APIError.transport(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw APIError.transport(URLError(.badServerResponse))
        }

        if http.statusCode == 401 {
            throw APIError.unauthorized
        }

        if !(200 ... 299).contains(http.statusCode) {
            if let apiError = try? decoder.decode(APIErrorResponse.self, from: data) {
                throw APIError.server(message: apiError.error)
            }
            throw APIError.server(message: "Ошибка сервера (\(http.statusCode))")
        }
    }

    private func request<T: Decodable>(
        path: String,
        method: String,
        body: Data? = nil,
        token: String? = nil
    ) async throws -> T {
        guard let url = makeURL(path) else {
            throw APIError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        if let token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        request.httpBody = body

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: request)
        } catch {
            throw APIError.transport(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw APIError.transport(URLError(.badServerResponse))
        }

        if http.statusCode == 401 {
            throw APIError.unauthorized
        }

        if !(200 ... 299).contains(http.statusCode) {
            if let apiError = try? decoder.decode(APIErrorResponse.self, from: data) {
                throw APIError.server(message: apiError.error)
            }
            throw APIError.server(message: "Ошибка сервера (\(http.statusCode))")
        }

        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw APIError.decoding
        }
    }
}
