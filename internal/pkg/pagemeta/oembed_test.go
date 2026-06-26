package pagemeta

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchOEmbedWithRetry(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch calls.Add(1) {
		case 1:
			http.Error(w, "timeout", http.StatusGatewayTimeout)
		case 2:
			http.Error(w, "busy", http.StatusServiceUnavailable)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":"Оригинальный title","thumbnail_url":"https://example.com/thumb.jpg"}`))
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	page, ok := fetchOEmbedWithRetry(ctx, server.Client(), server.URL)
	if !ok {
		t.Fatal("expected success after retries")
	}
	if page.Title != "Оригинальный title" {
		t.Fatalf("unexpected title: %q", page.Title)
	}
	if page.ImageURL != "https://example.com/thumb.jpg" {
		t.Fatalf("unexpected thumbnail: %q", page.ImageURL)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
}

func TestFetchOEmbedWithRetryDoesNotRetry404(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()

	page, ok := fetchOEmbedWithRetry(context.Background(), server.Client(), server.URL)
	if ok {
		t.Fatalf("expected failure, got page %+v", page)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected single call, got %d", calls.Load())
	}
}

func TestTikTokOEmbedIgnoresNonTikTok(t *testing.T) {
	page, ok := tiktokOEmbed(context.Background(), http.DefaultClient, "https://example.com/video")
	if ok {
		t.Fatalf("expected false for non-tiktok url, got %+v", page)
	}
}

func TestTikTokOEmbedParsesCaption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"title":"Как настроить Xcode за 30 секунд",
			"author_name":"devtips",
			"thumbnail_url":"https://p16-sign.tiktokcdn.com/thumb.jpg"
		}`))
	}))
	defer server.Close()

	page, err := oEmbedOnce(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Title != "Как настроить Xcode за 30 секунд" {
		t.Fatalf("unexpected title: %q", page.Title)
	}
	if page.ImageURL != "https://p16-sign.tiktokcdn.com/thumb.jpg" {
		t.Fatalf("unexpected thumbnail: %q", page.ImageURL)
	}
}

func TestIsRetryableOEmbedError(t *testing.T) {
	if !isRetryableOEmbedError(oEmbedError{status: http.StatusBadGateway}) {
		t.Fatal("502 should retry")
	}
	if isRetryableOEmbedError(oEmbedError{status: http.StatusNotFound}) {
		t.Fatal("404 should not retry")
	}
	if !isRetryableOEmbedError(oEmbedError{cause: context.DeadlineExceeded}) {
		t.Fatal("network errors should retry")
	}
}
