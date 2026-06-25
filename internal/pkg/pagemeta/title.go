package pagemeta

import "strings"

var seoTitleSuffixes = []string{
	" — смотреть аниме онлайн",
	" - смотреть аниме онлайн",
	" — смотреть онлайн",
	" - смотреть онлайн",
	" — смотреть сериал онлайн",
	" - смотреть сериал онлайн",
	" — смотреть фильм онлайн",
	" - смотреть фильм онлайн",
	" смотреть онлайн",
	" watch online",
}

// CleanPageTitle strips common SEO suffixes from HTML page titles.
func CleanPageTitle(raw string) string {
	title := strings.TrimSpace(raw)
	if title == "" {
		return ""
	}

	lower := strings.ToLower(title)
	for _, suffix := range seoTitleSuffixes {
		if idx := strings.Index(lower, strings.ToLower(suffix)); idx > 0 {
			title = strings.TrimSpace(title[:idx])
			lower = strings.ToLower(title)
		}
	}

	if parts := strings.Split(title, "|"); len(parts) > 1 {
		left := strings.TrimSpace(parts[0])
		if left != "" && len([]rune(left)) >= 3 {
			title = left
		}
	}

	return title
}
