package pagemeta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	oEmbedMaxAttempts    = 3
	oEmbedRetryBaseDelay = 400 * time.Millisecond
)

type oEmbedResponse struct {
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type oEmbedError struct {
	status int
	cause  error
}

func (e oEmbedError) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return fmt.Sprintf("oembed status %d", e.status)
}

func (e oEmbedError) Unwrap() error {
	return e.cause
}

// youtubeOEmbed fetches the real video title and thumbnail via YouTube's public
// oEmbed endpoint. It is far more reliable than scraping the HTML page, which
// often returns a "- YouTube" placeholder title from datacenter IPs.
func youtubeOEmbed(ctx context.Context, client *http.Client, rawURL string) (Page, bool) {
	if youtubeVideoID(rawURL) == "" {
		return Page{}, false
	}

	endpoint := "https://www.youtube.com/oembed?format=json&url=" + url.QueryEscape(rawURL)
	return fetchOEmbedWithRetry(ctx, client, endpoint)
}

// tiktokOEmbed fetches caption/title, author and thumbnail via TikTok's public
// oEmbed endpoint. Short links (vt.tiktok.com, vm.tiktok.com) are supported.
func tiktokOEmbed(ctx context.Context, client *http.Client, rawURL string) (Page, bool) {
	host, ok := normalizedHost(rawURL)
	if !ok || !isTikTokHost(host) {
		return Page{}, false
	}

	endpoint := "https://www.tiktok.com/oembed?url=" + url.QueryEscape(rawURL)
	return fetchOEmbedWithRetry(ctx, client, endpoint)
}

func fetchOEmbedWithRetry(ctx context.Context, client *http.Client, endpoint string) (Page, bool) {
	for attempt := 1; attempt <= oEmbedMaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return Page{}, false
		}

		page, err := oEmbedOnce(ctx, client, endpoint)
		if err == nil {
			return page, true
		}
		if !isRetryableOEmbedError(err) || attempt == oEmbedMaxAttempts {
			return Page{}, false
		}

		delay := oEmbedRetryBaseDelay * time.Duration(attempt)
		select {
		case <-ctx.Done():
			return Page{}, false
		case <-time.After(delay):
		}
	}

	return Page{}, false
}

func oEmbedOnce(ctx context.Context, client *http.Client, endpoint string) (Page, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Page{}, oEmbedError{cause: err}
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Page{}, oEmbedError{cause: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Page{}, oEmbedError{status: resp.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return Page{}, oEmbedError{cause: err}
	}

	var parsed oEmbedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Page{}, oEmbedError{cause: err}
	}

	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		return Page{}, oEmbedError{status: http.StatusNoContent}
	}

	// Intentionally no description: generic "video by <author>" lines carry no
	// meaning and pollute classification. A real summary is produced later by AI.
	return Page{
		Title:    title,
		ImageURL: strings.TrimSpace(parsed.ThumbnailURL),
	}, nil
}

func isRetryableOEmbedError(err error) bool {
	var oe oEmbedError
	if !errors.As(err, &oe) {
		return false
	}
	if oe.cause != nil {
		return true
	}
	switch oe.status {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return oe.status >= 500
	}
}
