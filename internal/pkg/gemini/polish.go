package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"pet-link/internal/domain"
)

func (e *Enricher) Polish(ctx context.Context, pageURL string, enrichment domain.BookmarkEnrichment) (domain.BookmarkEnrichment, error) {
	enrichment = NormalizeEnrichment(enrichment)
	if strings.TrimSpace(enrichment.Title) == "" && strings.TrimSpace(enrichment.Description) == "" {
		return domain.BookmarkEnrichment{}, fmt.Errorf("no content to polish")
	}

	prompt := polishPrompt(
		pageURL,
		enrichment.Title,
		enrichment.Description,
		enrichment.Category,
		strings.Join(enrichment.Tags, ", "),
	)
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.1)),
	}

	return e.generateEnrichment(ctx, "polish", pageURL, prompt, config)
}
