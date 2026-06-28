import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import type { Bookmark, Folder } from "../types";

type Props = {
  bookmark: Bookmark;
  folders: Folder[];
  canAssignFolder: boolean;
  deleting: boolean;
  onAssignFolder?: (bookmarkId: string, folderId: string | null) => void;
  onDelete: (id: string) => void;
};

export function BookmarkActions({
  bookmark,
  folders,
  canAssignFolder,
  deleting,
  onAssignFolder,
  onDelete,
}: Props) {
  const [open, setOpen] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const sheetRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    function handleKey(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, [open]);

  useEffect(() => {
    if (open) {
      setConfirmDelete(false);
      sheetRef.current?.focus();
    }
  }, [open]);

  function close() {
    setOpen(false);
  }

  function choose(folderId: string | null) {
    onAssignFolder?.(bookmark.id, folderId);
    close();
  }

  function handleDelete() {
    onDelete(bookmark.id);
    close();
  }

  return (
    <>
      <button
        type="button"
        className="icon-btn actions-btn"
        onClick={() => setOpen(true)}
        aria-haspopup="dialog"
        aria-expanded={open}
        title="Действия"
        aria-label="Действия с закладкой"
      >
        <DotsGlyph />
      </button>

      {open &&
        createPortal(
          <div className="sheet-backdrop" role="presentation" onClick={close}>
            <div
              ref={sheetRef}
              className="sheet-card"
              role="dialog"
              aria-modal="true"
              aria-label="Действия с закладкой"
              tabIndex={-1}
              onClick={(event) => event.stopPropagation()}
            >
              <div className="sheet-handle" aria-hidden />

              <header className="sheet-header">
                <h3>{bookmark.title || bookmark.url}</h3>
              </header>

              {canAssignFolder && (
                <section className="sheet-section">
                  <p className="sheet-section-label">Папка</p>
                  <div className="sheet-list">
                    <button
                      type="button"
                      className="sheet-item"
                      onClick={() => choose(null)}
                    >
                      <span>Без папки</span>
                      {!bookmark.folder_id && <CheckGlyph />}
                    </button>
                    {folders.map((folder) => (
                      <button
                        key={folder.id}
                        type="button"
                        className="sheet-item"
                        onClick={() => choose(folder.id)}
                      >
                        <span>{folder.name}</span>
                        {bookmark.folder_id === folder.id && <CheckGlyph />}
                      </button>
                    ))}
                  </div>
                </section>
              )}

              <section className="sheet-section">
                <p className="sheet-section-label">Действия</p>
                <div className="sheet-list">
                  <a
                    className="sheet-item"
                    href={bookmark.url}
                    target="_blank"
                    rel="noreferrer"
                    onClick={close}
                  >
                    <span>Открыть ссылку</span>
                    <ExternalGlyph />
                  </a>

                  {confirmDelete ? (
                    <button
                      type="button"
                      className="sheet-item danger"
                      onClick={handleDelete}
                      disabled={deleting}
                    >
                      <span>Точно удалить?</span>
                      <TrashGlyph />
                    </button>
                  ) : (
                    <button
                      type="button"
                      className="sheet-item danger"
                      onClick={() => setConfirmDelete(true)}
                      disabled={deleting}
                    >
                      <span>Удалить закладку</span>
                      <TrashGlyph />
                    </button>
                  )}
                </div>
              </section>

              <button type="button" className="sheet-cancel" onClick={close}>
                Отмена
              </button>
            </div>
          </div>,
          document.body,
        )}
    </>
  );
}

function DotsGlyph() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
      <circle cx="5" cy="12" r="1.8" />
      <circle cx="12" cy="12" r="1.8" />
      <circle cx="19" cy="12" r="1.8" />
    </svg>
  );
}

function CheckGlyph() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}

function ExternalGlyph() {
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
      <path d="M14 4h6v6M20 4l-9 9M19 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h5" />
    </svg>
  );
}

function TrashGlyph() {
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
      <path d="M3 6h18M8 6V4a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2m2 0v14a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V6" />
      <path d="M10 11v6M14 11v6" />
    </svg>
  );
}
