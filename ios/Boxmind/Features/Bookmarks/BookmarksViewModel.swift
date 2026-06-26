import Foundation
import Observation

@Observable
@MainActor
final class BookmarksViewModel {
    var bookmarks: [Bookmark] = []
    var isLoading = false
    var errorMessage: String?

    private var pollingTask: Task<Void, Never>?
    private let api = APIClient()

    func load(session: SessionStore) async {
        guard let token = session.accessToken else { return }

        if bookmarks.isEmpty {
            isLoading = true
        }
        defer { isLoading = false }

        do {
            bookmarks = try await api.listBookmarks(token: token)
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
