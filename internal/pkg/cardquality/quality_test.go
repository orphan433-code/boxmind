package cardquality

import (
	"testing"

	"pet-link/internal/domain"
)

func TestMergeKeepsGoodDescriptionOverJoke(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Ходячие мертвецы",
		Description: "Выжившие борются с зомби в постапокалипсисе.",
		Category:    "movies",
		Tags:        []string{"сериал", "ужасы"},
	}
	patch := domain.BookmarkEnrichment{
		Description: "Страница стесняется, содержимое не показала.",
		Category:    "other",
		Tags:        []string{"ссылка", "недоступно"},
	}

	got := Merge(base, patch)
	if got.Description != base.Description {
		t.Fatalf("description = %q, want %q", got.Description, base.Description)
	}
	if got.Category != "movies" {
		t.Fatalf("category = %q, want movies", got.Category)
	}
}

func TestMergePrefersCleanTitle(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title: "Ледяная стена — смотреть аниме онлайн",
	}
	patch := domain.BookmarkEnrichment{
		Title: "Ледяная стена",
	}

	got := Merge(base, patch)
	if got.Title != "Ледяная стена" {
		t.Fatalf("title = %q", got.Title)
	}
}

func TestIsGoodEnough(t *testing.T) {
	e := domain.BookmarkEnrichment{
		Title:       "Ледяная стена",
		Description: "Аниме о девушке, которая учится общаться с людьми.",
		Category:    "movies",
		Tags:        []string{"аниме", "комедия"},
	}
	if !IsGoodEnough(e, "https://img.example/thumb.jpg") {
		t.Fatal("expected good enough card")
	}
}
