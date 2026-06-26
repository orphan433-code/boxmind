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

// PlatformTag returns a display tag for known social/video platforms.
func PlatformTag(rawURL string) (string, bool) {
	host, ok := normalizedHost(rawURL)
	if !ok {
		return "", false
	}

	switch {
	case isTikTokHost(host):
		return "TikTok", true
	case isInstagramHost(host):
		return "Instagram", true
	case isYouTubeHost(host):
		return "YouTube", true
	default:
		return "", false
	}
}

// EnsurePlatformTag prepends a platform tag when the URL belongs to a known site.
// The tag is placed first so UIs that show only the first tags stay readable.
func EnsurePlatformTag(rawURL string, tags []string) []string {
	platform, ok := PlatformTag(rawURL)
	if !ok {
		return tags
	}

	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), platform) {
			return tags
		}
	}

	out := make([]string, 0, len(tags)+1)
	out = append(out, platform)
	out = append(out, tags...)
	return out
}

func normalizedHost(rawURL string) (string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return "", false
	}
	return host, true
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
	case isYouTubeHost(host):
		return strings.TrimSpace(parsed.Query().Get("v"))
	case host == "youtu.be":
		return strings.Trim(strings.TrimPrefix(parsed.Path, "/"), "/")
	default:
		return ""
	}
}

// isYouTubeHost matches youtube.com, m.youtube.com, youtube-nocookie.com,
// and regional domains like youtube.kz or youtube.co.uk.
func isYouTubeHost(host string) bool {
	host = strings.TrimPrefix(strings.ToLower(host), "www.")
	if host == "youtu.be" {
		return true
	}
	if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtube-nocookie.com") {
		return true
	}
	return strings.HasPrefix(host, "youtube.")
}

// isTikTokHost matches tiktok.com, vt.tiktok.com, vm.tiktok.com,
// and regional domains like tiktok.kz if they exist.
func isTikTokHost(host string) bool {
	host = strings.TrimPrefix(strings.ToLower(host), "www.")
	if strings.Contains(host, "tiktok.com") {
		return true
	}
	return strings.HasPrefix(host, "tiktok.")
}

// isInstagramHost matches instagram.com, m.instagram.com, instagr.am,
// and regional domains like instagram.co.uk.
func isInstagramHost(host string) bool {
	host = strings.TrimPrefix(strings.ToLower(host), "www.")
	if host == "instagr.am" {
		return true
	}
	if strings.Contains(host, "instagram.com") {
		return true
	}
	return strings.HasPrefix(host, "instagram.")
}
