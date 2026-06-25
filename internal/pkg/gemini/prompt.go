package gemini

import (
	_ "embed"
	"fmt"
)

//go:embed prompts/enrich.txt
var enrichPromptTemplate string

func enrichPrompt(pageURL string) string {
	return fmt.Sprintf(enrichPromptTemplate, pageURL)
}
