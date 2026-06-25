package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"pet-link/internal/domain"
)

func (e *Enricher) Classify(ctx context.Context, pageURL, title, description string) (domain.BookmarkEnrichment, error) {
	title = cleanHint(title)
	description = cleanHint(description)
	if title == "" && description == "" {
		return domain.BookmarkEnrichment{}, fmt.Errorf("no hints to classify")
	}

	prompt := classifyPrompt(pageURL, title, description)
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.1)),
	}

	return e.generateEnrichment(ctx, prompt, config)
}

func cleanHint(value string) string {
	return strings.TrimSpace(value)
}
