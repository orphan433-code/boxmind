package pagemeta

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(html []byte, pageURL *url.URL) Page {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return Page{}
	}

	title := strings.TrimSpace(firstNonEmpty(
		metaContent(doc, "property", "og:title"),
		metaContent(doc, "name", "twitter:title"),
		doc.Find("title").First().Text(),
		doc.Find("h1").First().Text(),
	))

	description := strings.TrimSpace(firstNonEmpty(
		metaContent(doc, "property", "og:description"),
		metaContent(doc, "name", "twitter:description"),
		metaContent(doc, "name", "description"),
	))

	imageURL := strings.TrimSpace(firstNonEmpty(
		metaContent(doc, "property", "og:image"),
		metaContent(doc, "property", "og:image:url"),
		metaContent(doc, "name", "twitter:image"),
		metaContent(doc, "name", "twitter:image:src"),
	))

	return Page{
		Title:       title,
		Description: description,
		ImageURL:    resolveURL(pageURL, imageURL),
	}
}

func metaContent(doc *goquery.Document, attr, value string) string {
	selection := doc.Find(fmtMetaSelector(attr, value)).First()
	if content, ok := selection.Attr("content"); ok {
		return strings.TrimSpace(content)
	}
	return ""
}

func fmtMetaSelector(attr, value string) string {
	return "meta[" + attr + "='" + value + "']"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func resolveURL(base *url.URL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || base == nil {
		return ""
	}

	ref, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	return base.ResolveReference(ref).String()
}
