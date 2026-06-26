import SwiftUI

struct CategoriesView: View {
    @Bindable var viewModel: BookmarksViewModel

    private let columns = [
        GridItem(.flexible(), spacing: 12),
        GridItem(.flexible(), spacing: 12)
    ]

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.bookmarks.isEmpty {
                    ProgressView("Загружаем…")
                } else if groups.isEmpty {
                    ContentUnavailableView(
                        "Категорий пока нет",
                        systemImage: "square.grid.2x2",
                        description: Text("Сохрани первую ссылку — AI сам разложит её по категориям.")
                    )
                } else {
                    ScrollView {
                        LazyVGrid(columns: columns, spacing: 12) {
                            ForEach(groups, id: \.self) { group in
                                NavigationLink {
                                    CategoryBookmarksView(
                                        group: group,
                                        viewModel: viewModel
                                    )
                                } label: {
                                    CategoryTile(
                                        title: group,
                                        icon: CategoryLabels.iconForGroup(group),
                                        count: count(for: group)
                                    )
                                }
                                .buttonStyle(.plain)
                            }
                        }
                        .padding()
                    }
                }
            }
            .navigationTitle("Категории")
        }
    }

    private var groups: [String] {
        CategoryLabels.sortedGroups(Set(viewModel.bookmarks.map(\.category)))
    }

    private func count(for group: String) -> Int {
        viewModel.bookmarks.filter { CategoryLabels.group(for: $0.category) == group }.count
    }
}

private struct CategoryTile: View {
    let title: String
    let icon: String
    let count: Int

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundStyle(Color.accentColor)

            Spacer(minLength: 0)

            Text(title)
                .font(.headline)
                .foregroundStyle(.primary)
                .lineLimit(2)
                .multilineTextAlignment(.leading)

            Text("\(count) \(countLabel)")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, minHeight: 120, alignment: .leading)
        .padding(16)
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
    }

    private var countLabel: String {
        let mod10 = count % 10
        let mod100 = count % 100
        if mod100 >= 11 && mod100 <= 14 { return "ссылок" }
        switch mod10 {
        case 1: return "ссылка"
        case 2, 3, 4: return "ссылки"
        default: return "ссылок"
        }
    }
}
