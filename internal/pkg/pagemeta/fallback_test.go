package pagemeta

import "testing"

func TestTitleFromSlug(t *testing.T) {
	got := titleFromSlug("82305-mandalorec-i-grogu-2026-latest.html")
	want := "Mandalorec и Grogu"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTitleHintFromURLHDRezkaCartoon(t *testing.T) {
	rawURL := "http://hdrezka.co/cartoons/fantasy/90118-klubnichnoe-pirozhnoe-berri-v-bolshom-gorode-2021.html"
	got, ok := TitleHintFromURL(rawURL)
	if !ok {
		t.Fatal("expected title hint from url slug")
	}
	if got == "" {
		t.Fatal("expected non-empty title")
	}
}

func TestTitleHintFromURLFilm(t *testing.T) {
	rawURL := "http://hdrezka.co/films/fiction/82305-mandalorec-i-grogu-2026-latest.html"
	got, ok := TitleHintFromURL(rawURL)
	if !ok {
		t.Fatal("expected title hint from url")
	}
	if got == "" {
		t.Fatal("expected non-empty title")
	}
}

func TestTitleHintFromURLPlainBlog(t *testing.T) {
	got, ok := TitleHintFromURL("https://example.com/blog/2024/my-cool-article-title")
	if !ok {
		t.Fatal("expected title from generic slug")
	}
	if got != "My Cool Article Title" {
		t.Fatalf("title: got %q", got)
	}
}

func TestFallbackEnrichmentTitleOnly(t *testing.T) {
	got, ok := titleHintEnrichment("http://hdrezka.co/series/drama/90507-vse-chto-ty-lyubish-2022.html")
	if !ok {
		t.Fatal("expected title hint")
	}
	if got.Title == "" {
		t.Fatal("expected title")
	}
	if got.Category != "" || len(got.Tags) > 0 {
		t.Fatalf("expected content-only hint, got category=%q tags=%v", got.Category, got.Tags)
	}
}
