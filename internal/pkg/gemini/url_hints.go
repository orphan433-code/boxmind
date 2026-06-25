package gemini

import (
	"strings"

	"pet-link/internal/domain"
)

func applyURLCategoryHints(pageURL string, enrichment domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	if isWatchableVideoURL(pageURL) {
		switch enrichment.Category {
		case "news", "articles":
			enrichment.Category = "entertainment"
			enrichment.Tags = fixMisclassifiedVideoTags(enrichment.Tags)
		}
	}

	if isDevProfileURL(pageURL) && enrichment.Category == "other" {
		enrichment.Category = "programming"
		enrichment.Tags = normalizeTags([]string{"профиль", "программирование"})
	}

	return enrichment
}

func isWatchableVideoURL(raw string) bool {
	u := strings.ToLower(raw)
	return strings.Contains(u, "youtube.com/") ||
		strings.Contains(u, "youtu.be/") ||
		strings.Contains(u, "rutube.ru/") ||
		strings.Contains(u, "vk.com/video") ||
		strings.Contains(u, "vkvideo.ru/")
}

func isDevProfileURL(raw string) bool {
	u := strings.ToLower(raw)
	return strings.Contains(u, "leetcode.com/u/") ||
		(strings.Contains(u, "github.com/") && !strings.Contains(u, "github.com/topics"))
}

func fixMisclassifiedVideoTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{"видео"}
	}

	fixed := make([]string, len(tags))
	copy(fixed, tags)

	switch fixed[0] {
	case "новость", "статья", "справка", "longread":
		fixed[0] = "видео"
	}

	return normalizeTags(fixed)
}
