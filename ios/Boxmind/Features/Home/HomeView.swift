import SwiftUI

struct HomeView: View {
    @Environment(SessionStore.self) private var session
    @Bindable var viewModel: BookmarksViewModel

    @State private var showAddSheet = false
    @State private var searchText = ""

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.bookmarks.isEmpty {
                    ProgressView("Загружаем…")
                } else if viewModel.bookmarks.isEmpty {
                    ContentUnavailableView(
                        "Пока пусто",
                        systemImage: "tray",
                        description: Text("Добавь первую ссылку — AI разложит её по полочкам.")
                    )
                } else if filteredBookmarks.isEmpty {
                    ContentUnavailableView.search(text: searchText)
                } else {
                    BookmarkListView(
                        bookmarks: filteredBookmarks,
                        session: session,
                        viewModel: viewModel
                    )
                    .refreshable {
                        await viewModel.load(session: session)
                    }
                }
            }
            .navigationTitle("Главная")
            .searchable(text: $searchText, prompt: "Поиск по закладкам")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        showAddSheet = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .sheet(isPresented: $showAddSheet) {
                AddBookmarkSheet { url in
                    await viewModel.add(url: url, session: session)
                }
            }
            .alert("Ошибка", isPresented: Binding(
                get: { viewModel.errorMessage != nil },
                set: { if !$0 { viewModel.errorMessage = nil } }
            )) {
                Button("OK", role: .cancel) {}
            } message: {
                Text(viewModel.errorMessage ?? "")
            }
        }
    }

    private var filteredBookmarks: [Bookmark] {
        let query = searchText.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !query.isEmpty else { return viewModel.bookmarks }

        return viewModel.bookmarks.filter { bookmark in
            let searchableText = [
                bookmark.title,
                bookmark.description,
                bookmark.url,
                CategoryLabels.label(for: bookmark.category),
                bookmark.tags.joined(separator: " ")
            ].joined(separator: " ").lowercased()

            return searchableText.contains(query)
        }
    }
}
