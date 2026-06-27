package service

import (
	"testing"

	"pet-link/internal/domain"
)

func TestMergePolishedEnrichmentFixesWeakTextOnly(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Ноты для фортепиано, голоса, легкие ноты - скачать бесплатно",
		Description: "Ноты для фортепиано скачать бесплатно, с видео разборами для начинающих и профессионалов.",
		Category:    "learning",
		Tags:        []string{"ноты", "фортепиано"},
	}
	polished := domain.BookmarkEnrichment{
		Title:       "Ноты для фортепиано",
		Description: "Подборка нот для практики и обучения игре.",
		Category:    "articles",
		Tags:        []string{"музыка", "обучение"},
	}

	got := mergePolishedEnrichment(base, polished)
	if got.Title != polished.Title {
		t.Fatalf("title = %q, want %q", got.Title, polished.Title)
	}
	if got.Description != polished.Description {
		t.Fatalf("description = %q, want %q", got.Description, polished.Description)
	}
	if got.Category != base.Category {
		t.Fatalf("category = %q, want %q", got.Category, base.Category)
	}
	if got.Tags[0] != base.Tags[0] || got.Tags[1] != base.Tags[1] {
		t.Fatalf("tags = %#v, want %#v", got.Tags, base.Tags)
	}
}

func TestMergePolishedEnrichmentKeepsCleanText(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Как выучить ноты на грифе гитары",
		Description: "Объяснение расположения нот на гитарном грифе.",
		Category:    "learning",
		Tags:        []string{"ноты", "гитара"},
	}
	polished := domain.BookmarkEnrichment{
		Title:       "Ноты на гитаре",
		Description: "Материал по нотам для гитары.",
		Category:    "learning",
		Tags:        []string{"музыка", "гитара"},
	}

	got := mergePolishedEnrichment(base, polished)
	if got.Title != base.Title {
		t.Fatalf("title changed: got %q", got.Title)
	}
	if got.Description != base.Description {
		t.Fatalf("description changed: got %q", got.Description)
	}
}

func TestMergePolishedEnrichmentRefreshesShoppingDescription(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Фильтр AQUASOFT Classic-5",
		Description: "Купить AQUASOFT Classic-5 за 67990 тг.",
		Category:    "shopping",
		Tags:        []string{"товар", "бытовая-техника"},
	}
	polished := domain.BookmarkEnrichment{
		Title:       "AQUASOFT Classic-5",
		Description: "Фильтр для очистки питьевой воды дома.",
		Category:    "shopping",
		Tags:        []string{"товар", "фильтр"},
	}

	got := mergePolishedEnrichment(base, polished)
	if got.Title != base.Title {
		t.Fatalf("title changed: got %q", got.Title)
	}
	if got.Description != polished.Description {
		t.Fatalf("description = %q, want %q", got.Description, polished.Description)
	}
	if got.Category != base.Category {
		t.Fatalf("category = %q, want %q", got.Category, base.Category)
	}
	if got.Tags[0] != base.Tags[0] || got.Tags[1] != base.Tags[1] {
		t.Fatalf("tags = %#v, want %#v", got.Tags, base.Tags)
	}
}

func TestMergePolishedEnrichmentRefreshesJobsCard(t *testing.T) {
	base := domain.BookmarkEnrichment{
		Title:       "Backend Go Developer — вакансия на hh.ru",
		Description: "Компания ищет разработчика на Go.",
		Category:    "jobs",
		Tags:        []string{"вакансия", "backend"},
	}
	polished := domain.BookmarkEnrichment{
		Title:       "Backend Go Developer",
		Description: "Backend на Go: удалённо, middle+, до 350 000 ₽.",
		Category:    "jobs",
		Tags:        []string{"вакансия", "golang"},
	}

	got := mergePolishedEnrichment(base, polished)
	if got.Title != polished.Title {
		t.Fatalf("title = %q, want %q", got.Title, polished.Title)
	}
	if got.Description != polished.Description {
		t.Fatalf("description = %q, want %q", got.Description, polished.Description)
	}
	if got.Category != base.Category {
		t.Fatalf("category = %q, want %q", got.Category, base.Category)
	}
	if got.Tags[0] != base.Tags[0] || got.Tags[1] != base.Tags[1] {
		t.Fatalf("tags = %#v, want %#v", got.Tags, base.Tags)
	}
}
