import type { Folder } from "../types";

type Props = {
  folders: Folder[];
  activeFolderId: string | null;
  counts: Record<string, number>;
  onSelect: (folderId: string) => void;
  onCreate: (name: string) => void;
  onDelete: (folderId: string) => void;
};

export function FolderNav({
  folders,
  activeFolderId,
  counts,
  onSelect,
  onCreate,
  onDelete,
}: Props) {
  function handleCreate() {
    const name = window.prompt("Название папки");
    if (!name?.trim()) return;
    onCreate(name.trim());
  }

  return (
    <section className="sidebar-folders" aria-label="Мои папки">
      <div className="sidebar-folders-header">
        <span>Мои папки</span>
        <button
          type="button"
          className="sidebar-folders-add"
          onClick={handleCreate}
          aria-label="Создать папку"
          title="Создать папку"
        >
          +
        </button>
      </div>

      {folders.length === 0 ? (
        <p className="sidebar-folders-empty">Создай папку для своих подборок</p>
      ) : (
        <ul className="sidebar-nav-list">
          {folders.map((folder) => (
            <li key={folder.id}>
              <button
                type="button"
                className={
                  activeFolderId === folder.id
                    ? "sidebar-nav-item active"
                    : "sidebar-nav-item"
                }
                onClick={() => onSelect(folder.id)}
                aria-current={activeFolderId === folder.id ? "page" : undefined}
              >
                <span className="sidebar-nav-icon">
                  <FolderIcon />
                </span>
                <span className="sidebar-nav-label">{folder.name}</span>
                <span className="sidebar-nav-count">{counts[folder.id] ?? 0}</span>
              </button>
              <button
                type="button"
                className="sidebar-folder-delete"
                onClick={() => onDelete(folder.id)}
                aria-label={`Удалить папку ${folder.name}`}
                title="Удалить папку"
              >
                ×
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function FolderIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M4 20h16a1 1 0 0 0 1-1V5a1 1 0 0 0-1-1H9l-2 3H4a1 1 0 0 0-1 1v11a1 1 0 0 0 1 1z" />
    </svg>
  );
}
