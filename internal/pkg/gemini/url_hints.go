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

	if isJobListingURL(pageURL) {
		switch enrichment.Category {
		case "programming", "articles", "other", "tools", "news":
			enrichment.Category = "jobs"
			enrichment.Tags = normalizeTagsForCategory("jobs", enrichment.Tags)
		}
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

func isJobListingURL(raw string) bool {
	u := strings.ToLower(raw)
	return strings.Contains(u, "hh.ru/vacancy") ||
		strings.Contains(u, "career.habr.com/vacancies") ||
		strings.Contains(u, "habr.com/vacancies/") ||
		strings.Contains(u, "linkedin.com/jobs/view") ||
		strings.Contains(u, "linkedin.com/jobs/collections") ||
		strings.Contains(u, "jobs.lever.co/") ||
		strings.Contains(u, "boards.greenhouse.io/") ||
		strings.Contains(u, "apply.workable.com/") ||
		strings.Contains(u, "indeed.com/viewjob") ||
		strings.Contains(u, "glassdoor.com/job-listing") ||
		strings.Contains(u, "superjob.ru/vakansii") ||
		strings.Contains(u, "zarplata.ru/vacancy")
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
