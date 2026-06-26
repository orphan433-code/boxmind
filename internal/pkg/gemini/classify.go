package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"pet-link/internal/domain"
)

func (e *Enricher) Classify(ctx context.Context, pageURL, title, description, titleSource string) (domain.BookmarkEnrichment, error) {
	title = cleanHint(title)
	description = cleanHint(description)
	titleSource = cleanHint(titleSource)
	if title == "" && description == "" {
		return domain.BookmarkEnrichment{}, fmt.Errorf("no hints to classify")
	}
	if titleSource == "" {
		titleSource = "unknown"
	}

	prompt := classifyPrompt(pageURL, title, description, titleSource)
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.1)),
	}

	return e.generateEnrichment(ctx, "classify", pageURL, prompt, config)
}

func cleanHint(value string) string {
	return strings.TrimSpace(value)
}
