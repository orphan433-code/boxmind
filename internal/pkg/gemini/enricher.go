package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/genai"

	"pet-link/internal/domain"
)

type Enricher struct {
	client *genai.Client
	model  string
}

func NewEnricher(ctx context.Context, apiKey, model string) (*Enricher, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	return &Enricher{
		client: client,
		model:  model,
	}, nil
}

const maxUnavailableAttempts = 3

func (e *Enricher) Enrich(ctx context.Context, pageURL string) (domain.BookmarkEnrichment, error) {
	prompt := enrichPrompt(pageURL)

	config := &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{URLContext: &genai.URLContext{}},
		},
		Temperature: genai.Ptr(float32(0.2)),
	}

	var last domain.BookmarkEnrichment

	for attempt := 1; attempt <= maxUnavailableAttempts; attempt++ {
		enrichment, err := e.generateEnrichment(ctx, "enrich", pageURL, prompt, config)
		if err != nil {
			return domain.BookmarkEnrichment{}, err
		}

		enrichment = applyURLCategoryHints(pageURL, enrichment)
		last = enrichment
		if !IsUnavailableEnrichment(enrichment) {
			return enrichment, nil
		}
		if attempt == maxUnavailableAttempts {
			break
		}

		backoff := time.Duration(attempt) * time.Second
		select {
		case <-ctx.Done():
			return domain.BookmarkEnrichment{}, ctx.Err()
		case <-time.After(backoff):
		}
	}

	return last, nil
}

func (e *Enricher) generateEnrichment(ctx context.Context, operation, pageURL, prompt string, config *genai.GenerateContentConfig) (domain.BookmarkEnrichment, error) {
	const maxAttempts = 3
	var resp *genai.GenerateContentResponse
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err = e.client.Models.GenerateContent(ctx, e.model, genai.Text(prompt), config)
		if err == nil {
			logTokenUsage(operation, e.model, pageURL, attempt, resp)
			break
		}
		if !isRetryable(err) || attempt == maxAttempts {
			return domain.BookmarkEnrichment{}, fmt.Errorf("gemini generate content: %w", err)
		}

		backoff := time.Duration(attempt) * time.Second
		select {
		case <-ctx.Done():
			return domain.BookmarkEnrichment{}, ctx.Err()
		case <-time.After(backoff):
		}
	}

	text := extractJSON(resp.Text())
	if text == "" {
		return domain.BookmarkEnrichment{}, fmt.Errorf("empty gemini response")
	}

	var parsed struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return domain.BookmarkEnrichment{}, fmt.Errorf("parse gemini json: %w", err)
	}

	return domain.BookmarkEnrichment{
		Title:       normalizeTitle(parsed.Title),
		Description: normalizeDescription(parsed.Description),
		Category:    normalizeCategory(parsed.Category),
		Tags:        normalizeTags(parsed.Tags),
	}, nil
}

// IsUnavailableEnrichment reports joke/blocked fallback payloads that must not be saved.
func IsUnavailableEnrichment(enrichment domain.BookmarkEnrichment) bool {
	for _, tag := range enrichment.Tags {
		if tag == "недоступно" {
			return true
		}
	}

	desc := strings.ToLower(enrichment.Description)
	jokePhrases := []string{
		"стесняется",
		"не пустили",
		"не в настроении",
		"закрылась",
		"загляни сам",
		"содержимое не показала",
	}
	for _, phrase := range jokePhrases {
		if strings.Contains(desc, phrase) {
			return true
		}
	}

	return false
}

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}

	return raw
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "503") ||
		strings.Contains(msg, "429") ||
		strings.Contains(msg, "unavailable") ||
		strings.Contains(msg, "high demand") ||
		strings.Contains(msg, "resource_exhausted")
}

func logTokenUsage(operation, model, pageURL string, attempt int, resp *genai.GenerateContentResponse) {
	if resp == nil {
		log.Printf("[GEMINI-TOKEN] op=%s model=%s url=%s attempt=%d usage=unavailable", operation, model, pageURL, attempt)
		return
	}

	usage := resp.UsageMetadata
	if usage == nil {
		log.Printf("[GEMINI-TOKEN] op=%s model=%s url=%s attempt=%d usage=empty", operation, model, pageURL, attempt)
		return
	}

	log.Printf(
		"[GEMINI-TOKEN] op=%s model=%s url=%s attempt=%d prompt=%d tool=%d output=%d thoughts=%d total=%d",
		operation,
		model,
		pageURL,
		attempt,
		usage.PromptTokenCount,
		usage.ToolUsePromptTokenCount,
		usage.CandidatesTokenCount,
		usage.ThoughtsTokenCount,
		usage.TotalTokenCount,
	)
}
