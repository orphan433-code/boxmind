import { useState, type CSSProperties } from "react";
import type { Bookmark, Folder } from "../types";
import { intentLabelForCategory } from "../utils/browseSections";
import { isBookmarkEnriching } from "../utils/enrichment";
import { FolderMenu } from "./FolderMenu";

type Props = {
  bookmark: Bookmark;
  index?: number;
  onDelete: (id: string) => void;
  onTagClick?: (tag: string) => void;
  activeTag?: string | null;
  deleting: boolean;
  showIntent?: boolean;
  folders?: Folder[];
  onAssignFolder?: (bookmarkId: string, folderId: string | null) => void;
};

export function BookmarkCard({
  bookmark,
  index = 0,
  onDelete,
  onTagClick,
  activeTag,
  deleting,
  showIntent = false,
  folders = [],
  onAssignFolder,
}: Props) {
  const [imageOk, setImageOk] = useState(Boolean(bookmark.image_url));
  const hasImage = imageOk && Boolean(bookmark.image_url);
  const enriching = isBookmarkEnriching(bookmark);

  const cardStyle = {
    "--card-index": Math.min(index, 14),
    ...(hasImage
      ? { "--card-image": `url("${bookmark.image_url}")` }
      : {}),
  } as CSSProperties;

  const canAssignFolder = Boolean(onAssignFolder) && folders.length > 0;
  const currentFolder = folders.find((folder) => folder.id === bookmark.folder_id);

  return (
    <article
      className={`bookmark-card cat-${bookmark.category || "other"}${hasImage ? " has-image" : ""}${deleting ? " is-deleting" : ""}`}
      style={cardStyle}
    >
      {bookmark.image_url && (
        <img
          src={bookmark.image_url}
          alt=""
          className="bookmark-card-preload"
          onLoad={() => setImageOk(true)}
          onError={() => setImageOk(false)}
        />
      )}

      <div className="bookmark-card-overlay" />

      <div className="bookmark-card-content">
        <div className="bookmark-card-top">
          <div className="bookmark-card-badges">
            {enriching ? (
              <span className="ai-badge" title="Дорабатываем карточку" aria-label="Дорабатываем карточку">
                <span className="ai-badge-spinner" aria-hidden="true" />
              </span>
            ) : (
              showIntent && (
                <span className="intent-badge">
                  {intentLabelForCategory(bookmark.category)}
                </span>
              )
            )}
            {currentFolder && (
              <span className="folder-chip" title={`Папка: ${currentFolder.name}`}>
                {currentFolder.name}
              </span>
            )}
          </div>
          <div className="bookmark-card-actions">
            {canAssignFolder && (
              <FolderMenu
                folders={folders}
                activeFolderId={bookmark.folder_id}
                onAssign={(folderId) => onAssignFolder?.(bookmark.id, folderId)}
              />
            )}
            <button
              type="button"
              className="icon-btn"
              onClick={() => onDelete(bookmark.id)}
              disabled={deleting}
              title="Удалить"
            >
              ×
            </button>
          </div>
        </div>

        <div className="bookmark-card-body">
          <h3>
            <a href={bookmark.url} target="_blank" rel="noreferrer">
              {bookmark.title || bookmark.url}
            </a>
          </h3>
          {bookmark.description && <p>{bookmark.description}</p>}
        </div>

        {bookmark.tags.length > 0 && (
          <div className="bookmark-card-footer">
            <div className="tag-row">
              {bookmark.tags.map((tag) => (
                <TagButton
                  key={tag}
                  tag={tag}
                  active={activeTag === tag}
                  onClick={onTagClick}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </article>
  );
}

function TagButton({
  tag,
  active,
  onClick,
}: {
  tag: string;
  active: boolean;
  onClick?: (tag: string) => void;
}) {
  if (!onClick) {
    return <span className="tag">{tag}</span>;
  }

  return (
    <button
      type="button"
      className={active ? "tag tag-btn active" : "tag tag-btn"}
      onClick={() => onClick(tag)}
    >
      {tag}
    </button>
  );
}
