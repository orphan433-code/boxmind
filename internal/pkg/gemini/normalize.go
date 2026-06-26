package gemini

import (
	"strings"
	"unicode/utf8"

	"pet-link/internal/domain"
)

const (
	maxTitleRunes       = 70
	maxDescriptionRunes = 100
	maxTags             = 2
)

var allowedCategories = map[string]struct{}{
	"programming":   {},
	"design":        {},
	"news":          {},
	"movies":        {},
	"shopping":      {},
	"articles":      {},
	"learning":      {},
	"music":         {},
	"tools":         {},
	"entertainment": {},
	"other":         {},
}

var categoryAliases = map[string]string{
	"film":        "movies",
	"films":       "movies",
	"cinema":      "movies",
	"video":       "entertainment",
	"hardware":    "shopping",
	"marketplace": "shopping",
	"tech":        "programming",
	"software":    "programming",
	"education":   "learning",
	"wiki":        "articles",
	"reference":   "articles",
	"article":     "articles",
	"sheet-music": "learning",
	"game":        "entertainment",
	"gaming":      "entertainment",
}

var tagAliases = map[string]string{
	"tutorial":      "туториал",
	"guide":         "гайд",
	"lesson":        "урок",
	"course":        "курс",
	"review":        "обзор",
	"article":       "статья",
	"sheet-music":   "ноты",
	"sheetmusic":    "ноты",
	"score":         "ноты",
	"scores":        "ноты",
	"tabs":          "ноты",
	"song":          "песня",
	"track":         "трек",
	"album":         "альбом",
	"playlist":      "плейлист",
	"artist":        "артист",
	"profile":       "профиль",
	"repository":    "репозиторий",
	"repo":          "репозиторий",
	"tool":          "инструмент",
	"docs":          "документация",
	"documentation": "документация",
}

func NormalizeEnrichment(enrichment domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	category := normalizeCategory(enrichment.Category)
	return domain.BookmarkEnrichment{
		Title:       normalizeTitle(enrichment.Title),
		Description: normalizeDescription(enrichment.Description),
		Category:    category,
		Tags:        normalizeTagsForCategory(category, enrichment.Tags),
	}
}

func normalizeCategory(raw string) string {
	category := strings.ToLower(strings.TrimSpace(raw))
	if category == "" {
		return "other"
	}

	if alias, ok := categoryAliases[category]; ok {
		category = alias
	}

	if _, ok := allowedCategories[category]; ok {
		return category
	}

	return "other"
}

func normalizeTitle(raw string) string {
	return truncateRunes(strings.TrimSpace(raw), maxTitleRunes)
}

func normalizeDescription(raw string) string {
	description := strings.TrimSpace(raw)
	description = strings.ReplaceAll(description, "\n", " ")
	description = strings.Join(strings.Fields(description), " ")

	if end := firstSentenceEnd(description); end > 0 {
		description = strings.TrimSpace(description[:end])
	}

	if utf8.RuneCountInString(description) <= maxDescriptionRunes {
		return description
	}

	return truncateAtWord(description, maxDescriptionRunes)
}

func firstSentenceEnd(value string) int {
	for i, r := range value {
		if r == '.' || r == '!' || r == '?' {
			return i + 1
		}
	}
	return 0
}

func truncateAtWord(value string, limit int) string {
	if limit <= 0 || value == "" {
		return value
	}

	if utf8.RuneCountInString(value) <= limit {
		return value
	}

	runes := []rune(value)
	cut := string(runes[:limit])
	if lastSpace := strings.LastIndex(cut, " "); lastSpace > limit/2 {
		cut = cut[:lastSpace]
	}

	return strings.TrimSpace(cut) + "…"
}

func normalizeTags(tags []string) []string {
	result := make([]string, 0, maxTags)
	seen := make(map[string]struct{}, len(tags))

	for _, tag := range tags {
		tag = slugifyTag(tag)
		if alias, ok := tagAliases[tag]; ok {
			tag = alias
		}
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)

		if len(result) >= maxTags {
			break
		}
	}

	return result
}

func normalizeTagsForCategory(category string, tags []string) []string {
	normalized := normalizeTags(tags)
	if len(normalized) == 0 {
		return normalized
	}

	switch category {
	case "programming", "design", "tools":
		return normalizeUsefulTags(normalized)
	default:
		return normalized
	}
}

func normalizeUsefulTags(tags []string) []string {
	firstAllowed := map[string]struct{}{
		"профиль":      {},
		"документация": {},
		"инструмент":   {},
		"репозиторий":  {},
	}
	secondAliases := map[string]string{
		"programming":  "программирование",
		"dev":          "программирование",
		"frontend":     "frontend",
		"backend":      "backend",
		"infra":        "infra",
		"devops":       "devops",
		"cloud":        "облако",
		"design":       "дизайн",
		"productivity": "продуктивность",
		"analytics":    "аналитика",
	}
	secondAllowed := map[string]struct{}{
		"программирование": {},
		"frontend":         {},
		"backend":          {},
		"infra":            {},
		"devops":           {},
		"облако":           {},
		"дизайн":           {},
		"продуктивность":   {},
		"аналитика":        {},
	}

	first := tags[0]
	if _, ok := firstAllowed[first]; !ok {
		first = "инструмент"
	}

	second := "программирование"
	if len(tags) > 1 {
		raw := tags[1]
		if alias, ok := secondAliases[raw]; ok {
			raw = alias
		}
		if _, ok := secondAllowed[raw]; ok {
			second = raw
		}
	}

	return []string{first, second}
}

func slugifyTag(raw string) string {
	tag := strings.ToLower(strings.TrimSpace(raw))
	tag = strings.ReplaceAll(tag, "_", "-")
	tag = strings.Join(strings.Fields(tag), "-")

	for strings.Contains(tag, "--") {
		tag = strings.ReplaceAll(tag, "--", "-")
	}

	return strings.Trim(tag, "-")
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 || value == "" {
		return value
	}

	if utf8.RuneCountInString(value) <= limit {
		return value
	}

	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}
