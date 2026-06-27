import SwiftUI

struct AccountView: View {
    @Environment(SessionStore.self) private var session
    @Bindable var viewModel: BookmarksViewModel
    @State private var showOnboarding = false

    var body: some View {
        NavigationStack {
            List {
                Section {
                    if let email = session.userEmail {
                        LabeledContent("Email", value: email)
                    }
                    LabeledContent("Закладок", value: "\(viewModel.bookmarks.count)")
                }

                Section {
                    Button("Как сохранять ссылки") {
                        showOnboarding = true
                    }
                }

                Section {
                    Button("Выйти", role: .destructive) {
                        viewModel.stopPolling()
                        session.logout()
                    }
                }

                Section {
                    LabeledContent("Версия", value: appVersion)
                } footer: {
                    Text(Brand.tagline)
                }
            }
            .navigationTitle("Аккаунт")
            .fullScreenCover(isPresented: $showOnboarding) {
                OnboardingView {
                    showOnboarding = false
                }
            }
        }
    }

    private var appVersion: String {
        let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "—"
        let build = Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "—"
        return "\(version) (\(build))"
    }
}
