import { ActiveTagFilter } from "./ActiveTagFilter";
import { BookmarkList } from "./BookmarkList";
import type { BrowseSectionId } from "../utils/browseSections";
import {
  sectionEmptyHint,
  sectionPanelSubtitle,
  sectionPanelTitle,
} from "../utils/browseSections";

type Props = {
  activeSection: BrowseSectionId;
  activeTag: string | null;
  onClearTag: () => void;
  panelTotal: number;
  bookmarks: Parameters<typeof BookmarkList>[0]["bookmarks"];
  onDelete: (id: string) => void;
  onTagClick: (tag: string) => void;
  deletingId: string | null;
  showIntent: boolean;
  isSearching: boolean;
  searchQuery: string;
};

export function BrowsePanel({
  activeSection,
  activeTag,
  onClearTag,
  panelTotal,
  bookmarks,
  onDelete,
  onTagClick,
  deletingId,
  showIntent,
  isSearching,
  searchQuery,
}: Props) {
  const subtitle = isSearching
    ? `по запросу «${searchQuery.trim()}»`
    : sectionPanelSubtitle(activeSection);

  const title = isSearching
    ? "Результаты поиска"
    : sectionPanelTitle(activeSection);

  return (
    <section className="browse-panel">
      <header className="browse-panel-header">
        <h2>{title}</h2>
        <p>
          {panelTotal} {pluralBookmarks(panelTotal)}
          {subtitle ? ` · ${subtitle}` : ""}
        </p>
      </header>

      {activeTag && <ActiveTagFilter tag={activeTag} onClear={onClearTag} />}

      <BookmarkList
        bookmarks={bookmarks}
        onDelete={onDelete}
        onTagClick={onTagClick}
        activeTag={activeTag}
        deletingId={deletingId}
        showIntent={showIntent || isSearching}
        emptyHint={
          isSearching
            ? "Ничего не нашлось — попробуй другие слова"
            : sectionEmptyHint(activeSection)
        }
      />
    </section>
  );
}

function pluralBookmarks(count: number): string {
  const mod10 = count % 10;
  const mod100 = count % 100;
  if (mod10 === 1 && mod100 !== 11) return "закладка";
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20))
    return "закладки";
  return "закладок";
}
