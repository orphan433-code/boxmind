import SwiftUI

struct CategoryBookmarksView: View {
    @Environment(SessionStore.self) private var session

    let group: String
    @Bindable var viewModel: BookmarksViewModel

    var body: some View {
        BookmarkListView(
            bookmarks: bookmarks,
            session: session,
            viewModel: viewModel,
            emptyMessage: "В этой категории пока пусто"
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
}
