import type { Bookmark } from "../types";

export type BrowseSectionId =
  | "recent"
  | "watch"
  | "listen"
  | "read"
  | "learn"
  | "code"
  | "shop"
  | "services"
  | "work"
  | "other"
  | "all";

export type BrowseSection = {
  id: BrowseSectionId;
  label: string;
  categories: readonly string[];
  emptyHint: string;
};

export const RECENT_WINDOW_DAYS = 7;
export const RECENT_MAX_ITEMS = 20;

export const BROWSE_SECTIONS: BrowseSection[] = [
  {
    id: "recent",
    label: "Недавние",
    categories: [],
    emptyHint: "За последнюю неделю ты ничего не добавлял",
  },
  {
    id: "watch",
    label: "Смотреть",
    categories: ["movies", "entertainment", "gaming"],
    emptyHint: "Здесь будут фильмы, сериалы и аниме",
  },
  {
    id: "listen",
    label: "Слушать",
    categories: ["music"],
    emptyHint: "Здесь будет музыка — треки, альбомы, плейлисты",
  },
  {
    id: "read",
    label: "Читать",
    categories: ["articles", "news"],
    emptyHint: "Здесь будут статьи, обзоры и новости",
  },
  {
    id: "learn",
    label: "Учиться",
    categories: ["learning"],
    emptyHint: "Здесь будут курсы, туториалы и ноты",
  },
  {
    id: "code",
    label: "Полезное",
    categories: ["programming", "design", "tools"],
    emptyHint: "Здесь будут сервисы, документация, дизайн и рабочие ссылки",
  },
  {
    id: "shop",
    label: "Покупки",
    categories: ["shopping"],
    emptyHint: "Здесь будут товары и магазины",
  },
  {
    id: "services",
    label: "Услуги",
    categories: ["services"],
    emptyHint: "Здесь будут услуги и исполнители",
  },
  {
    id: "work",
    label: "Работа",
    categories: ["jobs"],
    emptyHint: "Здесь будут вакансии и карьерные возможности",
  },
  {
    id: "other",
    label: "Другое",
    categories: ["other"],
    emptyHint: "Здесь будут ссылки без чёткой категории",
  },
  {
    id: "all",
    label: "Всё",
    categories: [],
    emptyHint: "Пока нет сохранённых ссылок",
  },
];

const SECTION_BY_ID = Object.fromEntries(
  BROWSE_SECTIONS.map((section) => [section.id, section]),
) as Record<BrowseSectionId, BrowseSection>;

const CATEGORY_TO_SECTION = new Map<string, BrowseSectionId>();

for (const section of BROWSE_SECTIONS) {
  for (const category of section.categories) {
    CATEGORY_TO_SECTION.set(category, section.id);
  }
}

export function normalizeCategory(category: string): string {
  return category.trim() || "other";
}

export function sectionForCategory(category: string): BrowseSection | undefined {
  const sectionId = CATEGORY_TO_SECTION.get(normalizeCategory(category));
  if (!sectionId) return undefined;
  return SECTION_BY_ID[sectionId];
}

export function intentLabelForCategory(category: string): string {
  return sectionForCategory(category)?.label ?? "Другое";
}

export function isRecentSection(sectionId: BrowseSectionId): boolean {
  return sectionId === "recent";
}

export function isAllSection(sectionId: BrowseSectionId): boolean {
  return sectionId === "all";
}

export function isOtherSection(sectionId: BrowseSectionId): boolean {
  return sectionId === "other";
}

function isOtherCategory(category: string): boolean {
  const normalized = normalizeCategory(category);
  return normalized === "other" || !CATEGORY_TO_SECTION.has(normalized);
}

export function sortByNewest(bookmarks: Bookmark[]): Bookmark[] {
  return [...bookmarks].sort(
    (a, b) =>
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
  );
}

export function listRecentBookmarks(bookmarks: Bookmark[]): Bookmark[] {
  const sorted = sortByNewest(bookmarks);
  const cutoff = Date.now() - RECENT_WINDOW_DAYS * 24 * 60 * 60 * 1000;
  const inWindow = sorted.filter(
    (bookmark) => new Date(bookmark.created_at).getTime() >= cutoff,
  );

  if (inWindow.length > 0) {
    return inWindow.slice(0, RECENT_MAX_ITEMS);
  }

  return sorted.slice(0, Math.min(RECENT_MAX_ITEMS, 10));
}

export function bookmarkInSection(
  bookmark: Bookmark,
  sectionId: BrowseSectionId,
): boolean {
  if (isRecentSection(sectionId) || isAllSection(sectionId)) return true;
  if (isOtherSection(sectionId)) {
    return isOtherCategory(bookmark.category);
  }
  const section = SECTION_BY_ID[sectionId];
  return section.categories.includes(normalizeCategory(bookmark.category));
}

export function bookmarksForSection(
  bookmarks: Bookmark[],
  sectionId: BrowseSectionId,
): Bookmark[] {
  if (isRecentSection(sectionId)) {
    return listRecentBookmarks(bookmarks);
  }
  if (isAllSection(sectionId)) {
    return sortByNewest(bookmarks);
  }
  return sortByNewest(bookmarks.filter((b) => bookmarkInSection(b, sectionId)));
}

export function countForSection(
  bookmarks: Bookmark[],
  sectionId: BrowseSectionId,
): number {
  if (isRecentSection(sectionId)) {
    return listRecentBookmarks(bookmarks).length;
  }
  if (isAllSection(sectionId)) {
    return bookmarks.length;
  }
  return bookmarks.filter((b) => bookmarkInSection(b, sectionId)).length;
}

export function listActiveSections(bookmarks: Bookmark[]): BrowseSection[] {
  if (bookmarks.length === 0) return [];

  const result: BrowseSection[] = [SECTION_BY_ID.recent];
  for (const section of BROWSE_SECTIONS) {
    if (section.id === "recent" || section.id === "all") continue;
    if (countForSection(bookmarks, section.id) > 0) {
      result.push(section);
    }
  }
  result.push(SECTION_BY_ID.all);
  return result;
}

export function countBySection(
  bookmarks: Bookmark[],
): Partial<Record<BrowseSectionId, number>> {
  const sections = listActiveSections(bookmarks);
  const result: Partial<Record<BrowseSectionId, number>> = {};
  for (const section of sections) {
    result[section.id] = countForSection(bookmarks, section.id);
  }
  return result;
}

export function sectionPanelTitle(sectionId: BrowseSectionId): string {
  return SECTION_BY_ID[sectionId]?.label ?? "Закладки";
}

export function sectionPanelSubtitle(
  sectionId: BrowseSectionId,
): string | null {
  if (isRecentSection(sectionId)) return "за последнюю неделю";
  return null;
}

export function sectionEmptyHint(sectionId: BrowseSectionId): string {
  return SECTION_BY_ID[sectionId]?.emptyHint ?? "Ничего не найдено";
}

export function showIntentOnCards(sectionId: BrowseSectionId): boolean {
  return isRecentSection(sectionId) || isAllSection(sectionId);
}
