package pagemeta

import (
	"context"
	"regexp"
	"strings"
	"unicode"

	"pet-link/internal/domain"
)

var yearSuffixPattern = regexp.MustCompile(`-\d{4}$`)

var junkSlugs = map[string]struct{}{
	"index":      {},
	"index.html": {},
	"login":      {},
	"register":   {},
	"signup":     {},
	"auth":       {},
	"search":     {},
	"home":       {},
	"api":        {},
	"watch":      {},
}

// FallbackEnrichment tries HTTP metadata, then URL-based hints.
func FallbackEnrichment(ctx context.Context, extractor Extractor, rawURL string) (domain.BookmarkEnrichment, bool) {
	if extractor != nil {
		page, err := extractor.Extract(ctx, rawURL)
		if err == nil && strings.TrimSpace(page.Title) != "" {
			enrichment := domain.BookmarkEnrichment{
				Title:       CleanPageTitle(page.Title),
				Description: page.Description,
			}
			if hints, ok := enrichmentFromKnownURL(rawURL); ok {
				if enrichment.Title == "" {
					enrichment.Title = hints.Title
				}
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
	return GenericURLHints(rawURL)
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

func isMeaningfulSlug(slug string) bool {
	normalized := strings.ToLower(strings.TrimSpace(slug))
	if normalized == "" {
		return false
	}
	if _, junk := junkSlugs[normalized]; junk {
		return false
	}

	clean := strings.TrimSuffix(normalized, ".html")
	clean = strings.TrimSuffix(clean, "-latest")
	clean = yearSuffixPattern.ReplaceAllString(clean, "")

	if dash := strings.Index(clean, "-"); dash > 0 && isDigits(clean[:dash]) {
		clean = clean[dash+1:]
	}

	if len([]rune(clean)) < 4 {
		return false
	}

	hasLetter := false
	for _, r := range clean {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}
	return hasLetter
}

func titleFromSlug(slug string) string {
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
