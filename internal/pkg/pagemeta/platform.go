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

var contentTypeKeywords = map[string]string{
	"films":      "фильм",
	"film":       "фильм",
	"series":     "сериал",
	"serial":     "сериал",
	"serials":    "сериал",
	"cartoons":   "мультфильм",
	"cartoon":    "мультфильм",
	"animation":  "мультфильм",
	"multfilmy":  "мультфильм",
	"anime":      "аниме",
	"aniserials": "аниме",
	"video":      "видео",
	"videos":     "видео",
	"show":       "шоу",
	"shows":      "шоу",
}

// PlatformThumbnailURL returns a direct thumbnail URL for known platforms.
func PlatformThumbnailURL(rawURL string) string {
	if id := youtubeVideoID(rawURL); id != "" {
		return "https://i.ytimg.com/vi/" + id + "/hqdefault.jpg"
	}
	return ""
}

// GenericURLHints derives a title from the URL path and optionally category/tags
// when common media path keywords are present. Works for any site, not just one host.
func GenericURLHints(rawURL string) (domain.BookmarkEnrichment, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return domain.BookmarkEnrichment{}, false
	}

	segments := splitPathSegments(parsed.Path)
	if len(segments) == 0 {
		return domain.BookmarkEnrichment{}, false
	}

	slug := segments[len(segments)-1]
	if !isMeaningfulSlug(slug) {
		return domain.BookmarkEnrichment{}, false
	}

	title := titleFromSlug(slug)
	if title == "" {
		return domain.BookmarkEnrichment{}, false
	}

	enrichment := domain.BookmarkEnrichment{
		Title:    title,
		Category: "other",
	}

	contentType, genreSegment := detectMediaPath(segments)
	if contentType == "" {
		return enrichment, true
	}

	genreTag := "драма"
	if genreSegment != "" {
		if tag, ok := mediaGenreTags[strings.ToLower(genreSegment)]; ok {
			genreTag = tag
		}
	}

	enrichment.Category = "movies"
	enrichment.Tags = []string{contentType, genreTag}
	return enrichment, true
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

func detectMediaPath(segments []string) (contentType, genreSegment string) {
	for i, seg := range segments {
		if ct, ok := contentTypeKeywords[strings.ToLower(seg)]; ok {
			return ct, genreAt(segments, i)
		}
	}
	return "", ""
}

func genreAt(segments []string, kindIndex int) string {
	if kindIndex+1 >= len(segments)-1 {
		return ""
	}
	return segments[kindIndex+1]
}
