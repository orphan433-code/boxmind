import { useCallback, useEffect, useMemo, useState } from "react";
import {
  createBookmark,
  deleteBookmark,
  getBookmark,
  listBookmarks,
} from "../api/client";
import { useAuth } from "../auth/AuthContext";
import { AddBookmarkForm } from "../components/AddBookmarkForm";
import { BrowsePanel } from "../components/BrowsePanel";
import { PendingQueue } from "../components/PendingQueue";
import { SearchBar } from "../components/SearchBar";
import { SidebarNav } from "../components/SidebarNav";
import type { PendingBookmark } from "../types";
import {
  bookmarksForSection,
  countBySection,
  listActiveSections,
  showIntentOnCards,
  type BrowseSectionId,
} from "../utils/browseSections";
import { normalizeSearchQuery, searchBookmarks } from "../utils/search";
import { bookmarkUrlsMatch } from "../utils/url";

export function DashboardPage() {
  const { token, user, logout } = useAuth();
  const [bookmarks, setBookmarks] = useState<
    Awaited<ReturnType<typeof listBookmarks>>
  >([]);
  const [pending, setPending] = useState<PendingBookmark[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeSection, setActiveSection] = useState<BrowseSectionId>("recent");
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
      const data = await listBookmarks(token);
      setBookmarks(data);
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
    () =>
      bookmarks.some((bookmark) => {
        const createdAt = new Date(bookmark.created_at).getTime();
        const isFresh = Date.now() - createdAt < 3 * 60 * 1000;
        const looksPending =
          bookmark.category === "other" &&
          (!bookmark.description || bookmark.title === bookmark.url);
        return isFresh && looksPending;
      }),
    [bookmarks],
  );

  useEffect(() => {
    if (!token || !needsEnrichmentRefresh) return;

    const intervalId = window.setInterval(() => {
      void listBookmarks(token)
        .then((data) => setBookmarks(data))
        .catch(() => {});
    }, 5000);

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
  }, [activeSection, searchQuery]);

  useEffect(() => {
    if (!sidebarOpen) return;
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") setSidebarOpen(false);
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [sidebarOpen]);

  const inPanel = useMemo(() => {
    if (isSearching) {
      return searchBookmarks(bookmarks, searchQuery);
    }
    return bookmarksForSection(bookmarks, activeSection);
  }, [bookmarks, activeSection, isSearching, searchQuery]);

  const visibleBookmarks = useMemo(() => {
    if (!activeTag) return inPanel;
    return inPanel.filter((bookmark) => bookmark.tags.includes(activeTag));
  }, [inPanel, activeTag]);

  function handleSectionChange(sectionId: BrowseSectionId) {
    setSearchQuery("");
    setActiveSection(sectionId);
    setSidebarOpen(false);
  }

  function handleTagClick(tag: string) {
    setActiveTag((current) => (current === tag ? null : tag));
  }

  function handleAdd(url: string) {
    if (!token) return;

    if (bookmarks.some((bookmark) => bookmarkUrlsMatch(bookmark.url, url))) {
      setError("эта ссылка уже в ящике");
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
          ? "эта ссылка уже в ящике"
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

        {bookmarks.length > 0 && (
          <SidebarNav
            sections={sections}
            active={isSearching ? null : activeSection}
            counts={sectionCounts}
            onChange={handleSectionChange}
          />
        )}

        <div className="sidebar-footer">
          <button type="button" className="ghost-btn sidebar-logout" onClick={logout}>
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
            activeSection={activeSection}
            activeTag={activeTag}
            onClearTag={() => setActiveTag(null)}
            panelTotal={visibleBookmarks.length}
            bookmarks={visibleBookmarks}
            onDelete={handleDelete}
            onTagClick={handleTagClick}
            deletingId={deletingId}
            showIntent={showIntentOnCards(activeSection)}
            isSearching={isSearching}
            searchQuery={searchQuery}
          />
        )}
      </div>
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
