import SwiftUI

struct BookmarkRowView: View {
    let bookmark: Bookmark

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            thumbnail

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
                    Text(CategoryLabels.label(for: bookmark.category))
                        .font(.caption.weight(.semibold))
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.accentColor.opacity(0.12))
                        .clipShape(Capsule())

                    ForEach(bookmark.tags.prefix(2), id: \.self) { tag in
                        Text(tag)
                            .font(.caption)
                            .padding(.horizontal, 8)
                            .padding(.vertical, 4)
                            .background(Color(.secondarySystemBackground))
                            .clipShape(Capsule())
                    }
                }
            }
        }
        .padding(.vertical, 6)
    }

    @ViewBuilder
    private var thumbnail: some View {
        if let url = URL(string: bookmark.imageURL), !bookmark.imageURL.isEmpty {
            AsyncImage(url: url) { phase in
                switch phase {
                case .success(let image):
                    image
                        .resizable()
                        .scaledToFill()
                default:
                    placeholder
                }
            }
            .frame(width: 56, height: 56)
            .clipShape(RoundedRectangle(cornerRadius: 10))
        } else {
            placeholder
        }
    }

    private var placeholder: some View {
        RoundedRectangle(cornerRadius: 10)
            .fill(Color(.secondarySystemBackground))
            .frame(width: 56, height: 56)
            .overlay {
                Image(systemName: "link")
                    .foregroundStyle(.secondary)
            }
    }

    private var displayTitle: String {
        let title = bookmark.title.trimmingCharacters(in: .whitespacesAndNewlines)
        return title.isEmpty ? bookmark.url : title
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
            enriched: false,
            createdAt: .now,
            updatedAt: .now
        ))
    }
}
