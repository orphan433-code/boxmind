import SwiftUI

struct LoginView: View {
    @Environment(SessionStore.self) private var session
    @State private var viewModel = LoginViewModel()

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                Spacer(minLength: 0)

                brandHeader
                    .padding(.bottom, 40)

                Group {
                    if viewModel.step == .email {
                        emailStep
                    } else {
                        codeStep
                    }
                }
                .animation(.default, value: viewModel.step)

                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.footnote)
                        .foregroundStyle(.red)
                        .multilineTextAlignment(.center)
                        .frame(maxWidth: .infinity)
                        .padding(.top, 16)
                }

                Spacer(minLength: 0)
                Spacer(minLength: 0)
            }
            .frame(maxWidth: 420)
            .frame(maxWidth: .infinity)
            .padding(.horizontal, 28)
            .navigationBarHidden(true)
        }
    }

    private var brandHeader: some View {
        VStack(spacing: 16) {
            Image("BrandIcon")
                .resizable()
                .frame(width: 88, height: 88)
                .clipShape(RoundedRectangle(cornerRadius: 22, style: .continuous))
                .shadow(color: .black.opacity(0.15), radius: 12, y: 6)

            VStack(spacing: 6) {
                Text("Boxmind")
                    .font(.largeTitle.bold())
                Text(Brand.tagline)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }
        }
    }

    private var emailStep: some View {
        VStack(spacing: 16) {
            field(title: "Email") {
                TextField("you@example.com", text: $viewModel.email)
                    .textInputAutocapitalization(.never)
                    .keyboardType(.emailAddress)
                    .autocorrectionDisabled()
                    .textContentType(.emailAddress)
            }

            primaryButton(viewModel.isLoading ? "Отправляем…" : "Получить код") {
                Task { await viewModel.sendCode(session: session) }
            }
            .disabled(viewModel.isLoading || viewModel.email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
        }
    }

    private var codeStep: some View {
        VStack(spacing: 16) {
            Text("Код отправлен на **\(viewModel.email)**")
                .font(.footnote)
                .foregroundStyle(.secondary)
                .frame(maxWidth: .infinity, alignment: .leading)

            field(title: "Код из письма") {
                TextField("123456", text: $viewModel.code)
                    .keyboardType(.numberPad)
                    .textContentType(.oneTimeCode)
            }

            primaryButton(viewModel.isLoading ? "Проверяем…" : "Войти") {
                Task { await viewModel.verifyCode(session: session) }
            }
            .disabled(viewModel.isLoading || viewModel.code.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)

            Button("Изменить email") {
                viewModel.backToEmail()
            }
            .font(.subheadline)
        }
    }

    private func field<Content: View>(title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.subheadline.weight(.medium))
                .foregroundStyle(.secondary)
            content()
                .padding(14)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(Color(.secondarySystemBackground))
                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
        }
    }

    private func primaryButton(_ title: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Text(title)
                .font(.headline)
                .frame(maxWidth: .infinity)
                .frame(height: 26)
        }
        .buttonStyle(.borderedProminent)
        .controlSize(.large)
    }
}

#Preview {
    LoginView()
        .environment(SessionStore())
}
