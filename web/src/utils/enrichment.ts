import type { Bookmark } from "../types";

const ENRICH_WINDOW_MS = 5 * 60 * 1000;

// A bookmark is "enriching" while the background AI job has not finished yet.
// The backend flips `enriched` to true once enrichment completes (or gives up).
// The freshness window is a safety net so a stuck flag can't spin forever.
export function isBookmarkEnriching(bookmark: Bookmark): boolean {
  if (bookmark.enriched) return false;

  const createdAt = new Date(bookmark.created_at).getTime();
  if (Number.isNaN(createdAt)) return false;

  return Date.now() - createdAt < ENRICH_WINDOW_MS;
}
