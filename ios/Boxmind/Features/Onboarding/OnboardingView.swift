import SwiftUI

struct OnboardingView: View {
    var onComplete: () -> Void

    @State private var page = 0
    private let pageCount = 3

    var body: some View {
        VStack(spacing: 0) {
            TabView(selection: $page) {
                welcomePage.tag(0)
                sharePage.tag(1)
                favoritesPage.tag(2)
            }
            .tabViewStyle(.page(indexDisplayMode: .always))
            .animation(.easeInOut(duration: 0.25), value: page)

            footer
                .padding(.horizontal, 24)
                .padding(.top, 8)
                .padding(.bottom, 20)
        }
        .background(Color(.systemGroupedBackground))
    }

    private var welcomePage: some View {
        ScrollView {
            VStack(spacing: 24) {
                Spacer(minLength: 24)

                Image("BrandIcon")
                    .resizable()
                    .frame(width: 88, height: 88)
                    .clipShape(RoundedRectangle(cornerRadius: 22, style: .continuous))
                    .shadow(color: .black.opacity(0.12), radius: 12, y: 6)

                VStack(spacing: 10) {
                    Text("Добро пожаловать")
                        .font(.title.bold())
                    Text(Brand.tagline)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }

                Text("Сохраняйте ссылки из Safari, Telegram, YouTube и других приложений — Boxmind сам разложит их по категориям и тегам.")
                    .font(.body)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: 340)

                Spacer(minLength: 24)
            }
            .frame(maxWidth: .infinity)
            .padding(.horizontal, 28)
        }
    }

    private var sharePage: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                pageHeader(
                    title: "Сохраняйте через «Поделиться»",
                    subtitle: "Самый быстрый способ — отправить ссылку прямо из приложения, где вы её нашли."
                )

                instructionStep(number: 1, text: "Откройте страницу и нажмите «Поделиться».")
                screenshot("OnboardingShare", height: 320)

                instructionStep(number: 2, text: "Выберите Boxmind. Если его нет в первом ряду — нажмите «Ещё».")
                screenshot("OnboardingShareMore", height: 170)
            }
            .padding(.horizontal, 24)
            .padding(.vertical, 8)
        }
    }

    private var favoritesPage: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                pageHeader(
                    title: "Добавьте в избранное",
                    subtitle: "Тогда иконка Boxmind всегда будет рядом — не придётся искать в «Ещё»."
                )

                instructionStep(number: 1, text: "В «Ещё» нажмите «Править».")
                screenshot("OnboardingShareEdit", height: 360)

                instructionStep(number: 2, text: "Нажмите «+» у Boxmind, чтобы добавить в избранное.")
                screenshot("OnboardingShareFavorite", height: 200)

                Text("После этого ссылки будут сохраняться в один тап — а Boxmind разложит их по категориям.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
            .padding(.horizontal, 24)
            .padding(.vertical, 8)
        }
    }

    private var footer: some View {
        VStack(spacing: 12) {
            Button {
                if page < pageCount - 1 {
                    page += 1
                } else {
                    onComplete()
                }
            } label: {
                Text(page < pageCount - 1 ? "Далее" : "Начать")
                    .font(.headline)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 14)
            }
            .buttonStyle(.borderedProminent)

            if page < pageCount - 1 {
                Button("Пропустить") {
                    onComplete()
                }
                .font(.subheadline)
                .foregroundStyle(.secondary)
            }
        }
    }

    private func pageHeader(title: String, subtitle: String) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title)
                .font(.title2.bold())
            Text(subtitle)
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.top, 4)
    }

    private func instructionStep(number: Int, text: String) -> some View {
        HStack(alignment: .top, spacing: 12) {
            Text("\(number)")
                .font(.caption.bold())
                .foregroundStyle(.white)
                .frame(width: 22, height: 22)
                .background(Circle().fill(Color.accentColor))

            Text(text)
                .font(.subheadline)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    private func screenshot(_ name: String, height: CGFloat) -> some View {
        Image(name)
            .resizable()
            .scaledToFit()
            .frame(maxWidth: .infinity)
            .frame(height: height)
            .padding(12)
            .background(Color(.secondarySystemGroupedBackground))
            .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
            .overlay {
                RoundedRectangle(cornerRadius: 18, style: .continuous)
                    .strokeBorder(Color.primary.opacity(0.06), lineWidth: 1)
            }
            .shadow(color: .black.opacity(0.06), radius: 12, y: 6)
            .accessibilityHidden(true)
    }
}

#Preview {
    OnboardingView(onComplete: {})
}
