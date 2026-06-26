package pagemeta

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

const (
	defaultTimeout   = 12 * time.Second
	imageTimeout     = 15 * time.Second
	defaultMaxBytes  = 2 << 20 // 2 MB
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type Extractor interface {
	Extract(ctx context.Context, rawURL string) (Page, error)
}

type HTTPExtractor struct {
	client    *http.Client
	maxBytes  int64
	userAgent string
}

func NewHTTPExtractor() *HTTPExtractor {
	return newHTTPExtractor(defaultTimeout)
}

func NewImageHTTPExtractor() *HTTPExtractor {
	return newHTTPExtractor(imageTimeout)
}

func newHTTPExtractor(timeout time.Duration) *HTTPExtractor {
	jar, _ := cookiejar.New(nil)
	return &HTTPExtractor{
		client: &http.Client{
			Timeout: timeout,
			Jar:     jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return rejectPrivateHost(req.URL.Hostname())
			},
		},
		maxBytes:  defaultMaxBytes,
		userAgent: defaultUserAgent,
	}
}

func (e *HTTPExtractor) Extract(ctx context.Context, rawURL string) (Page, error) {
	parsedURL, err := validateTargetURL(rawURL)
	if err != nil {
		return Page{}, err
	}

	if page, ok := youtubeOEmbed(ctx, e.client, parsedURL.String()); ok {
		return page, nil
	}

	if page, ok := tiktokOEmbed(ctx, e.client, parsedURL.String()); ok {
		return page, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return Page{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", e.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := e.client.Do(req)
	if err != nil {
		return Page{}, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return Page{}, fmt.Errorf("fetch page: status %d", resp.StatusCode)
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if contentType != "" && !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		return Page{}, fmt.Errorf("unsupported content type")
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, e.maxBytes))
	if err != nil {
		return Page{}, fmt.Errorf("read page: %w", err)
	}

	body = decodeToUTF8(body, resp.Header.Get("Content-Type"))

	page := parseHTML(body, parsedURL)
	if page.Title == "" && page.Description == "" && page.ImageURL == "" {
		return Page{}, fmt.Errorf("page metadata not found")
	}

	return page, nil
}

// decodeToUTF8 converts a page body to UTF-8 using the charset declared in the
// Content-Type header, an HTML <meta> tag, or a BOM. Pages served in legacy
// encodings (e.g. windows-1251) would otherwise yield invalid UTF-8 bytes that
// break downstream UTF-8 storage.
func decodeToUTF8(body []byte, contentType string) []byte {
	reader, err := charset.NewReader(bytes.NewReader(body), contentType)
	if err != nil {
		return body
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return body
	}
	return decoded
}
