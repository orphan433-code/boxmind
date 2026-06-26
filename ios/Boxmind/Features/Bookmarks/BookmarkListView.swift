import SwiftUI

struct BookmarkListView: View {
    @Environment(\.openURL) private var openURL

    let bookmarks: [Bookmark]
    let session: SessionStore
    let viewModel: BookmarksViewModel
    var emptyMessage: String = "Ничего не найдено"

    var body: some View {
        if bookmarks.isEmpty {
            ContentUnavailableView(emptyMessage, systemImage: "magnifyingglass")
        } else {
            List {
                ForEach(bookmarks) { bookmark in
                    Button {
                        openBookmark(bookmark)
                    } label: {
                        BookmarkRowView(bookmark: bookmark)
                    }
                    .buttonStyle(.plain)
                    .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                        Button(role: .destructive) {
                            Task { await viewModel.delete(bookmark, session: session) }
                        } label: {
                            Label("Удалить", systemImage: "trash")
                        }
                    }
                }
            }
            .listStyle(.plain)
        }
    }

    private func openBookmark(_ bookmark: Bookmark) {
        guard let url = URL(string: bookmark.url) else {
            viewModel.errorMessage = "Не удалось открыть ссылку"
            return
        }
        openURL(url)
    }
}
