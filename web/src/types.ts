export type User = {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
};

export type Bookmark = {
  id: string;
  user_id: string;
  url: string;
  title: string;
  description: string;
  image_url: string;
  category: string;
  tags: string[];
  enriched: boolean;
  created_at: string;
  updated_at: string;
};

export type VerifyLoginResult = {
  tokens: { access_token: string };
  user: User;
};

export type PendingBookmark = {
  id: string;
  url: string;
  status: "pending" | "error";
  error?: string;
};

export const CATEGORY_LABELS: Record<string, string> = {
  movies: "Видео",
  articles: "Статьи",
  programming: "Программирование",
  shopping: "Покупки",
  gaming: "Игры",
  learning: "Обучение",
  music: "Музыка",
  news: "Новости",
  design: "Дизайн",
  tools: "Полезное",
  entertainment: "Развлечения",
  other: "Другое",
};

export function categoryLabel(category: string): string {
  return CATEGORY_LABELS[category] ?? category;
}
