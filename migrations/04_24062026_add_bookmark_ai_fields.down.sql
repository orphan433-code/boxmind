DROP INDEX IF EXISTS idx_bookmarks_category;

ALTER TABLE bookmarks DROP COLUMN IF EXISTS tags;
ALTER TABLE bookmarks DROP COLUMN IF EXISTS category;
