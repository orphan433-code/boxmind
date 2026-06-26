package service

import (
	"testing"

	"pet-link/internal/domain"
)

func TestShouldAttemptMovieLookup(t *testing.T) {
	tests := []struct {
		name       string
		enrichment domain.BookmarkEnrichment
		imageURL   string
		want       bool
	}{
		{
			name: "movie tag always attempts even with good card",
			enrichment: domain.BookmarkEnrichment{
				Title:       "Побег из Шоушенка",
				Description: "Драма о надежде и дружбе в тюрьме.",
				Category:    "movies",
				Tags:        []string{"фильм", "драма"},
			},
			imageURL: "https://example.com/poster.jpg",
			want:     true,
		},
		{
			name: "non movie",
			enrichment: domain.BookmarkEnrichment{
				Title:       "Документация Go",
				Description: "Справочник по языку Go.",
				Category:    "programming",
				Tags:        []string{"документация", "golang"},
			},
			want: false,
		},
		{
			name: "misclassified entertainment with source-only description",
			enrichment: domain.BookmarkEnrichment{
				Title:       "Коммерсантъ",
				Description: "Видеоматериал от известного российского медиахолдинга.",
				Category:    "entertainment",
				Tags:        []string{"видео", "новости"},
			},
			want: true,
		},
		{
			name: "generic movie description with film tag",
			enrichment: domain.BookmarkEnrichment{
				Title:       "Marama",
				Description: "Фильм или сериал, доступный для просмотра онлайн.",
				Category:    "entertainment",
				Tags:        []string{"фильм", "драма"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldAttemptMovieLookup(tt.enrichment, tt.imageURL); got != tt.want {
				t.Fatalf("shouldAttemptMovieLookup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeMovieMetadataConservatively(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Побег из Шоушенка",
		Description: "Драма о надежде и дружбе в тюрьме.",
		Category:    "movies",
		Tags:        []string{"фильм", "драма"},
	}
	movie := domain.BookmarkEnrichment{
		Title:       "The Shawshank Redemption",
		Description: "История заключённого, который сохраняет надежду.",
		Category:    "movies",
		Tags:        []string{"фильм", "драма"},
	}

	got := mergeMovieMetadata(base, movie, "https://example.com/poster.jpg", 0.95)
	if got.Title != base.Title {
		t.Fatalf("title changed: got %q, want %q", got.Title, base.Title)
	}
	if got.Description != base.Description {
		t.Fatalf("description changed: got %q, want %q", got.Description, base.Description)
	}
}

func TestMergeMovieMetadataFixesWeakFields(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Ноты для фортепиано, голоса, легкие ноты - скачать бесплатно",
		Description: "Это драматический сериал.",
		Category:    "movies",
		Tags:        []string{"сериал", "драма"},
	}
	movie := domain.BookmarkEnrichment{
		Title:       "Пусть послужит вам уроком",
		Description: "Драма о людях, чьи ошибки становятся уроком.",
		Category:    "movies",
		Tags:        []string{"сериал", "драма"},
	}

	got := mergeMovieMetadata(base, movie, "", 0.90)
	if got.Title != movie.Title {
		t.Fatalf("title = %q, want %q", got.Title, movie.Title)
	}
	if got.Description != movie.Description {
		t.Fatalf("description = %q, want %q", got.Description, movie.Description)
	}
}

func TestMergeMovieMetadataLowConfidenceKeepsGoodText(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Коммерсантъ",
		Description: "Видеоматериал от известного российского медиахолдинга.",
		Category:    "entertainment",
		Tags:        []string{"видео", "новости"},
	}
	movie := domain.BookmarkEnrichment{
		Title:       "Wrong Match",
		Description: "Совсем другой сюжет.",
		Category:    "movies",
		Tags:        []string{"фильм", "драма"},
	}

	got := mergeMovieMetadata(base, movie, "", 0.58)
	if got.Title != base.Title {
		t.Fatalf("title changed at low confidence: got %q", got.Title)
	}
	if got.Description != base.Description {
		t.Fatalf("description changed at low confidence: got %q", got.Description)
	}
}

func TestGenericMovieDescription(t *testing.T) {
	cases := []string{
		"Фильм или сериал, доступный для просмотра онлайн.",
		"Это драматический сериал.",
		"Это фильм или сериал, доступный для просмотра онлайн.",
	}
	for _, description := range cases {
		if !genericMovieDescription(description) {
			t.Fatalf("genericMovieDescription(%q) = false, want true", description)
		}
	}
}

func TestTitleHintFromURLForMovieStripsHostExtensionAndFragment(t *testing.T) {
	got, ok := titleHintFromURLForMovie("https://mix.kinogo.mu/125718-kommersant.html#125718")
	if !ok {
		t.Fatal("titleHintFromURLForMovie() returned false")
	}
	if got != "kommersant" {
		t.Fatalf("titleHintFromURLForMovie() = %q, want %q", got, "kommersant")
	}
}
