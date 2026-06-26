package gemini

import (
	_ "embed"
	"fmt"
)

//go:embed prompts/enrich.txt
var enrichPromptTemplate string

//go:embed prompts/classify.txt
var classifyPromptTemplate string

//go:embed prompts/polish.txt
var polishPromptTemplate string

func enrichPrompt(pageURL string) string {
	return fmt.Sprintf(enrichPromptTemplate, pageURL)
}

func classifyPrompt(pageURL, title, description, titleSource string) string {
	return fmt.Sprintf(classifyPromptTemplate, pageURL, title, description, titleSource)
}

func polishPrompt(pageURL, title, description, category, tags string) string {
	return fmt.Sprintf(polishPromptTemplate, pageURL, title, description, category, tags)
}
