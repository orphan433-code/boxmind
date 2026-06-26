package service

import (
	"testing"

	"pet-link/internal/domain"
)

func TestTitleSourceForClassification(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		title string
		want  string
	}{
		{
			name:  "slug title",
			url:   "http://hdrezka.co/films/fantasy/123-garri-potter-i-uznik-azkabana-2004.html",
			title: "Garri Potter и Uznik Azkabana",
			want:  "url_slug",
		},
		{
			name:  "slug title action movie",
			url:   "http://hdrezka.co/films/action/90232-klinki-hraniteley-burya-v-pustyne-2026.html",
			title: "Klinki Hraniteley Burya V Pustyne",
			want:  "url_slug",
		},
		{
			name:  "trusted metadata",
			url:   "https://www.youtube.com/watch?v=abc123",
			title: "GOLANG ПОЛНЫЙ КУРС ДЛЯ НАЧИНАЮЩИХ",
			want:  "metadata_or_user",
		},
		{
			name:  "raw url title",
			url:   "https://example.com",
			title: "https://example.com",
			want:  "url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := titleSourceForClassification(tt.url, tt.title); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassificationCompleteForURLRequiresSlugTitleNormalization(t *testing.T) {
	enrichment := domain.BookmarkEnrichment{
		Title:    "Klinki Hraniteley Burya V Pustyne",
		Category: "movies",
		Tags:     []string{"фильм", "боевик"},
	}

	if classificationCompleteForURL("http://hdrezka.co/films/action/90232-klinki-hraniteley-burya-v-pustyne-2026.html", enrichment) {
		t.Fatal("expected slug title to require classify normalization")
	}

	enrichment.Title = "Blade of the Guardians"
	if !classificationCompleteForURL("https://www.youtube.com/watch?v=abc123", enrichment) {
		t.Fatal("expected trusted metadata title with category/tags to be complete")
	}
}

func TestMergeClassifiedEnrichmentPrefersClassifiedTitleForSlug(t *testing.T) {
	rawURL := "http://hdrezka.co/films/action/90232-klinki-hraniteley-burya-v-pustyne-2026.html"
	base := domain.BookmarkEnrichment{
		Title:    "Klinki Hraniteley Burya V Pustyne",
		Category: "movies",
		Tags:     []string{"фильм", "боевик"},
	}
	classified := domain.BookmarkEnrichment{
		Title:       "Клинки хранителей: Буря в пустыне",
		Description: "Фильм о приключениях и сражениях в пустыне.",
		Category:    "movies",
		Tags:        []string{"фильм", "боевик"},
	}

	got := mergeClassifiedEnrichment(rawURL, base, classified)
	if got.Title != classified.Title {
		t.Fatalf("title: got %q, want %q", got.Title, classified.Title)
	}
}

func TestMergeClassifiedEnrichmentKeepsTrustedMetadataTitle(t *testing.T) {
	rawURL := "https://www.youtube.com/watch?v=abc123"
	base := domain.BookmarkEnrichment{
		Title:    "Blade of the Guardians",
		Category: "movies",
		Tags:     []string{"фильм", "боевик"},
	}
	classified := domain.BookmarkEnrichment{
		Title:    "Клинки хранителей",
		Category: "movies",
		Tags:     []string{"фильм", "боевик"},
	}

	got := mergeClassifiedEnrichment(rawURL, base, classified)
	if got.Title != base.Title {
		t.Fatalf("title: got %q, want %q", got.Title, base.Title)
	}
}
