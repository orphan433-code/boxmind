CREATE TABLE IF NOT EXISTS bookmark_folders (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_bookmark_folders_user_id ON bookmark_folders (user_id);

ALTER TABLE bookmarks
    ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES bookmark_folders (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_bookmarks_folder_id ON bookmarks (folder_id);
