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
	" - youtube",
	" — youtube",
	" - vimeo",
}

// siteNameOnly lists titles that are just a platform name with no real content.
var siteNameOnly = map[string]struct{}{
	"youtube":   {},
	"vimeo":     {},
	"rutube":    {},
	"vk":        {},
	"vk видео":  {},
	"кинопоиск": {},
	"twitch":    {},
	"tiktok":    {},
	"instagram": {},
	"facebook":  {},
}

// CleanPageTitle strips common SEO suffixes and site-name junk from titles.
func CleanPageTitle(raw string) string {
	title := strings.TrimSpace(raw)
	if title == "" {
		return ""
	}

	lower := strings.ToLower(title)
	for _, suffix := range seoTitleSuffixes {
		if idx := strings.Index(lower, suffix); idx > 0 {
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

	// Drop leading separators left over from "- YouTube"-style titles.
	title = strings.TrimSpace(strings.TrimLeft(title, "-—–|· "))

	if _, ok := siteNameOnly[strings.ToLower(title)]; ok {
		return ""
	}

	return title
}
