type Props = {
  tag: string;
  onClear: () => void;
};

export function ActiveTagFilter({ tag, onClear }: Props) {
  return (
    <div className="active-tag-filter">
      <button
        type="button"
        className="tag tag-btn active-tag-filter-chip"
        onClick={onClear}
        aria-label={`Сбросить фильтр «${tag}»`}
      >
        <span>{tag}</span>
        <span className="active-tag-filter-close" aria-hidden>
          ×
        </span>
      </button>
    </div>
  );
}
