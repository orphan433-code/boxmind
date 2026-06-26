package service

import (
	"testing"

	"pet-link/internal/domain"
)

func TestIsUsableCacheEntry(t *testing.T) {
	tests := []struct {
		name  string
		entry domain.URLEnrichmentCacheEntry
		want  bool
	}{
		{
			name: "complete entry",
			entry: domain.URLEnrichmentCacheEntry{
				Title:       "GOLANG ПОЛНЫЙ КУРС",
				Description: "Курс по Go для начинающих.",
				Category:    "learning",
				Tags:        []string{"курс", "golang"},
				ImageURL:    "https://img.example/thumb.jpg",
			},
			want: true,
		},
		{
			name: "other category",
			entry: domain.URLEnrichmentCacheEntry{
				Title:    "Some page",
				Category: "other",
				Tags:     []string{"видео", "topic"},
			},
			want: false,
		},
		{
			name: "missing tags",
			entry: domain.URLEnrichmentCacheEntry{
				Title:    "Some page",
				Category: "learning",
				Tags:     []string{"курс"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUsableCacheEntry(tt.entry); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyCacheToCreateInput(t *testing.T) {
	input := domain.CreateBookmarkInput{
		URL:      "https://www.youtube.com/watch?v=abc123",
		Category: "other",
		Tags:     []string{},
	}
	entry := domain.URLEnrichmentCacheEntry{
		Title:       "GOLANG ПОЛНЫЙ КУРС",
		Description: "Курс по Go для начинающих.",
		Category:    "learning",
		Tags:        []string{"курс", "golang"},
		ImageURL:    "https://img.example/thumb.jpg",
	}

	applyCacheToCreateInput(&input, entry)

	if input.Title != entry.Title {
		t.Fatalf("title: got %q, want %q", input.Title, entry.Title)
	}
	if input.Category != "learning" {
		t.Fatalf("category: got %q", input.Category)
	}
	if len(input.Tags) != 2 {
		t.Fatalf("tags: got %v", input.Tags)
	}
	if input.ImageURL != entry.ImageURL {
		t.Fatalf("image: got %q", input.ImageURL)
	}
}

func TestApplyCacheToCreateInputDoesNotOverwriteExisting(t *testing.T) {
	input := domain.CreateBookmarkInput{
		URL:         "https://example.com",
		Title:       "Custom title",
		Description: "Custom description",
		Category:    "news",
		Tags:        []string{"новость", "tech"},
		ImageURL:    "https://example.com/img.jpg",
	}
	entry := domain.URLEnrichmentCacheEntry{
		Title:       "Cached title",
		Description: "Cached description",
		Category:    "learning",
		Tags:        []string{"курс", "golang"},
		ImageURL:    "https://cached/img.jpg",
	}

	applyCacheToCreateInput(&input, entry)

	if input.Title != "Custom title" {
		t.Fatalf("title overwritten: %q", input.Title)
	}
	if input.Category != "news" {
		t.Fatalf("category overwritten: %q", input.Category)
	}
}
