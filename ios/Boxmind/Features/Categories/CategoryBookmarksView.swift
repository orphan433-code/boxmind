import SwiftUI

struct CategoryBookmarksView: View {
    @Environment(SessionStore.self) private var session

    let group: String
    @Bindable var viewModel: BookmarksViewModel

    @State private var activeTag: String?

    var body: some View {
        BookmarkListView(
            bookmarks: filteredBookmarks,
            session: session,
            viewModel: viewModel,
            activeTag: activeTag,
            onTagTap: toggleTagFilter,
            onClearTag: { activeTag = nil },
            emptyMessage: activeTag == nil
                ? "В этой категории пока пусто"
                : "Нет закладок с этим тегом"
        )
        .navigationTitle(group)
        .navigationBarTitleDisplayMode(.large)
        .refreshable {
            await viewModel.load(session: session)
        }
    }

    private var bookmarks: [Bookmark] {
        viewModel.bookmarks.filter { CategoryLabels.group(for: $0.category) == group }
    }

    private var filteredBookmarks: [Bookmark] {
        guard let activeTag, !activeTag.isEmpty else { return bookmarks }
        return bookmarks.filter { BookmarkFiltering.matchesTag($0, tag: activeTag) }
    }

    private func toggleTagFilter(_ tag: String) {
        if activeTag?.caseInsensitiveCompare(tag) == .orderedSame {
            activeTag = nil
        } else {
            activeTag = tag
        }
    }
}
