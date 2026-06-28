import SwiftUI

struct BookmarkDetailView: View {
    @Environment(SessionStore.self) private var session
    @Environment(\.openURL) private var openURL
    @Environment(\.dismiss) private var dismiss

    let bookmarkID: String
    @Bindable var viewModel: BookmarksViewModel

    @State private var showDeleteConfirm = false

    private var bookmark: Bookmark? {
        viewModel.bookmarks.first { $0.id == bookmarkID }
    }

    var body: some View {
        Group {
            if let bookmark {
                detailContent(for: bookmark)
            } else {
                ContentUnavailableView("Закладка не найдена", systemImage: "link.badge.plus")
            }
        }
        .background(Color(.systemGroupedBackground))
        .navigationTitle("")
        .navigationBarTitleDisplayMode(.inline)
    }

    @ViewBuilder
    private func detailContent(for bookmark: Bookmark) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                if bookmark.hasImage {
                    hero(for: bookmark)
                }
                header(for: bookmark)
                if !bookmark.description.isEmpty {
                    Text(bookmark.description)
                        .font(.body)
                        .foregroundStyle(.secondary)
                }
                if !bookmark.tags.isEmpty {
                    tags(for: bookmark)
                }
                metadata(for: bookmark)
            }
            .padding(20)
        }
        .safeAreaInset(edge: .bottom) {
            actionBar(for: bookmark)
        }
        .confirmationDialog("Удалить закладку?", isPresented: $showDeleteConfirm, titleVisibility: .visible) {
            Button("Удалить", role: .destructive) {
                Task {
                    await viewModel.delete(bookmark, session: session)
                    dismiss()
                }
            }
            Button("Отмена", role: .cancel) {}
        }
    }

    private func hero(for bookmark: Bookmark) -> some View {
        BookmarkThumbnail(bookmark: bookmark, size: heroSize, cornerRadius: 20)
            .frame(maxWidth: .infinity)
            .shadow(color: .black.opacity(0.12), radius: 12, y: 6)
    }

    private var heroSize: CGFloat { 200 }

    private func header(for bookmark: Bookmark) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            if !bookmark.hasImage {
                BookmarkThumbnail(bookmark: bookmark, size: 64, cornerRadius: 16)
            }

            VStack(alignment: .leading, spacing: 10) {
                HStack(spacing: 8) {
                    categoryChip(for: bookmark)

                    if EnrichmentState.isEnriching(bookmark) {
                        HStack(spacing: 4) {
                            ProgressView().controlSize(.small)
                            Text("Дорабатывается")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                Text(bookmark.displayTitle)
                    .font(.title.bold())
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func categoryChip(for bookmark: Bookmark) -> some View {
        Label(
            CategoryLabels.label(for: bookmark.category),
            systemImage: CategoryLabels.icon(for: bookmark.category)
        )
        .font(.caption.weight(.semibold))
        .padding(.horizontal, 10)
        .padding(.vertical, 5)
        .background(CategoryLabels.color(for: bookmark.category).opacity(0.15))
        .foregroundStyle(CategoryLabels.color(for: bookmark.category))
        .clipShape(Capsule())
    }

    private func tags(for bookmark: Bookmark) -> some View {
        FlowLayout(spacing: 8) {
            ForEach(bookmark.tags, id: \.self) { tag in
                TagChip(tag: tag)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func metadata(for bookmark: Bookmark) -> some View {
        VStack(spacing: 0) {
            if !viewModel.folders.isEmpty {
                folderRow(for: bookmark)
                Divider()
            }
            metaRow(title: "Добавлено", value: bookmark.createdAt.formatted(date: .abbreviated, time: .shortened))
            Divider()
            metaRow(title: "Ссылка", value: bookmark.readableHost ?? "Открыть сайт")
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 4)
        .background(Color(.secondarySystemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
    }

    private func folderRow(for bookmark: Bookmark) -> some View {
        HStack {
            Text("Папка")
                .foregroundStyle(.secondary)
            Spacer()
            Picker("Папка", selection: folderSelection(for: bookmark)) {
                Text("Без папки").tag(Optional<String>.none)
                ForEach(viewModel.folders) { folder in
                    Text(folder.name).tag(Optional(folder.id))
                }
            }
            .labelsHidden()
            .pickerStyle(.menu)
        }
        .font(.subheadline)
        .padding(.vertical, 12)
    }

    private func folderSelection(for bookmark: Bookmark) -> Binding<String?> {
        Binding(
            get: { viewModel.bookmarks.first(where: { $0.id == bookmark.id })?.folderID },
            set: { newValue in
                Task {
                    await viewModel.assignFolder(newValue, to: bookmark, session: session)
                }
            }
        )
    }

    private func metaRow(title: String, value: String) -> some View {
        HStack {
            Text(title)
                .foregroundStyle(.secondary)
            Spacer()
            Text(value)
                .multilineTextAlignment(.trailing)
        }
        .font(.subheadline)
        .padding(.vertical, 12)
    }

    private func actionBar(for bookmark: Bookmark) -> some View {
        HStack(spacing: 12) {
            Button {
                if let url = URL(string: bookmark.url) { openURL(url) }
            } label: {
                Label("Открыть", systemImage: "safari")
                    .font(.headline)
                    .frame(maxWidth: .infinity)
                    .frame(height: 26)
            }
            .buttonStyle(.borderedProminent)
            .controlSize(.large)

            if let url = URL(string: bookmark.url) {
                ShareLink(item: url) {
                    Image(systemName: "square.and.arrow.up")
                        .font(.headline)
                        .frame(width: 30, height: 26)
                }
                .buttonStyle(.bordered)
                .controlSize(.large)
            }

            Button(role: .destructive) {
                showDeleteConfirm = true
            } label: {
                Image(systemName: "trash")
                    .font(.headline)
                    .frame(width: 30, height: 26)
            }
            .buttonStyle(.bordered)
            .controlSize(.large)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 12)
        .background(.bar)
    }
}

/// Simple wrapping layout for tag chips.
struct FlowLayout: Layout {
    var spacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout Void) -> CGSize {
        let maxWidth = proposal.width ?? .infinity
        var rows: [[LayoutSubviews.Element]] = [[]]
        var x: CGFloat = 0
        var totalHeight: CGFloat = 0
        var rowHeight: CGFloat = 0

        for view in subviews {
            let size = view.sizeThatFits(.unspecified)
            if x + size.width > maxWidth, !rows[rows.count - 1].isEmpty {
                totalHeight += rowHeight + spacing
                rowHeight = 0
                x = 0
                rows.append([])
            }
            rows[rows.count - 1].append(view)
            x += size.width + spacing
            rowHeight = max(rowHeight, size.height)
        }
        totalHeight += rowHeight
        return CGSize(width: maxWidth == .infinity ? x : maxWidth, height: totalHeight)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout Void) {
        var x = bounds.minX
        var y = bounds.minY
        var rowHeight: CGFloat = 0

        for view in subviews {
            let size = view.sizeThatFits(.unspecified)
            if x + size.width > bounds.maxX, x > bounds.minX {
                x = bounds.minX
                y += rowHeight + spacing
                rowHeight = 0
            }
            view.place(at: CGPoint(x: x, y: y), proposal: ProposedViewSize(size))
            x += size.width + spacing
            rowHeight = max(rowHeight, size.height)
        }
    }
}
