import type { Folder } from "../types";

type Props = {
  folders: Folder[];
  activeFolderId: string | null;
  counts: Record<string, number>;
  onSelect: (folderId: string) => void;
  onCreate: () => void;
};

export function FolderNav({
  folders,
  activeFolderId,
  counts,
  onSelect,
  onCreate,
}: Props) {
  return (
    <section className="sidebar-folders" aria-label="Мои папки">
      <div className="sidebar-folders-header">
        <span>Мои папки</span>
        <button
          type="button"
          className="sidebar-folders-add"
          onClick={onCreate}
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
      <path d="M3 8a2 2 0 0 1 2-2h3.6a2 2 0 0 1 1.4.6L11.8 8H19a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
    </svg>
  );
}
