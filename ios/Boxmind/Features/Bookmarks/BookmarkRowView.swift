import SwiftUI

struct BookmarkRowView: View {
    let bookmark: Bookmark
    var activeTag: String?
    var onTagTap: ((String) -> Void)?

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            BookmarkThumbnail(bookmark: bookmark)

            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 8) {
                    Text(displayTitle)
                        .font(.headline)
                        .lineLimit(2)

                    if EnrichmentState.isEnriching(bookmark) {
                        ProgressView()
                            .controlSize(.small)
                    }
                }

                if !bookmark.description.isEmpty {
                    Text(bookmark.description)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }

                HStack(spacing: 6) {
                    Text(CategoryLabels.group(for: bookmark.category))
                        .font(.caption.weight(.semibold))
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.accentColor.opacity(0.12))
                        .clipShape(Capsule())

                    ForEach(bookmark.tags.prefix(3), id: \.self) { tag in
                        TagChip(
                            tag: tag,
                            isActive: activeTag?.caseInsensitiveCompare(tag) == .orderedSame,
                            onTap: onTagTap.map { handler in { handler(tag) } }
                        )
                    }

                    if bookmark.tags.count > 3 {
                        Text("+\(bookmark.tags.count - 3)")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .padding(.vertical, 6)
    }

    private var displayTitle: String {
        bookmark.displayTitle
    }
}

#Preview {
    List {
        BookmarkRowView(bookmark: Bookmark(
            id: "1",
            userID: "u1",
            url: "https://example.com",
            title: "Пример закладки",
            description: "Короткое описание карточки.",
            imageURL: "",
            category: "learning",
            tags: ["курс", "golang"],
            folderID: nil,
            enriched: false,
            createdAt: .now,
            updatedAt: .now
        ))
    }
}
