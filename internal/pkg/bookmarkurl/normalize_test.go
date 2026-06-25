package bookmarkurl

import "testing"

func TestNormalize(t *testing.T) {
	got, err := Normalize("https://Example.com/path/#section")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "https://example.com/path"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	got, err = Normalize("https://example.com/path/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("trailing slash: got %q, want %q", got, want)
	}
}
