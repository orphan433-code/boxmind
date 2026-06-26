package service

import (
	"context"
	"errors"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/cardquality"
	"pet-link/internal/pkg/gemini"
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
	if !needsMovieMetadata(enrichment, imageURL) {
		log.Printf("[MOVIE-LOOKUP] skip reason=good_card url=%s category=%s image=%t", rawURL, enrichment.Category, imageURL != "")
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

	merged := mergeMovieMetadata(enrichment, meta.Enrichment, imageURL)
	if imageURL == "" && strings.TrimSpace(meta.ImageURL) != "" {
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

func needsMovieMetadata(enrichment domain.BookmarkEnrichment, imageURL string) bool {
	if !looksLikeMovie(enrichment) {
		return false
	}
	if strings.TrimSpace(imageURL) == "" {
		return true
	}
	if movieTitleNoise.MatchString(enrichment.Title) {
		return true
	}
	if !cardquality.GoodDescription(enrichment.Description) || genericMovieDescription(enrichment.Description) {
		return true
	}
	return !cardquality.IsAcceptable(enrichment, imageURL)
}

func looksLikeMovie(enrichment domain.BookmarkEnrichment) bool {
	if enrichment.Category == "movies" {
		return true
	}
	for _, tag := range enrichment.Tags {
		switch strings.ToLower(strings.TrimSpace(tag)) {
		case "фильм", "сериал", "аниме", "кино":
			return true
		}
	}
	return false
}

func mergeMovieMetadata(base, movie domain.BookmarkEnrichment, imageURL string) domain.BookmarkEnrichment {
	movie = gemini.NormalizeEnrichment(movie)
	out := base

	if movie.Title != "" && (!cardquality.GoodTitle(base.Title) || movieTitleNoise.MatchString(base.Title)) {
		out.Title = movie.Title
	}
	if movie.Description != "" && (!cardquality.GoodDescription(base.Description) || genericMovieDescription(base.Description)) {
		out.Description = movie.Description
	}
	if out.Category == "" || out.Category == "other" {
		out.Category = "movies"
	}
	if len(out.Tags) < 2 && len(movie.Tags) >= 2 {
		out.Tags = movie.Tags
	}
	if strings.TrimSpace(imageURL) == "" && movie.Title != "" && movie.Title != out.Title && movieTitleNoise.MatchString(out.Title) {
		out.Title = movie.Title
	}

	return out
}

func genericMovieDescription(description string) bool {
	value := strings.ToLower(strings.TrimSpace(description))
	switch value {
	case "", "это драматический сериал.", "это драматический фильм.", "это короткое видео.", "это сериал.", "это фильм.":
		return true
	default:
		return false
	}
}

func movieLookupTitle(enrichment domain.BookmarkEnrichment, rawURL string) string {
	title := strings.TrimSpace(enrichment.Title)
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
	parts := strings.FieldsFunc(rawURL, func(r rune) bool {
		return r == '/' || r == '-' || r == '_' || r == '.' || r == '?'
	})
	if len(parts) == 0 {
		return "", false
	}

	words := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || yearPattern.MatchString(part) {
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
