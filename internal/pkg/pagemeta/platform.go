package pagemeta

import (
	"net/url"
	"strings"
)

// PlatformThumbnailURL returns a direct thumbnail URL for known platforms.
func PlatformThumbnailURL(rawURL string) string {
	if id := youtubeVideoID(rawURL); id != "" {
		return "https://i.ytimg.com/vi/" + id + "/hqdefault.jpg"
	}
	return ""
}

// TitleHintFromURL derives a placeholder title from the last URL path segment.
// It never returns category or tags — classification is handled later by AI.
func TitleHintFromURL(rawURL string) (string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}

	segments := splitPathSegments(parsed.Path)
	if len(segments) == 0 {
		return "", false
	}

	slug := segments[len(segments)-1]
	if !isMeaningfulSlug(slug) {
		return "", false
	}

	title := titleFromSlug(slug)
	if title == "" {
		return "", false
	}

	return title, true
}

func youtubeVideoID(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	switch {
	case strings.Contains(host, "youtube.com"), strings.Contains(host, "youtube-nocookie.com"):
		return strings.TrimSpace(parsed.Query().Get("v"))
	case host == "youtu.be":
		return strings.Trim(strings.TrimPrefix(parsed.Path, "/"), "/")
	default:
		return ""
	}
}
