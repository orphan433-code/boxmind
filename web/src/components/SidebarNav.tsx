import { GroupIcon } from "./GroupIcon";
import type { BrowseSection, BrowseSectionId } from "../utils/browseSections";

type Props = {
  sections: BrowseSection[];
  active: BrowseSectionId | null;
  counts: Partial<Record<BrowseSectionId, number>>;
  onChange: (sectionId: BrowseSectionId) => void;
};

export function SidebarNav({ sections, active, counts, onChange }: Props) {
  if (sections.length === 0) return null;

  const mainSections = sections.filter((section) => section.id !== "all");
  const allSection = sections.find((section) => section.id === "all");

  return (
    <nav className="sidebar-nav" aria-label="Разделы">
      <ul className="sidebar-nav-list">
        {mainSections.map((section) => (
          <SidebarItem
            key={section.id}
            section={section}
            active={active === section.id}
            count={counts[section.id] ?? 0}
            onChange={onChange}
          />
        ))}
      </ul>

      {allSection && (
        <>
          <div className="sidebar-nav-divider" />
          <ul className="sidebar-nav-list">
            <SidebarItem
              section={allSection}
              active={active === allSection.id}
              count={counts[allSection.id] ?? 0}
              onChange={onChange}
            />
          </ul>
        </>
      )}
    </nav>
  );
}

function SidebarItem({
  section,
  active,
  count,
  onChange,
}: {
  section: BrowseSection;
  active: boolean;
  count: number;
  onChange: (sectionId: BrowseSectionId) => void;
}) {
  return (
    <li>
      <button
        type="button"
        className={active ? "sidebar-nav-item active" : "sidebar-nav-item"}
        onClick={() => onChange(section.id)}
        aria-current={active ? "page" : undefined}
      >
        <span className="sidebar-nav-icon">
          <GroupIcon section={section.id} />
        </span>
        <span className="sidebar-nav-label">{section.label}</span>
        <span className="sidebar-nav-count">{count}</span>
      </button>
    </li>
  );
}
