package service

import (
	"context"
	"errors"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/cardquality"
	"pet-link/internal/pkg/gemini"
)

// Confidence tiers: always query TMDB for movie-like cards, but only apply
// fields when the match score clears the relevant bar.
const (
	minMovieConfidencePoster = 0.55
	minMovieConfidenceDesc   = 0.62
	minMovieConfidenceTitle  = 0.75
	minMovieConfidenceMeta   = 0.85
)

// isNilProvider guards against a typed-nil interface (e.g. a nil *TMDBProvider
// wrapped in a non-nil MovieMetadataProvider interface).
func isNilProvider(p MovieMetadataProvider) bool {
	if p == nil {
		return true
	}
	value := reflect.ValueOf(p)
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return value.IsNil()
	default:
		return false
	}
}

var (
	movieTitleNoise = regexp.MustCompile(`(?i)(смотреть\s+онлайн|скачать\s+бесплатно|скачать|бесплатно|без\s+регистрации|watch\s+online|free|1080p|720p|hd)`)
	yearPattern     = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	titleYearSuffix = regexp.MustCompile(`\s*[\(\[]\s*(19\d{2}|20\d{2})\s*[\)\]]\s*$`)

	genericMovieDescRe = regexp.MustCompile(`(?i)(доступн[а-яё]*\s+для\s+просмотра|для\s+просмотра\s+онлайн|смотреть\s+онлайн|фильм\s+или\s+сериал|это\s+(драматическ[а-яё]+\s+)?(короткое\s+видео|фильм|сериал|аниме))`)
	sourceFocusedDescRe = regexp.MustCompile(`(?i)(видеоматериал|медиахолдинг|новостн[а-яё]*\s+(лент[а-яё]*|сайт|портал)|из\s+известн[а-яё]*\s+(российск[а-яё]*\s+)?(медиа|издан))`)
)

func (s *bookmarkService) applyMovieMetadataIfNeeded(ctx context.Context, userID, bookmarkID, rawURL string, enrichment domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	if isNilProvider(s.movieMeta) {
		return enrichment
	}

	bookmark, err := s.repo.GetByIDForUser(ctx, userID, bookmarkID)
	if err != nil {
		return enrichment
	}

	imageURL := strings.TrimSpace(bookmark.ImageURL)
	if !shouldAttemptMovieLookup(enrichment, imageURL) {
		return enrichment
	}

	query := domain.MovieMetadataQuery{
		Title: movieLookupTitle(enrichment, rawURL),
		Year:  movieLookupYear(enrichment.Title, rawURL),
		Kind:  movieLookupKind(enrichment),
	}
	if strings.TrimSpace(query.Title) == "" {
		log.Printf("[MOVIE-LOOKUP] skip reason=no_title url=%s", rawURL)
		return enrichment
	}

	meta, ok := s.movieMeta.Lookup(ctx, query)
	if !ok {
		log.Printf("[MOVIE-LOOKUP] miss query=%q url=%s", query.Title, rawURL)
		return enrichment
	}

	if meta.Confidence < minMovieConfidencePoster {
		log.Printf(
			"[MOVIE-LOOKUP] low_confidence query=%q url=%s confidence=%.2f",
			query.Title,
			rawURL,
			meta.Confidence,
		)
		return enrichment
	}

	merged := mergeMovieMetadata(enrichment, meta.Enrichment, imageURL, meta.Confidence)
	if imageURL == "" && strings.TrimSpace(meta.ImageURL) != "" && meta.Confidence >= minMovieConfidencePoster {
		if err := s.repo.UpdateImageURL(ctx, userID, bookmarkID, meta.ImageURL); err != nil && !errors.Is(err, domain.ErrBookmarkNotFound) {
			log.Printf("[MOVIE-LOOKUP] image update failed url=%s err=%v", rawURL, err)
		} else {
			imageURL = meta.ImageURL
		}
	}

	log.Printf(
		"[MOVIE-LOOKUP] hit query=%q title=%q confidence=%.2f poster=%t",
		query.Title,
		meta.Enrichment.Title,
		meta.Confidence,
		strings.TrimSpace(meta.ImageURL) != "",
	)

	s.storeEnrichmentCache(ctx, rawURL, merged, imageURL)
	return merged
}

// shouldAttemptMovieLookup decides whether to call TMDB. We query whenever the
// card looks movie-related, without skipping "good enough" cards upfront —
// confidence tiers in merge decide what to apply.
func shouldAttemptMovieLookup(enrichment domain.BookmarkEnrichment, imageURL string) bool {
	if looksLikeMovie(enrichment) {
		return true
	}
	return probableMovieCard(enrichment, imageURL)
}

