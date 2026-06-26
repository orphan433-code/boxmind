import SwiftUI

/// Square thumbnail for a bookmark. Falls back to a category-tinted gradient
/// with the category glyph when there is no remote image.
struct BookmarkThumbnail: View {
    let bookmark: Bookmark
    var size: CGFloat = 56
    var cornerRadius: CGFloat = 12

    var body: some View {
        Group {
            if let url = imageURL {
                AsyncImage(url: url) { phase in
                    switch phase {
                    case .success(let image):
                        image
                            .resizable()
                            .scaledToFill()
                    case .empty:
                        placeholder.overlay(ProgressView().controlSize(.small))
                    default:
                        placeholder
                    }
                }
            } else {
                placeholder
            }
        }
        .frame(width: size, height: size)
        .clipShape(RoundedRectangle(cornerRadius: cornerRadius, style: .continuous))
    }

    private var imageURL: URL? {
        guard !bookmark.imageURL.isEmpty else { return nil }
        return URL(string: bookmark.imageURL)
    }

    private var placeholder: some View {
        let tint = CategoryLabels.color(for: bookmark.category)
        return LinearGradient(
            colors: [tint.opacity(0.85), tint.opacity(0.5)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
        .overlay {
            Image(systemName: CategoryLabels.icon(for: bookmark.category))
                .font(.system(size: size * 0.38, weight: .semibold))
                .foregroundStyle(.white.opacity(0.95))
        }
    }
}

extension Bookmark {
    var hasImage: Bool {
        !imageURL.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            && URL(string: imageURL) != nil
    }

    /// Human-friendly host, e.g. "youtube.com" from a full URL.
    var displayHost: String {
        guard let host = URL(string: url)?.host() else { return url }
        return host.hasPrefix("www.") ? String(host.dropFirst(4)) : host
    }

    /// Host suitable for UI. Punycode domains (`xn--...`) are technically valid,
    /// but they look broken to users, so we hide them instead of showing noise.
    var readableHost: String? {
        let host = displayHost.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !host.isEmpty else { return nil }
        guard !host.lowercased().contains("xn--") else { return nil }
        return host
    }

    var displayTitle: String {
        let trimmed = title.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? (readableHost ?? "Ссылка") : trimmed
    }
}
