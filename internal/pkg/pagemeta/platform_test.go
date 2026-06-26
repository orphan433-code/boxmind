package pagemeta

import "testing"

func TestPlatformThumbnailURLYouTube(t *testing.T) {
	got := PlatformThumbnailURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	want := "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPlatformThumbnailURLYouTubeRegional(t *testing.T) {
	got := PlatformThumbnailURL("https://youtube.kz/watch?v=dQw4w9WgXcQ")
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

func TestPlatformTag(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://vt.tiktok.com/ZSC6hT7J2/", "TikTok"},
		{"https://vm.tiktok.com/abc123/", "TikTok"},
		{"https://www.tiktok.com/@user/video/123", "TikTok"},
		{"https://www.instagram.com/reel/DaBYoNgNmI_/", "Instagram"},
		{"https://instagr.am/reel/abc/", "Instagram"},
		{"https://instagram.co.uk/reel/abc/", "Instagram"},
		{"https://www.youtube.com/watch?v=abc", "YouTube"},
		{"https://youtube.kz/watch?v=abc", "YouTube"},
		{"https://www.youtube.co.uk/watch?v=abc", "YouTube"},
		{"https://youtu.be/abc", "YouTube"},
		{"https://example.com/page", ""},
	}

	for _, tc := range tests {
		got, ok := PlatformTag(tc.url)
		if tc.want == "" {
			if ok {
				t.Fatalf("%s: expected no tag, got %q", tc.url, got)
			}
			continue
		}
		if !ok || got != tc.want {
			t.Fatalf("%s: got %q ok=%v, want %q", tc.url, got, ok, tc.want)
		}
	}
}

func TestEnsurePlatformTag(t *testing.T) {
	got := EnsurePlatformTag("https://vt.tiktok.com/ZSC6hT7J2/", []string{"видео", "комедия"})
	if len(got) != 3 || got[0] != "TikTok" {
		t.Fatalf("unexpected tags: %v", got)
	}

	dup := EnsurePlatformTag("https://vt.tiktok.com/ZSC6hT7J2/", []string{"TikTok", "видео"})
	if len(dup) != 2 {
		t.Fatalf("expected no duplicate platform tag: %v", dup)
	}
}
