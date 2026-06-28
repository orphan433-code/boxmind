import { useEffect, useRef, useState } from "react";
import type { Folder } from "../types";

type Props = {
  folders: Folder[];
  activeFolderId?: string;
  onAssign: (folderId: string | null) => void;
};

export function FolderMenu({ folders, activeFolderId, onAssign }: Props) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    function handlePointer(event: MouseEvent) {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    function handleKey(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }

    document.addEventListener("mousedown", handlePointer);
    document.addEventListener("keydown", handleKey);
    return () => {
      document.removeEventListener("mousedown", handlePointer);
      document.removeEventListener("keydown", handleKey);
    };
  }, [open]);

  function choose(folderId: string | null) {
    onAssign(folderId);
    setOpen(false);
  }

  const isAssigned = Boolean(activeFolderId);

  return (
    <div className="folder-menu" ref={rootRef}>
      <button
        type="button"
        className={isAssigned ? "icon-btn folder-btn assigned" : "icon-btn folder-btn"}
        onClick={() => setOpen((value) => !value)}
        aria-haspopup="menu"
        aria-expanded={open}
        title="Папка"
      >
        <FolderGlyph />
      </button>

      {open && (
        <div className="folder-menu-popover" role="menu">
          <button
            type="button"
            role="menuitemradio"
            aria-checked={!isAssigned}
            className={!isAssigned ? "folder-menu-item active" : "folder-menu-item"}
            onClick={() => choose(null)}
          >
            <span>Без папки</span>
            {!isAssigned && <CheckGlyph />}
          </button>

          {folders.map((folder) => {
            const active = folder.id === activeFolderId;
            return (
              <button
                key={folder.id}
                type="button"
                role="menuitemradio"
                aria-checked={active}
                className={active ? "folder-menu-item active" : "folder-menu-item"}
                onClick={() => choose(folder.id)}
              >
                <span>{folder.name}</span>
                {active && <CheckGlyph />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

function FolderGlyph() {
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
      <path d="M4 20h16a1 1 0 0 0 1-1V8a1 1 0 0 0-1-1h-8L9.5 4.5A1.5 1.5 0 0 0 8.4 4H4a1 1 0 0 0-1 1v14a1 1 0 0 0 1 1z" />
    </svg>
  );
}

function CheckGlyph() {
  return (
    <svg
      width="15"
      height="15"
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
