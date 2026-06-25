package pagemeta

import (
	"net/url"
	"testing"
)

func TestParseHTML(t *testing.T) {
	html := []byte(`<!doctype html>
<html>
<head>
  <title>Fallback title</title>
  <meta property="og:title" content="OG Title">
  <meta property="og:description" content="OG Description">
  <meta property="og:image" content="/poster.jpg">
</head>
<body><h1>H1 Title</h1></body>
</html>`)

	base, err := url.Parse("https://example.com/films/1")
	if err != nil {
		t.Fatal(err)
	}

	page := parseHTML(html, base)

	if page.Title != "OG Title" {
		t.Fatalf("expected OG title, got %q", page.Title)
	}
	if page.Description != "OG Description" {
		t.Fatalf("expected OG description, got %q", page.Description)
	}
	if page.ImageURL != "https://example.com/poster.jpg" {
		t.Fatalf("expected resolved image url, got %q", page.ImageURL)
	}
}

func TestRejectPrivateHost(t *testing.T) {
	cases := []string{"localhost", "127.0.0.1", "10.0.0.1", "192.168.0.5"}
	for _, host := range cases {
		if err := rejectPrivateHost(host); err == nil {
			t.Fatalf("expected private host %q to be rejected", host)
		}
	}
}
