package service

import (
	"context"
	"testing"

	"pet-link/internal/domain"
)

func TestEligibleForClassifyFirst(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		enrichment domain.BookmarkEnrichment
		want       bool
	}{
		{
			name: "youtube metadata title",
			url:  "https://www.youtube.com/watch?v=abc123",
			enrichment: domain.BookmarkEnrichment{
				Title: "GOLANG ПОЛНЫЙ КУРС ДЛЯ НАЧИНАЮЩИХ",
			},
			want: true,
		},
		{
			name: "hdrezka slug title",
			url:  "http://hdrezka.co/animation/fantasy/90535-geroinya-svyataya-net-ya-vsemoguschaya-gornichnaya-2026.html",
			enrichment: domain.BookmarkEnrichment{
				Title: "Geroinya Svyataya Net Ya Vsemoguschaya Gornichnaya",
			},
			want: true,
		},
		{
			name: "raw url title",
			url:  "https://example.com/article",
			enrichment: domain.BookmarkEnrichment{
				Title: "https://example.com/article",
			},
			want: false,
		},
		{
			name: "empty title",
			url:  "https://example.com/article",
			enrichment: domain.BookmarkEnrichment{
				Description: "Only description",
			},
			want: false,
		},
		{
			name: "youtube placeholder title",
			url:  "https://www.youtube.com/watch?v=abc123",
			enrichment: domain.BookmarkEnrichment{
				Title: "- YouTube",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := eligibleForClassifyFirst(tt.url, tt.enrichment); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

type classifyFirstEnricher struct {
	classifyCalls int
	enrichCalls   int
	classified    domain.BookmarkEnrichment
}

func (m *classifyFirstEnricher) Enrich(_ context.Context, _ string) (domain.BookmarkEnrichment, error) {
	m.enrichCalls++
	return domain.BookmarkEnrichment{}, nil
}

func (m *classifyFirstEnricher) Classify(_ context.Context, _, _, _, _ string) (domain.BookmarkEnrichment, error) {
	m.classifyCalls++
	return m.classified, nil
}

type classifyFirstRepo struct {
	bookmark domain.Bookmark
}

func (r *classifyFirstRepo) Create(_ context.Context, _ string, input domain.CreateBookmarkInput) (domain.Bookmark, error) {
	r.bookmark.Title = input.Title
	r.bookmark.Description = input.Description
	r.bookmark.Category = input.Category
	r.bookmark.Tags = input.Tags
	r.bookmark.ImageURL = input.ImageURL
	return r.bookmark, nil
}

func (r *classifyFirstRepo) ExistsByURLForUser(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (r *classifyFirstRepo) ListByUserID(_ context.Context, _ string) ([]domain.Bookmark, error) {
	return nil, nil
}

func (r *classifyFirstRepo) GetByIDForUser(_ context.Context, _, _ string) (domain.Bookmark, error) {
	return r.bookmark, nil
}

func (r *classifyFirstRepo) UpdateImageURL(_ context.Context, _, _, imageURL string) error {
	r.bookmark.ImageURL = imageURL
	return nil
}

func (r *classifyFirstRepo) UpdateEnrichment(_ context.Context, _, _ string, enrichment domain.BookmarkEnrichment) error {
	if enrichment.Title != "" {
		r.bookmark.Title = enrichment.Title
	}
	if enrichment.Description != "" {
		r.bookmark.Description = enrichment.Description
	}
	if enrichment.Category != "" {
		r.bookmark.Category = enrichment.Category
	}
	if len(enrichment.Tags) > 0 {
		r.bookmark.Tags = enrichment.Tags
	}
	return nil
}

func (r *classifyFirstRepo) MarkEnriched(_ context.Context, _, _ string) error {
	r.bookmark.Enriched = true
	return nil
}

func (r *classifyFirstRepo) Delete(_ context.Context, _, _ string) error {
	return nil
}

func TestRunClassifyFirstSkipsEnrichWhenAcceptable(t *testing.T) {
	repo := &classifyFirstRepo{
		bookmark: domain.Bookmark{
			ID:       "b1",
			UserID:   "u1",
			URL:      "https://www.youtube.com/watch?v=abc123",
			Title:    "GOLANG ПОЛНЫЙ КУРС ДЛЯ НАЧИНАЮЩИХ",
			ImageURL: "https://img.example/thumb.jpg",
			Category: "other",
			Tags:     []string{},
		},
	}
	enricher := &classifyFirstEnricher{
		classified: domain.BookmarkEnrichment{
			Title:       "GOLANG ПОЛНЫЙ КУРС ДЛЯ НАЧИНАЮЩИХ",
			Description: "Курс по Go для начинающих.",
			Category:    "learning",
			Tags:        []string{"курс", "golang"},
		},
	}

	svc := NewBookmarkServiceWithCache(repo, nil, enricher, nil, nil).(*bookmarkService)

	_, done := svc.runClassifyFirst(
		t.Context(),
		"u1",
		"b1",
		repo.bookmark.URL,
		enrichmentFromBookmark(repo.bookmark),
	)
	if !done {
		t.Fatal("expected classify-first to finish enrichment")
	}
	if enricher.classifyCalls != 1 {
		t.Fatalf("classify calls: got %d, want 1", enricher.classifyCalls)
	}
	if enricher.enrichCalls != 0 {
		t.Fatalf("enrich calls: got %d, want 0", enricher.enrichCalls)
	}
	if repo.bookmark.Category != "learning" {
		t.Fatalf("category: got %q", repo.bookmark.Category)
	}
}

func TestRunClassifyFirstFallsBackWhenInsufficient(t *testing.T) {
	repo := &classifyFirstRepo{
		bookmark: domain.Bookmark{
			ID:     "b1",
			UserID: "u1",
			URL:    "https://www.youtube.com/watch?v=abc123",
			Title:  "GOLANG ПОЛНЫЙ КУРС ДЛЯ НАЧИНАЮЩИХ",
		},
	}
	enricher := &classifyFirstEnricher{
		classified: domain.BookmarkEnrichment{
			Category: "other",
			Tags:     []string{},
		},
	}

	svc := NewBookmarkServiceWithCache(repo, nil, enricher, nil, nil).(*bookmarkService)

	_, done := svc.runClassifyFirst(
		t.Context(),
		repo.bookmark.UserID,
		repo.bookmark.ID,
		repo.bookmark.URL,
		enrichmentFromBookmark(repo.bookmark),
	)
	if done {
		t.Fatal("expected classify-first to fall back to enrich")
	}
	if enricher.classifyCalls != 1 {
		t.Fatalf("classify calls: got %d, want 1", enricher.classifyCalls)
	}
}
