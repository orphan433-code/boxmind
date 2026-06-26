package pagemeta

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchYouTubeOEmbedWithRetry(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch calls.Add(1) {
		case 1:
			http.Error(w, "timeout", http.StatusGatewayTimeout)
		case 2:
			http.Error(w, "busy", http.StatusServiceUnavailable)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":"Оригинальный title","thumbnail_url":"https://i.ytimg.com/vi/abc/hqdefault.jpg"}`))
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	page, ok := fetchYouTubeOEmbedWithRetry(ctx, server.Client(), server.URL)
	if !ok {
		t.Fatal("expected success after retries")
	}
	if page.Title != "Оригинальный title" {
		t.Fatalf("unexpected title: %q", page.Title)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
}

func TestFetchYouTubeOEmbedWithRetryDoesNotRetry404(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()

	page, ok := fetchYouTubeOEmbedWithRetry(context.Background(), server.Client(), server.URL)
	if ok {
		t.Fatalf("expected failure, got page %+v", page)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected single call, got %d", calls.Load())
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
