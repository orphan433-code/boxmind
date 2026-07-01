import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePageMeta } from "../hooks/usePageMeta";
import {
  assignBookmarkFolder,
  createBookmark,
  createFolder,
  deleteBookmark,
  deleteFolder,
  getBookmark,
  listBookmarks,
  listFolders,
} from "../api/client";
import { useAuth } from "../auth/AuthContext";
import { AddBookmarkForm } from "../components/AddBookmarkForm";
import { BrowsePanel } from "../components/BrowsePanel";
import { FolderNav } from "../components/FolderNav";
import { Modal } from "../components/Modal";
import { PendingQueue } from "../components/PendingQueue";
import { SearchBar } from "../components/SearchBar";
import { SidebarNav } from "../components/SidebarNav";
import type { Folder, PendingBookmark } from "../types";
import {
  bookmarksForSection,
  countBySection,
  listActiveSections,
  showIntentOnCards,
  type BrowseSectionId,
} from "../utils/browseSections";
import { isBookmarkEnriching } from "../utils/enrichment";
import { normalizeSearchQuery, searchBookmarks } from "../utils/search";
import { bookmarkUrlsMatch } from "../utils/url";

export function DashboardPage() {
  const { token, user, logout } = useAuth();
  const navigate = useNavigate();

  usePageMeta({
    title: "Приложение - Boxmind",
    description: "Личная библиотека ссылок в Boxmind.",
    path: "/app",
    noindex: true,
  });
  const [bookmarks, setBookmarks] = useState<
    Awaited<ReturnType<typeof listBookmarks>>
  >([]);
  const [pending, setPending] = useState<PendingBookmark[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeSection, setActiveSection] = useState<BrowseSectionId>("recent");
  const [activeFolderId, setActiveFolderId] = useState<string | null>(null);
  const [folders, setFolders] = useState<Folder[]>([]);
  const [createFolderOpen, setCreateFolderOpen] = useState(false);
  const [newFolderName, setNewFolderName] = useState("");
  const [creatingFolder, setCreatingFolder] = useState(false);
  const [folderToDelete, setFolderToDelete] = useState<Folder | null>(null);
  const [activeTag, setActiveTag] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const isSearching = normalizeSearchQuery(searchQuery).length > 0;

  const loadBookmarks = useCallback(async () => {
    if (!token) return;
    setError("");
    setLoading(true);
    try {
      const [data, folderData] = await Promise.all([
        listBookmarks(token),
        listFolders(token),
      ]);
      setBookmarks(data);
      setFolders(folderData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "не удалось загрузить");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    void loadBookmarks();
  }, [loadBookmarks]);

  const needsEnrichmentRefresh = useMemo(
    () => bookmarks.some(isBookmarkEnriching),
    [bookmarks],
  );

  useEffect(() => {
    if (!token || !needsEnrichmentRefresh) return;

    const intervalId = window.setInterval(() => {
      void listBookmarks(token)
        .then((data) => setBookmarks(data))
        .catch(() => {});
    }, 4000);

    return () => window.clearInterval(intervalId);
  }, [token, needsEnrichmentRefresh]);

  const sections = useMemo(() => listActiveSections(bookmarks), [bookmarks]);
  const sectionCounts = useMemo(() => countBySection(bookmarks), [bookmarks]);

  useEffect(() => {
    if (bookmarks.length === 0) return;
    if (isSearching) return;
    if (!sections.some((section) => section.id === activeSection)) {
      setActiveSection("recent");
    }
  }, [bookmarks.length, sections, activeSection, isSearching]);

  useEffect(() => {
    setActiveTag(null);
  }, [activeSection, activeFolderId, searchQuery]);

  useEffect(() => {
    if (activeFolderId && !folders.some((folder) => folder.id === activeFolderId)) {
      setActiveFolderId(null);
    }
  }, [folders, activeFolderId]);

  useEffect(() => {
    if (!sidebarOpen) return;
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setSidebarOpen(false);
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [sidebarOpen]);

  const folderCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const folder of folders) {
      counts[folder.id] = bookmarks.filter((bookmark) => bookmark.folder_id === folder.id).length;
    }
    return counts;
  }, [folders, bookmarks]);

  const activeFolder = useMemo(
    () => folders.find((folder) => folder.id === activeFolderId) ?? null,
    [folders, activeFolderId],
  );

  const inPanel = useMemo(() => {
    if (isSearching) {
      return searchBookmarks(bookmarks, searchQuery);
    }
    if (activeFolderId) {
      return bookmarks.filter((bookmark) => bookmark.folder_id === activeFolderId);
    }
    return bookmarksForSection(bookmarks, activeSection);
  }, [bookmarks, activeSection, activeFolderId, isSearching, searchQuery]);

  const visibleBookmarks = useMemo(() => {
    if (!activeTag) return inPanel;
    return inPanel.filter((bookmark) => bookmark.tags.includes(activeTag));
  }, [inPanel, activeTag]);

  function handleSectionChange(sectionId: BrowseSectionId) {
    setSearchQuery("");
    setActiveFolderId(null);
    setActiveSection(sectionId);
    setSidebarOpen(false);
  }

  function handleFolderChange(folderId: string) {
    setSearchQuery("");
    setActiveFolderId(folderId);
    setSidebarOpen(false);
  }

  function openCreateFolder() {
    setNewFolderName("");
    setCreateFolderOpen(true);
    setSidebarOpen(false);
  }

  async function handleCreateFolder() {
    if (!token) return;
    const name = newFolderName.trim();
    if (!name) return;

    setCreatingFolder(true);
    try {
      const folder = await createFolder(token, name);
      setFolders((prev) => [...prev, folder].sort((a, b) => a.name.localeCompare(b.name, "ru")));
      setActiveFolderId(folder.id);
      setSearchQuery("");
      setError("");
      setCreateFolderOpen(false);
      setNewFolderName("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "не удалось создать папку");
    } finally {
      setCreatingFolder(false);
    }
  }

  async function confirmDeleteFolder() {
    if (!token || !folderToDelete) return;
    const folderId = folderToDelete.id;

    try {
      await deleteFolder(token, folderId);
      setFolders((prev) => prev.filter((item) => item.id !== folderId));
      setBookmarks((prev) =>
        prev.map((bookmark) =>
          bookmark.folder_id === folderId ? { ...bookmark, folder_id: undefined } : bookmark,
        ),
      );
      if (activeFolderId === folderId) {
        setActiveFolderId(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "не удалось удалить папку");
    } finally {
      setFolderToDelete(null);
    }
  }

  async function handleAssignFolder(bookmarkId: string, folderId: string | null) {
    if (!token) return;
    try {
      const updated = await assignBookmarkFolder(token, bookmarkId, folderId);
      setBookmarks((prev) =>
        prev.map((bookmark) => (bookmark.id === updated.id ? updated : bookmark)),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "не удалось переместить ссылку");
    }
  }

  function handleTagClick(tag: string) {
    setActiveTag((current) => (current === tag ? null : tag));
  }

  function handleAdd(url: string) {
    if (!token) return;

    if (bookmarks.some((bookmark) => bookmarkUrlsMatch(bookmark.url, url))) {
      setError("Уже сохранена");
      return;
    }

    const pendingId = crypto.randomUUID();
    setPending((prev) => [...prev, { id: pendingId, url, status: "pending" }]);

    void (async () => {
      try {
        const created = await createBookmark(token, url);
        setBookmarks((prev) => [created, ...prev]);
        setSearchQuery("");
        setActiveSection("recent");
        setActiveTag(null);
        setPending((prev) => prev.filter((item) => item.id !== pendingId));

        if (!created.image_url) {
          window.setTimeout(async () => {
            try {
              const updated = await getBookmark(token, created.id);
              if (!updated.image_url) return;
              setBookmarks((prev) =>
                prev.map((item) => (item.id === updated.id ? updated : item)),
              );
            } catch {
              // ignore background image refresh errors
            }
          }, 5000);
        }
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "не удалось сохранить";
        const friendlyMessage = message.includes("already exists")
          ? "Уже сохранена"
          : message;
        setPending((prev) =>
          prev.map((item) =>
            item.id === pendingId
              ? { ...item, status: "error", error: friendlyMessage }
              : item,
          ),
        );
      }
    })();
  }

  async function handleDelete(id: string) {
    if (!token) return;
    setDeletingId(id);
    try {
      await deleteBookmark(token, id);
      setBookmarks((prev) => prev.filter((b) => b.id !== id));
    } finally {
      setDeletingId(null);
    }
  }

  function dismissPending(id: string) {
    setPending((prev) => prev.filter((item) => item.id !== id));
  }

  return (
    <div className={`dashboard-shell${sidebarOpen ? " sidebar-open" : ""}`}>
      <div
        className="sidebar-backdrop"
        onClick={() => setSidebarOpen(false)}
        aria-hidden={!sidebarOpen}
      />

      <aside className="dashboard-sidebar" aria-label="Навигация">
        <div className="sidebar-brand">
          <h1>Boxmind</h1>
          <p className="sidebar-email">{user?.email}</p>
        </div>

        <SearchBar
          value={searchQuery}
          onChange={setSearchQuery}
          className="sidebar-search"
        />

        <FolderNav
          folders={folders}
          activeFolderId={isSearching ? null : activeFolderId}
          counts={folderCounts}
          onSelect={handleFolderChange}
          onCreate={openCreateFolder}
        />

        {bookmarks.length > 0 && folders.length > 0 && (
          <div className="sidebar-nav-divider" />
        )}

        {bookmarks.length > 0 && (
          <SidebarNav
            sections={sections}
            active={isSearching || activeFolderId ? null : activeSection}
            counts={sectionCounts}
            onChange={handleSectionChange}
          />
        )}

        <div className="sidebar-footer">
          <button
            type="button"
            className="ghost-btn sidebar-logout"
            onClick={() => {
              logout();
              navigate("/login", { replace: true });
            }}
          >
            Выйти
          </button>
        </div>
      </aside>

      <div className="dashboard-main">
        <header className="main-topbar">
          <button
            type="button"
            className="sidebar-toggle"
            onClick={() => setSidebarOpen((open) => !open)}
            aria-label="Открыть меню"
            aria-expanded={sidebarOpen}
          >
            <MenuIcon />
          </button>
        </header>

        <AddBookmarkForm onSubmit={handleAdd} />
        <PendingQueue items={pending} onDismiss={dismissPending} />

        {loading && <p className="status" aria-live="polite">Загружаем закладки…</p>}
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}

        {!loading && bookmarks.length === 0 && pending.length === 0 && (
          <div className="empty-hero">
            <h2>Пока пусто</h2>
            <p>Вставь первую ссылку — она появится в «Недавних».</p>
          </div>
        )}

        {bookmarks.length > 0 && (
          <BrowsePanel
            key={isSearching ? "search" : activeFolderId ?? activeSection}
            activeSection={activeSection}
            activeTag={activeTag}
            onClearTag={() => setActiveTag(null)}
            panelTotal={visibleBookmarks.length}
            bookmarks={visibleBookmarks}
            onDelete={handleDelete}
            onTagClick={handleTagClick}
            deletingId={deletingId}
            showIntent={showIntentOnCards(activeSection) && !activeFolderId}
            isSearching={isSearching}
            searchQuery={searchQuery}
            titleOverride={activeFolder ? activeFolder.name : undefined}
            subtitleOverride={activeFolder ? "твоя подборка" : undefined}
            emptyHintOverride={
              activeFolder ? "В этой папке пока пусто — добавь ссылки с карточек" : undefined
            }
            headerAction={
              activeFolder && !isSearching ? (
                <button
                  type="button"
                  className="panel-action-btn danger"
                  onClick={() => setFolderToDelete(activeFolder)}
                >
                  <TrashIcon />
                  Удалить папку
                </button>
              ) : undefined
            }
            folders={folders}
            onAssignFolder={handleAssignFolder}
          />
        )}
      </div>

      <Modal
        open={createFolderOpen}
        title="Новая папка"
        onClose={() => setCreateFolderOpen(false)}
      >
        <form
          className="folder-form"
          onSubmit={(event) => {
            event.preventDefault();
            void handleCreateFolder();
          }}
        >
          <label className="folder-form-field">
            <span>Название</span>
            <input
              type="text"
              value={newFolderName}
              onChange={(event) => setNewFolderName(event.target.value)}
              placeholder="Например, Работа"
              maxLength={80}
              autoFocus
            />
          </label>
          <p className="folder-form-hint">
            Складывай сюда уже сохранённые ссылки — категории от AI не меняются.
          </p>
          <div className="modal-actions">
            <button
              type="button"
              className="ghost-btn"
              onClick={() => setCreateFolderOpen(false)}
            >
              Отмена
            </button>
            <button
              type="submit"
              className="primary-btn"
              disabled={!newFolderName.trim() || creatingFolder}
            >
              {creatingFolder ? "Создаём…" : "Создать"}
            </button>
          </div>
        </form>
      </Modal>

      <Modal
        open={Boolean(folderToDelete)}
        title="Удалить папку"
        onClose={() => setFolderToDelete(null)}
      >
        <p className="modal-text">
          Удалить папку «{folderToDelete?.name}»? Ссылки внутри останутся — пропадёт только
          сама подборка.
        </p>
        <div className="modal-actions">
          <button
            type="button"
            className="ghost-btn"
            onClick={() => setFolderToDelete(null)}
          >
            Отмена
          </button>
          <button
            type="button"
            className="danger-btn"
            onClick={() => void confirmDeleteFolder()}
          >
            Удалить
          </button>
        </div>
      </Modal>
    </div>
  );
}

function MenuIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      aria-hidden
    >
      <path d="M4 7h16M4 12h16M4 17h16" />
    </svg>
  );
}

function TrashIcon() {
  return (
    <svg
      width="15"
      height="15"
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
