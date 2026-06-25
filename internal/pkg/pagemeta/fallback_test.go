package pagemeta

import "testing"

func TestTitleFromHDRezkaSlug(t *testing.T) {
	got := titleFromHDRezkaSlug("82305-mandalorec-i-grogu-2026-latest.html")
	want := "Mandalorec и Grogu"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEnrichmentFromHDRezkaURL(t *testing.T) {
	rawURL := "http://hdrezka.co/films/fiction/82305-mandalorec-i-grogu-2026-latest.html"
	got, ok := enrichmentFromKnownURL(rawURL)
	if !ok {
		t.Fatal("expected enrichment from hdrezka url")
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
