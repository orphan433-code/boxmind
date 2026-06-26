import SwiftUI

struct AccountView: View {
    @Environment(SessionStore.self) private var session
    @Bindable var viewModel: BookmarksViewModel

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
                    Button("Выйти", role: .destructive) {
                        viewModel.stopPolling()
                        session.logout()
                    }
                }

                Section {
                    LabeledContent("Версия", value: appVersion)
                } footer: {
                    Text("Boxmind сохраняет ссылки и помогает AI разложить их по категориям.")
                }
            }
            .navigationTitle("Аккаунт")
        }
    }

    private var appVersion: String {
        let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "—"
        let build = Bundle.main.infoDictionary?["CFBundleVersion"] as? String ?? "—"
        return "\(version) (\(build))"
    }
}
