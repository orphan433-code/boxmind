import Foundation
import Observation

@Observable
@MainActor
final class BookmarksViewModel {
    var bookmarks: [Bookmark] = []
    var folders: [BookmarkFolder] = []
    var isLoading = false
    var errorMessage: String?

    private var pollingTask: Task<Void, Never>?
    private let api = APIClient()

    func load(session: SessionStore) async {
        guard let token = session.accessToken else { return }

        if bookmarks.isEmpty && folders.isEmpty {
            isLoading = true
        }
        defer { isLoading = false }

        do {
            async let bookmarkData = api.listBookmarks(token: token)
            async let folderData = api.listFolders(token: token)
            bookmarks = try await bookmarkData
            folders = try await folderData
            startPollingIfNeeded(session: session)
        } catch let error as APIError {
            if case .unauthorized = error {
                session.logout()
            } else {
                errorMessage = error.localizedDescription
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func createFolder(name: String, session: SessionStore) async {
        guard let token = session.accessToken else { return }

        do {
            let folder = try await api.createFolder(token: token, name: name)
            folders.append(folder)
            folders.sort { $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func deleteFolder(_ folder: BookmarkFolder, session: SessionStore) async {
        guard let token = session.accessToken else { return }

        do {
            try await api.deleteFolder(token: token, id: folder.id)
            folders.removeAll { $0.id == folder.id }
            bookmarks = bookmarks.map { bookmark in
                guard bookmark.folderID == folder.id else { return bookmark }
                return Bookmark(
                    id: bookmark.id,
                    userID: bookmark.userID,
                    url: bookmark.url,
                    title: bookmark.title,
                    description: bookmark.description,
                    imageURL: bookmark.imageURL,
                    category: bookmark.category,
                    tags: bookmark.tags,
                    folderID: nil,
                    enriched: bookmark.enriched,
                    createdAt: bookmark.createdAt,
                    updatedAt: bookmark.updatedAt
                )
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func assignFolder(_ folderID: String?, to bookmark: Bookmark, session: SessionStore) async {
        guard let token = session.accessToken else { return }

        do {
            let updated = try await api.assignBookmarkFolder(
                token: token,
                bookmarkID: bookmark.id,
                folderID: folderID
            )
            if let index = bookmarks.firstIndex(where: { $0.id == updated.id }) {
                bookmarks[index] = updated
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func folderName(for bookmark: Bookmark) -> String? {
        guard let folderID = bookmark.folderID else { return nil }
        return folders.first(where: { $0.id == folderID })?.name
    }

    func add(url: String, session: SessionStore) async -> Bool {
        guard let token = session.accessToken else { return false }

        do {
            let bookmark = try await api.createBookmark(token: token, url: url)
            bookmarks.insert(bookmark, at: 0)
            startPollingIfNeeded(session: session)
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func delete(_ bookmark: Bookmark, session: SessionStore) async {
        guard let token = session.accessToken else { return }

        do {
            try await api.deleteBookmark(token: token, id: bookmark.id)
            bookmarks.removeAll { $0.id == bookmark.id }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func startPollingIfNeeded(session: SessionStore) {
        guard bookmarks.contains(where: { EnrichmentState.isEnriching($0) }) else {
            stopPolling()
            return
        }
        guard pollingTask == nil else { return }

        pollingTask = Task {
            while !Task.isCancelled {
                try? await Task.sleep(for: .seconds(4))
                guard !Task.isCancelled else { break }
                await load(session: session)
                if !bookmarks.contains(where: { EnrichmentState.isEnriching($0) }) {
                    break
                }
            }
            pollingTask = nil
        }
    }

    func stopPolling() {
        pollingTask?.cancel()
        pollingTask = nil
    }
}
