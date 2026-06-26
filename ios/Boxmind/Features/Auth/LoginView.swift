import SwiftUI

struct LoginView: View {
    @Environment(SessionStore.self) private var session
    @State private var viewModel = LoginViewModel()

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: 24) {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Boxmind")
                        .font(.largeTitle.bold())
                    Text("Закинь ссылку в ящик — AI сам разложит.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                if viewModel.step == .email {
                    emailStep
                } else {
                    codeStep
                }

                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.footnote)
                        .foregroundStyle(.red)
                }

                Spacer()
            }
            .padding(24)
            .navigationBarHidden(true)
        }
    }

    private var emailStep: some View {
        VStack(alignment: .leading, spacing: 16) {
            VStack(alignment: .leading, spacing: 8) {
                Text("Email")
                    .font(.subheadline.weight(.medium))
                TextField("you@example.com", text: $viewModel.email)
                    .textInputAutocapitalization(.never)
                    .keyboardType(.emailAddress)
                    .autocorrectionDisabled()
                    .textContentType(.emailAddress)
                    .padding(12)
                    .background(Color(.secondarySystemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            Button {
                Task { await viewModel.sendCode(session: session) }
            } label: {
                Text(viewModel.isLoading ? "Отправляем…" : "Получить код")
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.borderedProminent)
            .disabled(viewModel.isLoading || viewModel.email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
        }
    }

    private var codeStep: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Код отправлен на **\(viewModel.email)**")
                .font(.footnote)
                .foregroundStyle(.secondary)

            VStack(alignment: .leading, spacing: 8) {
                Text("Код из письма")
                    .font(.subheadline.weight(.medium))
                TextField("123456", text: $viewModel.code)
                    .keyboardType(.numberPad)
                    .textContentType(.oneTimeCode)
                    .padding(12)
                    .background(Color(.secondarySystemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            Button {
                Task { await viewModel.verifyCode(session: session) }
            } label: {
                Text(viewModel.isLoading ? "Проверяем…" : "Войти")
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.borderedProminent)
            .disabled(viewModel.isLoading || viewModel.code.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)

            Button("Изменить email") {
                viewModel.backToEmail()
            }
            .font(.footnote)
        }
    }
}

#Preview {
    LoginView()
        .environment(SessionStore())
}
