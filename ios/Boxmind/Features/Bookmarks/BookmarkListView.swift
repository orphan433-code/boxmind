import SwiftUI

struct BookmarkListView: View {
    let bookmarks: [Bookmark]
    let session: SessionStore
    @Bindable var viewModel: BookmarksViewModel
    var activeTag: String?
    var onTagTap: ((String) -> Void)?
    var onClearTag: (() -> Void)?
    var emptyMessage: String = "Ничего не найдено"

    @State private var selected: Bookmark?

    var body: some View {
        if bookmarks.isEmpty {
            ContentUnavailableView(emptyMessage, systemImage: "magnifyingglass")
        } else {
            List {
                if let activeTag, let onClearTag {
                    ActiveTagFilterBar(tag: activeTag, onClear: onClearTag)
                        .listRowInsets(EdgeInsets())
                        .listRowSeparator(.hidden)
                        .listRowBackground(Color.clear)
                }

                ForEach(bookmarks) { bookmark in
                    Button {
                        selected = bookmark
                    } label: {
                        BookmarkRowView(
                            bookmark: bookmark,
                            activeTag: activeTag,
                            onTagTap: onTagTap
                        )
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
            .navigationDestination(item: $selected) { bookmark in
                BookmarkDetailView(bookmarkID: bookmark.id, viewModel: viewModel)
            }
        }
    }
}
