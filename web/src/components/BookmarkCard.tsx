import { useState, type CSSProperties } from "react";
import type { Bookmark } from "../types";
import { intentLabelForCategory } from "../utils/browseSections";

type Props = {
  bookmark: Bookmark;
  onDelete: (id: string) => void;
  onTagClick?: (tag: string) => void;
  activeTag?: string | null;
  deleting: boolean;
  showIntent?: boolean;
};

export function BookmarkCard({
  bookmark,
  onDelete,
  onTagClick,
  activeTag,
  deleting,
  showIntent = false,
}: Props) {
  const [imageOk, setImageOk] = useState(Boolean(bookmark.image_url));
  const hasImage = imageOk && Boolean(bookmark.image_url);

  return (
    <article
      className={`bookmark-card cat-${bookmark.category || "other"}${hasImage ? " has-image" : ""}`}
      style={
        hasImage
          ? ({ "--card-image": `url("${bookmark.image_url}")` } as CSSProperties)
          : undefined
      }
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
            {showIntent && (
              <span className="intent-badge">
                {intentLabelForCategory(bookmark.category)}
              </span>
            )}
          </div>
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
