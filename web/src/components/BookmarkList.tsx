import type { Bookmark } from "../types";
import { BookmarkCard } from "./BookmarkCard";

type Props = {
  bookmarks: Bookmark[];
  onDelete: (id: string) => void;
  onTagClick?: (tag: string) => void;
  activeTag?: string | null;
  deletingId: string | null;
  showIntent?: boolean;
  emptyHint?: string;
};

export function BookmarkList({
  bookmarks,
  onDelete,
  onTagClick,
  activeTag,
  deletingId,
  showIntent = false,
  emptyHint = "Ничего не найдено",
}: Props) {
  if (bookmarks.length === 0) {
    return (
      <div className="empty-state-box">
        <p>{emptyHint}</p>
        <span>Кликни по тегу на карточке, чтобы отфильтровать</span>
      </div>
    );
  }

  return (
    <div className="bookmark-grid">
      {bookmarks.map((bookmark, index) => (
        <BookmarkCard
          key={bookmark.id}
          bookmark={bookmark}
          index={index}
          onDelete={onDelete}
          onTagClick={onTagClick}
          activeTag={activeTag}
          deleting={deletingId === bookmark.id}
          showIntent={showIntent}
        />
      ))}
    </div>
  );
}
