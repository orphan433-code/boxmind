import type { Bookmark } from "../types";
import { categoryLabel } from "../types";

export function normalizeSearchQuery(query: string): string {
  return query.trim().toLowerCase();
}

export function searchBookmarks(
  bookmarks: Bookmark[],
  rawQuery: string,
): Bookmark[] {
  const query = normalizeSearchQuery(rawQuery);
  if (!query) return bookmarks;

  const terms = query.split(/\s+/).filter(Boolean);

  return bookmarks.filter((bookmark) => {
    const haystack = [
      bookmark.title,
      bookmark.description,
      bookmark.url,
      categoryLabel(bookmark.category),
      ...bookmark.tags,
    ]
      .join(" ")
      .toLowerCase();

    return terms.every((term) => haystack.includes(term));
  });
}
