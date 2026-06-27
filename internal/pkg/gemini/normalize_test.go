package gemini

import (
	"testing"

	"pet-link/internal/domain"
)

func TestNormalizeDescription(t *testing.T) {
	got := normalizeDescription("Боевик про героя. Здесь можно смотреть бесплатно.")
	want := "Боевик про героя."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTruncateAtWord(t *testing.T) {
	long := "Аниме-сериал о юноше без магии, который прокладывает себе путь к вершине волшебного мира с помощью меча и ума"
	got := truncateAtWord(long, 100)
	if len([]rune(got)) > 101 {
		t.Fatalf("too long: %d runes", len([]rune(got)))
	}
	if !stringsHasSuffix(got, "…") {
		t.Fatalf("expected ellipsis suffix, got %q", got)
	}
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func TestIsUnavailableEnrichment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input domain.BookmarkEnrichment
		want  bool
	}{
		{
			name: "fallback when page blocked",
			input: domain.BookmarkEnrichment{
				Category: "other",
				Tags:     []string{"ссылка", "недоступно"},
			},
			want: true,
		},
		{
			name: "joke description with real category",
			input: domain.BookmarkEnrichment{
				Title:       "Ходячие мертвецы",
				Description: "Страница стесняется, содержимое не показала.",
				Category:    "movies",
				Tags:        []string{"сериал", "ужасы"},
			},
			want: true,
		},
		{
			name: "real other bookmark",
			input: domain.BookmarkEnrichment{
				Category: "other",
				Tags:     []string{"ссылка", "личное"},
			},
			want: false,
		},
		{
			name: "successful enrich",
			input: domain.BookmarkEnrichment{
				Category: "learning",
				Tags:     []string{"курс", "golang"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsUnavailableEnrichment(tt.input); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyURLCategoryHints(t *testing.T) {
	t.Parallel()

	youtube := "https://www.youtube.com/watch?v=abc"
	input := domain.BookmarkEnrichment{
		Category: "news",
		Tags:     []string{"новость", "политика"},
	}
	got := applyURLCategoryHints(youtube, input)

	if got.Category != "entertainment" {
		t.Fatalf("category: got %q, want entertainment", got.Category)
	}
	if got.Tags[0] != "видео" || got.Tags[1] != "политика" {
		t.Fatalf("tags: got %v, want [видео политика]", got.Tags)
	}

	article := applyURLCategoryHints(
		"https://www.bbc.com/news/world",
		input,
	)
	if article.Category != "news" {
		t.Fatalf("article should stay news, got %q", article.Category)
	}

	profile := applyURLCategoryHints(
		"https://leetcode.com/u/eabramov1993/",
		domain.BookmarkEnrichment{Category: "other", Tags: []string{"ссылка", "недоступно"}},
	)
	if profile.Category != "programming" {
		t.Fatalf("profile category: got %q, want programming", profile.Category)
	}
	if len(profile.Tags) != 2 || profile.Tags[0] != "профиль" || profile.Tags[1] != "программирование" {
		t.Fatalf("profile tags: got %v, want [профиль программирование]", profile.Tags)
	}

	vacancy := applyURLCategoryHints(
		"https://hh.ru/vacancy/12345678",
		domain.BookmarkEnrichment{
			Category: "programming",
			Tags:     []string{"документация", "golang"},
		},
	)
	if vacancy.Category != "jobs" {
		t.Fatalf("vacancy category: got %q, want jobs", vacancy.Category)
	}
	if len(vacancy.Tags) != 2 || vacancy.Tags[0] != "вакансия" || vacancy.Tags[1] != "golang" {
		t.Fatalf("vacancy tags: got %v, want [вакансия golang]", vacancy.Tags)
	}
}

func TestNormalizeEnrichmentUsefulTags(t *testing.T) {
	t.Parallel()

	got := NormalizeEnrichment(domain.BookmarkEnrichment{
		Category: "tools",
		Tags:     []string{"docs", "cloud"},
	})
	if len(got.Tags) != 2 || got.Tags[0] != "документация" || got.Tags[1] != "облако" {
		t.Fatalf("got %v, want [документация облако]", got.Tags)
	}

	got = NormalizeEnrichment(domain.BookmarkEnrichment{
		Category: "programming",
		Tags:     []string{"random-tag", "weird-topic"},
	})
	if len(got.Tags) != 2 || got.Tags[0] != "инструмент" || got.Tags[1] != "программирование" {
		t.Fatalf("got %v, want [инструмент программирование]", got.Tags)
	}
}

func TestNormalizeEnrichmentJobsCategory(t *testing.T) {
	t.Parallel()

	got := NormalizeEnrichment(domain.BookmarkEnrichment{
		Category: "work",
		Tags:     []string{"backend", "golang"},
	})
	if got.Category != "jobs" {
		t.Fatalf("category = %q, want jobs", got.Category)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "вакансия" || got.Tags[1] != "backend" {
		t.Fatalf("tags = %v, want [вакансия backend]", got.Tags)
	}
}
