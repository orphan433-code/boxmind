package pagemeta

import (
	"net/url"
	"strings"

	"pet-link/internal/domain"
)

var mediaGenreTags = map[string]string{
	"fiction":     "фантастика",
	"fantasy":     "фэнтези",
	"fant":        "фэнтези",
	"thriller":    "триллер",
	"horrors":     "ужасы",
	"horror":      "ужасы",
	"dramas":      "драма",
	"drama":       "драма",
	"comedies":    "комедия",
	"comedy":      "комедия",
	"action":      "боевик",
	"melodramas":  "мелодрама",
	"melodrama":   "мелодрама",
	"detective":   "детектив",
	"cartoons":    "мультфильм",
	"cartoon":     "мультфильм",
	"anime":       "аниме",
	"adventures":  "приключения",
	"adventure":   "приключения",
	"biography":   "биография",
	"documentary": "документалка",
	"family":      "семейный",
	"military":    "военный",
	"sport":       "спорт",
}

// PlatformThumbnailURL returns a direct thumbnail URL for known platforms.
func PlatformThumbnailURL(rawURL string) string {
	if id := youtubeVideoID(rawURL); id != "" {
		return "https://i.ytimg.com/vi/" + id + "/hqdefault.jpg"
	}
	return ""
}

// GenericURLHints derives title/category/tags from common media URL shapes.
func GenericURLHints(rawURL string) (domain.BookmarkEnrichment, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return domain.BookmarkEnrichment{}, false
	}

	host := strings.ToLower(parsed.Hostname())
	if strings.Contains(host, "rezka") {
		return enrichmentFromHDRezkaURL(parsed)
	}

	segments := splitPathSegments(parsed.Path)
	if len(segments) < 2 {
		return domain.BookmarkEnrichment{}, false
	}

	contentType, genreSegment, slug := mediaPathParts(segments)
	if contentType == "" || slug == "" {
		return domain.BookmarkEnrichment{}, false
	}

	title := titleFromMediaSlug(slug)
	if title == "" {
		return domain.BookmarkEnrichment{}, false
	}

	firstTag := contentType
	if strings.Contains(host, "anime") || strings.Contains(strings.Join(segments, "/"), "anime") {
		firstTag = "аниме"
	}

	genreTag := "драма"
	if genreSegment != "" {
		if tag, ok := mediaGenreTags[strings.ToLower(genreSegment)]; ok {
			genreTag = tag
		}
	}

	return domain.BookmarkEnrichment{
		Title:    title,
		Category: "movies",
		Tags:     []string{firstTag, genreTag},
	}, true
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

func mediaPathParts(segments []string) (contentType, genreSegment, slug string) {
	for i, seg := range segments {
		switch strings.ToLower(seg) {
		case "films", "film":
			return "фильм", genreAt(segments, i), lastSlug(segments)
		case "series", "serial", "serials":
			return "сериал", genreAt(segments, i), lastSlug(segments)
		case "anime", "aniserials", "cartoon", "cartoons", "video", "videos", "watch":
			return "аниме", genreAt(segments, i), lastSlug(segments)
		}
	}
	return "", "", ""
}

func genreAt(segments []string, kindIndex int) string {
	if kindIndex+1 >= len(segments)-1 {
		return ""
	}
	return segments[kindIndex+1]
}

func lastSlug(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	return segments[len(segments)-1]
}

func titleFromMediaSlug(slug string) string {
	return titleFromHDRezkaSlug(slug)
}
