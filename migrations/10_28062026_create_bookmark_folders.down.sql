DROP INDEX IF EXISTS idx_bookmarks_folder_id;

ALTER TABLE bookmarks
    DROP COLUMN IF EXISTS folder_id;

DROP INDEX IF EXISTS idx_bookmark_folders_user_id;

DROP TABLE IF EXISTS bookmark_folders;
