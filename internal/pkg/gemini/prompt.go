package gemini

import (
	_ "embed"
	"fmt"
)

//go:embed prompts/enrich.txt
var enrichPromptTemplate string

//go:embed prompts/classify.txt
var classifyPromptTemplate string

func enrichPrompt(pageURL string) string {
	return fmt.Sprintf(enrichPromptTemplate, pageURL)
}

func classifyPrompt(pageURL, title, description string) string {
	return fmt.Sprintf(classifyPromptTemplate, pageURL, title, description)
}
