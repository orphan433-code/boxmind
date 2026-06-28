import Foundation

struct BookmarkFolder: Codable, Identifiable, Equatable, Hashable, Sendable {
    let id: String
    let userID: String
    let name: String
    let createdAt: Date
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case userID = "user_id"
        case name
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}

struct CreateFolderRequest: Encodable, Sendable {
    let name: String
}

struct AssignBookmarkFolderRequest: Encodable, Sendable {
    let folderID: String?

    enum CodingKeys: String, CodingKey {
        case folderID = "folder_id"
    }
}
