import SwiftUI

struct HomeView: View {
    @Environment(SessionStore.self) private var session
    @Bindable var viewModel: BookmarksViewModel

    @State private var showAddSheet = false
    @State private var searchText = ""
    @State private var activeTag: String?

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.bookmarks.isEmpty {
                    ProgressView("Загружаем…")
                } else if viewModel.bookmarks.isEmpty {
                    ContentUnavailableView(
                        "Пока пусто",
                        systemImage: "tray",
                        description: Text("Добавь первую ссылку.")
                    )
                } else if filteredBookmarks.isEmpty {
                    ContentUnavailableView(
                        activeTag == nil ? "Ничего не найдено" : "Нет закладок с этим тегом",
                        systemImage: "magnifyingglass",
                        description: Text(
                            activeTag == nil
                                ? "Попробуй другой запрос."
                                : "Сбрось фильтр или выбери другой тег."
                        )
                    )
                } else {
                    BookmarkListView(
                        bookmarks: filteredBookmarks,
                        session: session,
                        viewModel: viewModel,
                        activeTag: activeTag,
                        onTagTap: toggleTagFilter,
                        onClearTag: { activeTag = nil }
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
        BookmarkFiltering.filter(
            viewModel.bookmarks,
            searchText: searchText,
            activeTag: activeTag
        )
    }

    private func toggleTagFilter(_ tag: String) {
        if activeTag?.caseInsensitiveCompare(tag) == .orderedSame {
            activeTag = nil
        } else {
            activeTag = tag
        }
    }
}
