package pagemeta

import "testing"

func TestPlatformThumbnailURLYouTube(t *testing.T) {
	got := PlatformThumbnailURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	want := "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTitleHintFromURLAnimeSite(t *testing.T) {
	got, ok := TitleHintFromURL("https://animesss.com/aniserials/video/comedy/3263-ledjanaja-stena.html")
	if !ok {
		t.Fatal("expected title hint")
	}
	if got == "" {
		t.Fatal("empty title")
	}
}
