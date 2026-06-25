package pagemeta

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(html []byte, pageURL *url.URL) Page {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return Page{}
	}

	jsonLD := jsonLDPage(doc)

	title := strings.TrimSpace(firstNonEmpty(
		metaContent(doc, "property", "og:title"),
		metaContent(doc, "name", "twitter:title"),
		jsonLD.Title,
		doc.Find("title").First().Text(),
		doc.Find("h1").First().Text(),
	))

	description := strings.TrimSpace(firstNonEmpty(
		metaContent(doc, "property", "og:description"),
		metaContent(doc, "name", "twitter:description"),
		metaContent(doc, "name", "description"),
		jsonLD.Description,
	))

	imageURL := strings.TrimSpace(firstNonEmpty(
		metaContent(doc, "property", "og:image"),
		metaContent(doc, "property", "og:image:url"),
		metaContent(doc, "name", "twitter:image"),
		metaContent(doc, "name", "twitter:image:src"),
		jsonLD.ImageURL,
	))

	return Page{
		Title:       title,
		Description: description,
		ImageURL:    resolveURL(pageURL, imageURL),
	}
}

func jsonLDPage(doc *goquery.Document) Page {
	var page Page

	doc.Find("script[type='application/ld+json']").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		candidate := parseJSONLD(selection.Text())
		page = mergePage(page, candidate)
		return page.Title == "" || page.Description == "" || page.ImageURL == ""
	})

	return page
}

func parseJSONLD(raw string) Page {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Page{}
	}

	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return Page{}
	}

	return pageFromJSONLD(value)
}

func pageFromJSONLD(value any) Page {
	switch typed := value.(type) {
	case []any:
		var page Page
		for _, item := range typed {
			page = mergePage(page, pageFromJSONLD(item))
		}
		return page
	case map[string]any:
		if graph, ok := typed["@graph"]; ok {
			if page := pageFromJSONLD(graph); page.Title != "" || page.Description != "" || page.ImageURL != "" {
				return mergePage(pageFromJSONLDMap(typed), page)
			}
		}
		return pageFromJSONLDMap(typed)
	default:
		return Page{}
	}
}

func pageFromJSONLDMap(object map[string]any) Page {
	return Page{
		Title:       firstString(object, "headline", "name"),
		Description: firstString(object, "description"),
		ImageURL:    imageFromJSONLD(object["image"]),
	}
}

func firstString(object map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringFromJSONLD(object[key]); value != "" {
			return value
		}
	}
	return ""
}

func stringFromJSONLD(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		for _, item := range typed {
			if value := stringFromJSONLD(item); value != "" {
				return value
			}
		}
	case map[string]any:
		return firstString(typed, "url", "contentUrl", "@id", "name")
	}
	return ""
}

func imageFromJSONLD(value any) string {
	return stringFromJSONLD(value)
}

func mergePage(base, patch Page) Page {
	if base.Title == "" {
		base.Title = patch.Title
	}
	if base.Description == "" {
		base.Description = patch.Description
	}
	if base.ImageURL == "" {
		base.ImageURL = patch.ImageURL
	}
	return base
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
