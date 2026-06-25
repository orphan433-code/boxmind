package pagemeta

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"pet-link/internal/domain"
)

var yearSuffixPattern = regexp.MustCompile(`-\d{4}$`)

var hdrezkaGenreTags = map[string]string{
	"fiction":    "фантастика",
	"fantasy":    "фэнтези",
	"thriller":   "триллер",
	"horrors":    "ужасы",
	"horror":     "ужасы",
	"dramas":     "драма",
	"drama":      "драма",
	"comedies":   "комедия",
	"comedy":     "комедия",
	"action":     "боевик",
	"melodramas": "мелодрама",
	"melodrama":  "мелодрама",
	"detective":  "детектив",
	"cartoons":   "мультфильм",
	"cartoon":    "мультфильм",
	"anime":      "аниме",
	"adventures": "приключения",
	"adventure":  "приключения",
	"biography":  "биография",
	"documentary": "документалка",
	"family":     "семейный",
	"military":   "военный",
	"sport":      "спорт",
}

// FallbackEnrichment tries HTTP metadata, then URL-based hints for known blocked sites.
func FallbackEnrichment(ctx context.Context, extractor Extractor, rawURL string) (domain.BookmarkEnrichment, bool) {
	if extractor != nil {
		page, err := extractor.Extract(ctx, rawURL)
		if err == nil && strings.TrimSpace(page.Title) != "" {
			enrichment := domain.BookmarkEnrichment{
				Title:       CleanPageTitle(page.Title),
				Description: page.Description,
			}
			if hints, ok := enrichmentFromKnownURL(rawURL); ok {
				enrichment.Category = hints.Category
				enrichment.Tags = hints.Tags
			}
			if enrichment.Category == "" {
				enrichment.Category = "other"
			}
			return enrichment, true
		}
	}

	return enrichmentFromKnownURL(rawURL)
}

func enrichmentFromKnownURL(rawURL string) (domain.BookmarkEnrichment, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return domain.BookmarkEnrichment{}, false
	}

	host := strings.ToLower(parsed.Hostname())
	if strings.Contains(host, "rezka") {
		return enrichmentFromHDRezkaURL(parsed)
	}

	return domain.BookmarkEnrichment{}, false
}

func enrichmentFromHDRezkaURL(parsed *url.URL) (domain.BookmarkEnrichment, bool) {
	segments := splitPathSegments(parsed.Path)
	if len(segments) < 3 {
		return domain.BookmarkEnrichment{}, false
	}

	kind := segments[0]
	if kind != "films" && kind != "series" {
		return domain.BookmarkEnrichment{}, false
	}

	contentType := "фильм"
	if kind == "series" {
		contentType = "сериал"
	}

	genreTag := "драма"
	if len(segments) >= 2 {
		if tag, ok := hdrezkaGenreTags[strings.ToLower(segments[1])]; ok {
			genreTag = tag
		}
	}

	slug := segments[len(segments)-1]
	title := titleFromHDRezkaSlug(slug)
	if title == "" {
		return domain.BookmarkEnrichment{}, false
	}

	return domain.BookmarkEnrichment{
		Title:    title,
		Category: "movies",
		Tags:     []string{contentType, genreTag},
	}, true
}

func splitPathSegments(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

func titleFromHDRezkaSlug(slug string) string {
	slug = strings.TrimSuffix(strings.ToLower(slug), ".html")
	slug = strings.TrimSuffix(slug, "-latest")
	slug = yearSuffixPattern.ReplaceAllString(slug, "")

	if dash := strings.Index(slug, "-"); dash > 0 && isDigits(slug[:dash]) {
		slug = slug[dash+1:]
	}

	words := strings.Split(slug, "-")
	parts := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" || word == "i" {
			if word == "i" {
				parts = append(parts, "и")
			}
			continue
		}
		parts = append(parts, titleCaseWord(word))
	}

	return strings.Join(parts, " ")
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func titleCaseWord(word string) string {
	if word == "" {
		return ""
	}
	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
