package service

import (
	"testing"

	"pet-link/internal/domain"
)

func TestNeedsMovieMetadata(t *testing.T) {
	tests := []struct {
		name       string
		enrichment domain.BookmarkEnrichment
		imageURL   string
		want       bool
	}{
		{
			name: "movie without image",
			enrichment: domain.BookmarkEnrichment{
				Title:       "Пусть послужит вам уроком",
				Description: "Это драматический сериал.",
				Category:    "movies",
				Tags:        []string{"сериал", "драма"},
			},
			want: true,
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
			name: "movie with seo title",
			enrichment: domain.BookmarkEnrichment{
				Title:       "Ноты для фортепиано - скачать бесплатно",
				Description: "Подборка нот для фортепиано.",
				Category:    "movies",
				Tags:        []string{"фильм", "драма"},
			},
			imageURL: "https://example.com/poster.jpg",
			want:     true,
		},
		{
			name: "good movie card",
			enrichment: domain.BookmarkEnrichment{
				Title:       "Побег из Шоушенка",
				Description: "Драма о надежде и дружбе в тюрьме.",
				Category:    "movies",
				Tags:        []string{"фильм", "драма"},
			},
			imageURL: "https://example.com/poster.jpg",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsMovieMetadata(tt.enrichment, tt.imageURL); got != tt.want {
				t.Fatalf("needsMovieMetadata() = %v, want %v", got, tt.want)
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

	got := mergeMovieMetadata(base, movie, "https://example.com/poster.jpg")
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

	got := mergeMovieMetadata(base, movie, "")
	if got.Title != movie.Title {
		t.Fatalf("title = %q, want %q", got.Title, movie.Title)
	}
	if got.Description != movie.Description {
		t.Fatalf("description = %q, want %q", got.Description, movie.Description)
	}
}
