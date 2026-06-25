import type { Bookmark } from "../types";

const ENRICH_WINDOW_MS = 3 * 60 * 1000;

// A bookmark is considered "enriching" while the AI background job has not yet
// filled in tags. We only treat freshly created bookmarks as enriching so a
// genuinely tag-less old bookmark doesn't spin forever.
export function isBookmarkEnriching(bookmark: Bookmark): boolean {
  if (bookmark.tags.length > 0) return false;

  const createdAt = new Date(bookmark.created_at).getTime();
  if (Number.isNaN(createdAt)) return false;

  return Date.now() - createdAt < ENRICH_WINDOW_MS;
}
