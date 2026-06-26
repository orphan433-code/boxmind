import SwiftUI

struct AddBookmarkSheet: View {
    @Environment(\.dismiss) private var dismiss
    @State private var url = ""
    @State private var isSaving = false
    @State private var errorMessage: String?

    let onSave: (String) async -> Bool

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("https://…", text: $url)
                        .textInputAutocapitalization(.never)
                        .keyboardType(.URL)
                        .autocorrectionDisabled()
                } header: {
                    Text("Ссылка")
                } footer: {
                    Text("Вставь URL — AI обработает карточку в фоне.")
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.footnote)
                    }
                }
            }
            .navigationTitle("Новая ссылка")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Отмена") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(isSaving ? "Сохраняем…" : "Сохранить") {
                        Task { await save() }
                    }
                    .disabled(isSaving || url.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                }
            }
        }
    }

    private func save() async {
        isSaving = true
        defer { isSaving = false }

        let trimmed = url.trimmingCharacters(in: .whitespacesAndNewlines)
        let ok = await onSave(trimmed)
        if ok {
            dismiss()
        } else {
            errorMessage = "Не удалось сохранить ссылку"
        }
    }
}

#Preview {
    AddBookmarkSheet { _ in true }
}
