-- Remove duplicate bookmarks (same user + url), keeping the earliest one.
DELETE FROM bookmarks a
USING bookmarks b
WHERE a.user_id = b.user_id
  AND a.url = b.url
  AND (a.created_at > b.created_at
       OR (a.created_at = b.created_at AND a.id > b.id));

CREATE UNIQUE INDEX IF NOT EXISTS idx_bookmarks_user_url ON bookmarks (user_id, url);