func probableMovieCard(enrichment domain.BookmarkEnrichment, imageURL string) bool {
	category := strings.TrimSpace(enrichment.Category)
	if category != "movies" && category != "entertainment" {
		return false
	}
	if strings.TrimSpace(enrichment.Title) == "" {
		return false
	}
	if strings.TrimSpace(imageURL) == "" {
		return true
	}
	if genericMovieDescription(enrichment.Description) || sourceFocusedDescription(enrichment.Description) {
		return true
	}
	return !cardquality.IsGoodEnough(enrichment, imageURL)
}

func looksLikeMovie(enrichment domain.BookmarkEnrichment) bool {
	if enrichment.Category == "movies" {
		return true
	}
	for _, tag := range enrichment.Tags {
		switch strings.ToLower(strings.TrimSpace(tag)) {
		case "фильм", "сериал", "аниме", "кино",
			"драма", "комедия", "боевик", "триллер", "ужасы",
			"фантастика", "приключения", "детектив", "романтика",
			"мультфильм", "семейное", "документальное":
			return true
		}
	}
	return false
}

func mergeMovieMetadata(base, movie domain.BookmarkEnrichment, imageURL string, confidence float64) domain.BookmarkEnrichment {
	movie = gemini.NormalizeEnrichment(movie)
	out := base

	if confidence >= minMovieConfidenceTitle && movie.Title != "" {
		if !cardquality.GoodTitle(base.Title) || movieTitleNoise.MatchString(base.Title) {
			out.Title = movie.Title
		}
	}
	if confidence >= minMovieConfidenceDesc && movie.Description != "" {
		if !cardquality.GoodDescription(base.Description) ||
			genericMovieDescription(base.Description) ||
			sourceFocusedDescription(base.Description) {
			out.Description = movie.Description
		}
	}
	if confidence >= minMovieConfidenceMeta {
		if out.Category == "" || out.Category == "other" {
			out.Category = "movies"
		}
		if len(out.Tags) < 2 && len(movie.Tags) >= 2 {
			out.Tags = movie.Tags
		}
	}
	if strings.TrimSpace(imageURL) == "" && confidence >= minMovieConfidenceTitle &&
		movie.Title != "" && movie.Title != out.Title && movieTitleNoise.MatchString(out.Title) {
		out.Title = movie.Title
	}

	return out
}

func genericMovieDescription(description string) bool {
	value := strings.TrimSpace(description)
	if value == "" {
		return true
	}
	return genericMovieDescRe.MatchString(value)
}

func sourceFocusedDescription(description string) bool {
	return sourceFocusedDescRe.MatchString(strings.TrimSpace(description))
}

func movieLookupTitle(enrichment domain.BookmarkEnrichment, rawURL string) string {
	title := strings.TrimSpace(enrichment.Title)
	title = titleYearSuffix.ReplaceAllString(title, "")
	title = movieTitleNoise.ReplaceAllString(title, " ")
	title = strings.Join(strings.Fields(title), " ")
	if title != "" {
		return title
	}
	if hint, ok := titleHintFromURLForMovie(rawURL); ok {
		return hint
	}
	return ""
}

func movieLookupKind(enrichment domain.BookmarkEnrichment) string {
	for _, tag := range enrichment.Tags {
		switch strings.ToLower(strings.TrimSpace(tag)) {
		case "фильм":
			return "movie"
		case "сериал", "аниме":
			return "series"
		}
	}
	return ""
}

func movieLookupYear(values ...string) int {
	for _, value := range values {
		match := yearPattern.FindString(value)
		if match == "" {
			continue
		}
		year, err := strconv.Atoi(match)
		if err == nil {
			return year
		}
	}
	return 0
}

func titleHintFromURLForMovie(rawURL string) (string, bool) {
	value := strings.TrimSpace(rawURL)
	if parsed, err := url.Parse(value); err == nil && parsed.Path != "" {
		value = parsed.Path
	}

	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '/' || r == '-' || r == '_' || r == '.' || r == '?' || r == '#'
	})
	if len(parts) == 0 {
		return "", false
	}

	words := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || yearPattern.MatchString(part) || isURLTitleJunk(part) {
			continue
		}
		if _, err := strconv.Atoi(part); err == nil {
			continue
		}
		words = append(words, part)
	}
	if len(words) == 0 {
		return "", false
	}
	return strings.Join(words, " "), true
}

func isURLTitleJunk(part string) bool {
	switch strings.ToLower(strings.TrimSpace(part)) {
	case "http", "https", "www", "html", "htm", "php", "asp", "aspx", "watch", "online":
		return true
	default:
		return false
	}
}
