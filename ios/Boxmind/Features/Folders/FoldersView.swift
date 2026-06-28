import SwiftUI

struct FoldersView: View {
    @Environment(SessionStore.self) private var session
    @Bindable var viewModel: BookmarksViewModel

    @State private var showCreateAlert = false
    @State private var newFolderName = ""

    private let columns = [
        GridItem(.flexible(), spacing: 12),
        GridItem(.flexible(), spacing: 12)
    ]

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading && viewModel.folders.isEmpty && viewModel.bookmarks.isEmpty {
                    ProgressView("Загружаем…")
                } else if viewModel.folders.isEmpty {
                    ContentUnavailableView {
                        Label("Папок пока нет", systemImage: "folder")
                    } description: {
                        Text("Создай папку и складывай туда уже сохранённые ссылки.")
                    } actions: {
                        Button("Создать папку") {
                            showCreateAlert = true
                        }
                        .buttonStyle(.borderedProminent)
                    }
                } else {
                    ScrollView {
                        LazyVGrid(columns: columns, spacing: 12) {
                            ForEach(viewModel.folders) { folder in
                                NavigationLink {
                                    FolderBookmarksView(folder: folder, viewModel: viewModel)
                                } label: {
                                    FolderTile(
                                        title: folder.name,
                                        count: count(for: folder)
                                    )
                                }
                                .buttonStyle(.plain)
                                .contextMenu {
                                    Button("Удалить папку", role: .destructive) {
                                        Task {
                                            await viewModel.deleteFolder(folder, session: session)
                                        }
                                    }
                                }
                            }
                        }
                        .padding()
                    }
                }
            }
            .navigationTitle("Моё")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        showCreateAlert = true
                    } label: {
                        Image(systemName: "plus")
                    }
                    .accessibilityLabel("Создать папку")
                }
            }
            .alert("Новая папка", isPresented: $showCreateAlert) {
                TextField("Название", text: $newFolderName)
                Button("Создать") {
                    let name = newFolderName.trimmingCharacters(in: .whitespacesAndNewlines)
                    guard !name.isEmpty else { return }
                    Task {
                        await viewModel.createFolder(name: name, session: session)
                        newFolderName = ""
                    }
                }
                Button("Отмена", role: .cancel) {
                    newFolderName = ""
                }
            } message: {
                Text("Например: Работочка, Go, Клиенты")
            }
        }
    }

    private func count(for folder: BookmarkFolder) -> Int {
        viewModel.bookmarks.filter { $0.folderID == folder.id }.count
    }
}

private struct FolderTile: View {
    let title: String
    let count: Int

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Image(systemName: "folder.fill")
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
