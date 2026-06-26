package pagemeta

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type oEmbedResponse struct {
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// youtubeOEmbed fetches the real video title and thumbnail via YouTube's public
// oEmbed endpoint. It is far more reliable than scraping the HTML page, which
// often returns a "- YouTube" placeholder title from datacenter IPs.
func youtubeOEmbed(ctx context.Context, client *http.Client, rawURL string) (Page, bool) {
	if youtubeVideoID(rawURL) == "" {
		return Page{}, false
	}

	endpoint := "https://www.youtube.com/oembed?format=json&url=" + url.QueryEscape(rawURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Page{}, false
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Page{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Page{}, false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return Page{}, false
	}

	var parsed oEmbedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Page{}, false
	}

	title := strings.TrimSpace(parsed.Title)
	if title == "" {
		return Page{}, false
	}

	// Intentionally no description: the generic "video by <author>" line carries no
	// meaning and pollutes classification. A real Russian summary is produced later by AI.
	page := Page{
		Title:    title,
		ImageURL: strings.TrimSpace(parsed.ThumbnailURL),
	}

	return page, true
}
