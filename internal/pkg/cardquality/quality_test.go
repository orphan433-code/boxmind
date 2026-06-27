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

func TestNeedsPolishForSEOTitleAndWeakDescription(t *testing.T) {
	e := domain.BookmarkEnrichment{
		Title:       "Ноты для фортепиано, голоса, легкие ноты - скачать бесплатно",
		Description: "Ноты для фортепиано скачать бесплатно, с видео разборами для начинающих и профессионалов.",
		Category:    "learning",
		Tags:        []string{"ноты", "фортепиано"},
	}

	if GoodTitle(e.Title) {
		t.Fatal("expected SEO title to be bad")
	}
	if GoodDescription(e.Description) {
		t.Fatal("expected weak SEO description to be bad")
	}
	if !NeedsPolish(e, "") {
		t.Fatal("expected card to need polish")
	}
	if IsAcceptable(e, "") {
		t.Fatal("expected weak text card to be unacceptable")
	}
}

func TestNeedsPolishKeepsCleanCard(t *testing.T) {
	e := domain.BookmarkEnrichment{
		Title:       "Как выучить ноты на грифе гитары",
		Description: "Краткое объяснение расположения нот на гитарном грифе.",
		Category:    "learning",
		Tags:        []string{"ноты", "гитара"},
	}

	if NeedsPolish(e, "") {
		t.Fatal("expected clean card to skip polish")
	}
}

func TestNeedsPolishForShoppingCard(t *testing.T) {
	e := domain.BookmarkEnrichment{
		Title:       "Фильтр AQUASOFT Classic-5",
		Description: "Купить AQUASOFT Classic-5 за 67990 тг.",
		Category:    "shopping",
		Tags:        []string{"товар", "бытовая-техника"},
	}

	if !NeedsPolish(e, "") {
		t.Fatal("expected shopping card to need polish")
	}
	if BadTitle(e.Title) {
		t.Fatal("expected product title to stay valid")
	}
}

func TestNeedsPolishForJobsCard(t *testing.T) {
	e := domain.BookmarkEnrichment{
		Title:       "Backend Go Developer — вакансия на hh.ru",
		Description: "Компания ищет разработчика на Go.",
		Category:    "jobs",
		Tags:        []string{"вакансия", "backend"},
	}

	if !NeedsPolish(e, "") {
		t.Fatal("expected jobs card to need polish")
	}
}
