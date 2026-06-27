import SwiftUI

struct TagChip: View {
    let tag: String
    var isActive: Bool = false
    var onTap: (() -> Void)?

    var body: some View {
        Group {
            if let onTap {
                Button(action: onTap) {
                    label
                }
                .buttonStyle(.borderless)
            } else {
                label
            }
        }
    }

    private var label: some View {
        Text(tag)
            .font(.caption)
            .padding(.horizontal, 8)
            .padding(.vertical, 4)
            .background(isActive ? Color.accentColor.opacity(0.2) : Color(.secondarySystemBackground))
            .foregroundStyle(isActive ? Color.accentColor : .primary)
            .clipShape(Capsule())
    }
}

struct ActiveTagFilterBar: View {
    let tag: String
    let onClear: () -> Void

    var body: some View {
        HStack(spacing: 8) {
            Text("Фильтр")
                .font(.caption)
                .foregroundStyle(.secondary)
            TagChip(tag: tag, isActive: true)
            Spacer()
            Button(action: onClear) {
                Label("Сбросить", systemImage: "xmark.circle.fill")
                    .labelStyle(.titleAndIcon)
                    .font(.caption)
            }
            .buttonStyle(.borderless)
            .foregroundStyle(.secondary)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 8)
    }
}

enum BookmarkFiltering {
    static func matchesTag(_ bookmark: Bookmark, tag: String) -> Bool {
        bookmark.tags.contains { $0.caseInsensitiveCompare(tag) == .orderedSame }
    }

    static func filter(_ bookmarks: [Bookmark], searchText: String, activeTag: String?) -> [Bookmark] {
        var result = bookmarks

        if let activeTag, !activeTag.isEmpty {
            result = result.filter { matchesTag($0, tag: activeTag) }
        }

        let query = searchText.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard !query.isEmpty else { return result }

        return result.filter { bookmark in
            let searchable = [
                bookmark.title,
                bookmark.description,
                bookmark.url,
                CategoryLabels.group(for: bookmark.category),
                bookmark.tags.joined(separator: " ")
            ].joined(separator: " ").lowercased()

            return searchable.contains(query)
        }
    }
}
