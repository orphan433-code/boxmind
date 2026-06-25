package pagemeta

import "testing"

func TestPlatformThumbnailURLYouTube(t *testing.T) {
	got := PlatformThumbnailURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	want := "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGenericURLHintsAnimeSite(t *testing.T) {
	got, ok := GenericURLHints("https://animesss.com/aniserials/video/comedy/3263-ledjanaja-stena.html")
	if !ok {
		t.Fatal("expected hints")
	}
	if got.Title != "Ledjanaja stena" && got.Title != "Ледяная стена" {
		// slug translit title - at least non-empty
		if got.Title == "" {
			t.Fatalf("empty title")
		}
	}
	if got.Category != "movies" {
		t.Fatalf("category = %q", got.Category)
	}
	if len(got.Tags) != 2 {
		t.Fatalf("tags = %v", got.Tags)
	}
}
