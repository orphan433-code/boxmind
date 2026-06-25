package pagemeta

import "testing"

func TestTitleFromSlug(t *testing.T) {
	got := titleFromSlug("82305-mandalorec-i-grogu-2026-latest.html")
	want := "Mandalorec и Grogu"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGenericURLHintsHDRezkaCartoon(t *testing.T) {
	rawURL := "http://hdrezka.co/cartoons/fantasy/90118-klubnichnoe-pirozhnoe-berri-v-bolshom-gorode-2021.html"
	got, ok := GenericURLHints(rawURL)
	if !ok {
		t.Fatal("expected hints from url slug")
	}
	if got.Title == "" {
		t.Fatal("expected non-empty title")
	}
	if got.Category != "movies" {
		t.Fatalf("category: got %q, want movies", got.Category)
	}
	if got.Tags[0] != "мультфильм" || got.Tags[1] != "фэнтези" {
		t.Fatalf("tags: got %v, want [мультфильм фэнтези]", got.Tags)
	}
}

func TestGenericURLHintsFilm(t *testing.T) {
	rawURL := "http://hdrezka.co/films/fiction/82305-mandalorec-i-grogu-2026-latest.html"
	got, ok := GenericURLHints(rawURL)
	if !ok {
		t.Fatal("expected enrichment from url")
	}
	if got.Category != "movies" {
		t.Fatalf("category: got %q, want movies", got.Category)
	}
	if got.Tags[0] != "фильм" || got.Tags[1] != "фантастика" {
		t.Fatalf("tags: got %v, want [фильм фантастика]", got.Tags)
	}
	if got.Title == "" {
		t.Fatal("expected non-empty title")
	}
}

func TestGenericURLHintsPlainBlog(t *testing.T) {
	got, ok := GenericURLHints("https://example.com/blog/2024/my-cool-article-title")
	if !ok {
		t.Fatal("expected title from generic slug")
	}
	if got.Title != "My Cool Article Title" {
		t.Fatalf("title: got %q", got.Title)
	}
	if got.Category != "other" {
		t.Fatalf("category: got %q, want other", got.Category)
	}
	if len(got.Tags) != 0 {
		t.Fatalf("tags: got %v, want none", got.Tags)
	}
}
