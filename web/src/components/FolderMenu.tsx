import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import type { Folder } from "../types";

type Props = {
  folders: Folder[];
  activeFolderId?: string;
  onAssign: (folderId: string | null) => void;
};

const MENU_WIDTH = 200;
const VIEWPORT_MARGIN = 8;

export function FolderMenu({ folders, activeFolderId, onAssign }: Props) {
  const [open, setOpen] = useState(false);
  const [coords, setCoords] = useState({ top: 0, left: 0 });
  const buttonRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  const isAssigned = Boolean(activeFolderId);

  const updatePosition = useCallback(() => {
    const button = buttonRef.current;
    if (!button) return;
    const rect = button.getBoundingClientRect();
    const left = Math.max(
      VIEWPORT_MARGIN,
      Math.min(rect.right - MENU_WIDTH, window.innerWidth - MENU_WIDTH - VIEWPORT_MARGIN),
    );
    setCoords({ top: rect.bottom + 6, left });
  }, []);

  useLayoutEffect(() => {
    if (open) updatePosition();
  }, [open, updatePosition]);

  useEffect(() => {
    if (!open) return;

    function handlePointer(event: MouseEvent) {
      const target = event.target as Node;
      if (buttonRef.current?.contains(target) || menuRef.current?.contains(target)) {
        return;
      }
      setOpen(false);
    }
    function handleKey(event: KeyboardEvent) {
      if (event.key === "Escape") setOpen(false);
    }
    function handleReflow() {
      setOpen(false);
    }

    document.addEventListener("mousedown", handlePointer);
    document.addEventListener("keydown", handleKey);
    window.addEventListener("resize", handleReflow);
    window.addEventListener("scroll", handleReflow, true);
    return () => {
      document.removeEventListener("mousedown", handlePointer);
      document.removeEventListener("keydown", handleKey);
      window.removeEventListener("resize", handleReflow);
      window.removeEventListener("scroll", handleReflow, true);
    };
  }, [open]);

  function choose(folderId: string | null) {
    onAssign(folderId);
    setOpen(false);
  }

  return (
    <>
      <button
        ref={buttonRef}
        type="button"
        className={isAssigned ? "icon-btn folder-btn assigned" : "icon-btn folder-btn"}
        onClick={() => setOpen((value) => !value)}
        aria-haspopup="menu"
        aria-expanded={open}
        title="Папка"
      >
        <FolderGlyph />
      </button>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            className="folder-menu-popover"
            role="menu"
            style={{ top: coords.top, left: coords.left, width: MENU_WIDTH }}
          >
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
          </div>,
          document.body,
        )}
    </>
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
      <path d="M3 8a2 2 0 0 1 2-2h3.6a2 2 0 0 1 1.4.6L11.8 8H19a2 2 0 0 1 2 2v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
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
