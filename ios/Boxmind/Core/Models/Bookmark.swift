import Foundation

struct Bookmark: Codable, Identifiable, Equatable, Sendable {
    let id: String
    let userID: String
    let url: String
    let title: String
    let description: String
    let imageURL: String
    let category: String
    let tags: [String]
    let enriched: Bool
    let createdAt: Date
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userID = "user_id"
        case url
        case title
        case description
        case imageURL = "image_url"
        case category
        case tags
        case enriched
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct CreateBookmarkRequest: Encodable, Sendable {
    let url: String
}
