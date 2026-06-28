import SwiftUI

struct MainTabView: View {
    @Environment(SessionStore.self) private var session
    @Environment(\.scenePhase) private var scenePhase
    @State private var viewModel = BookmarksViewModel()

    var body: some View {
        TabView {
            HomeView(viewModel: viewModel)
                .tabItem {
                    Label("Главная", systemImage: "house.fill")
                }

            CategoriesView(viewModel: viewModel)
                .tabItem {
                    Label("Категории", systemImage: "square.grid.2x2.fill")
                }

            FoldersView(viewModel: viewModel)
                .tabItem {
                    Label("Моё", systemImage: "folder.fill")
                }

            AccountView(viewModel: viewModel)
                .tabItem {
                    Label("Аккаунт", systemImage: "person.fill")
                }
        }
        .task {
            await viewModel.load(session: session)
            viewModel.startPollingIfNeeded(session: session)
        }
        .onChange(of: scenePhase) {
            if scenePhase == .active {
                Task { await viewModel.load(session: session) }
            }
        }
        .onDisappear {
            viewModel.stopPolling()
        }
    }
}
