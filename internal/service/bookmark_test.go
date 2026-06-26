package service

import "testing"

func TestTitleSourceForClassification(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		title string
		want  string
	}{
		{
			name:  "slug title",
			url:   "http://hdrezka.co/films/fantasy/123-garri-potter-i-uznik-azkabana-2004.html",
			title: "Garri Potter и Uznik Azkabana",
			want:  "url_slug",
		},
		{
			name:  "trusted metadata",
			url:   "https://www.youtube.com/watch?v=abc123",
			title: "GOLANG ПОЛНЫЙ КУРС ДЛЯ НАЧИНАЮЩИХ",
			want:  "metadata_or_user",
		},
		{
			name:  "raw url title",
			url:   "https://example.com",
			title: "https://example.com",
			want:  "url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := titleSourceForClassification(tt.url, tt.title); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
