package cardquality

import (
	"strings"
	"unicode/utf8"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/gemini"
	"pet-link/internal/pkg/pagemeta"
)

const (
	minGoodScore      = 7
	minAcceptScore    = 5
	maxDescription    = 100
	targetDescription = 90
)

var seoTitleMarkers = []string{
	"смотреть онлайн",
	"watch online",
	"— смотреть",
	" - смотреть",
	"скачать бесплатно",
	"скачать",
	"бесплатно",
	"без регистрации",
	"купить",
	"цена",
	"отзывы",
	"топ ",
	"лучшие ",
}

var weakDescriptionMarkers = []string{
	"доступн",
	"для просмотра онлайн",
	"смотреть онлайн",
	"фильм или сериал",
	"видеоматериал от",
	"известного российского медиахолдинга",
	"с видео разборами",
	"для начинающих и профессионалов",
	"на нашем сайте",
	"скачать бесплатно",
	"можно скачать",
}

// Score estimates how complete and usable a bookmark card is.
func Score(e domain.BookmarkEnrichment, imageURL string) int {
	score := 0

	if GoodTitle(e.Title) {
		score += 2
	}
	if GoodDescription(e.Description) {
		score += 2
	}
	if e.Category != "" && e.Category != "other" {
		score += 2
	}
	if len(e.Tags) >= 2 {
		score += 2
	}
	if strings.TrimSpace(imageURL) != "" {
		score += 1
	}
	if !gemini.IsUnavailableEnrichment(e) {
		score += 1
	}

	return score
}

// IsGoodEnough reports whether enrichment can stop retrying.
func IsGoodEnough(e domain.BookmarkEnrichment, imageURL string) bool {
	return Score(e, imageURL) >= minGoodScore
}

// IsAcceptable is a softer threshold for partial saves after retries.
func IsAcceptable(e domain.BookmarkEnrichment, imageURL string) bool {
	s := Score(e, imageURL)
	if s >= minAcceptScore && GoodTitle(e.Title) && len(e.Tags) >= 2 && e.Category != "" && e.Category != "other" {
		return true
	}
	return s >= minGoodScore
}

func GoodTitle(title string) bool {
	title = pagemeta.CleanPageTitle(strings.TrimSpace(title))
	if title == "" {
		return false
	}
	return !BadTitle(title)
}

func BadTitle(title string) bool {
	title = pagemeta.CleanPageTitle(strings.TrimSpace(title))
	if title == "" {
		return true
	}
	if strings.HasPrefix(title, "http://") || strings.HasPrefix(title, "https://") {
		return true
	}
	lower := strings.ToLower(title)
	for _, marker := range seoTitleMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	runes := utf8.RuneCountInString(title)
	return runes < 2 || runes > 70
}

func GoodDescription(description string) bool {
	return !BadDescription(description, "")
}

func BadDescription(description, title string) bool {
	description = strings.TrimSpace(description)
	if description == "" {
		return true
	}
	if gemini.IsUnavailableEnrichment(domain.BookmarkEnrichment{Description: description}) {
		return true
	}
	if strings.HasSuffix(description, "…") {
		return true
	}
	if utf8.RuneCountInString(description) > maxDescription {
		return true
	}
	lower := strings.ToLower(description)
	for _, marker := range weakDescriptionMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return repeatsTitle(description, title)
}

func NeedsPolish(e domain.BookmarkEnrichment, imageURL string) bool {
	if strings.TrimSpace(e.Title) == "" && strings.TrimSpace(e.Description) == "" {
		return false
	}
	if strings.TrimSpace(e.Category) == "shopping" && strings.TrimSpace(e.Description) != "" {
		return true
	}
	if strings.TrimSpace(e.Category) == "jobs" && (strings.TrimSpace(e.Title) != "" || strings.TrimSpace(e.Description) != "") {
		return true
	}
	return BadTitle(e.Title) || BadDescription(e.Description, e.Title)
}

func repeatsTitle(description, title string) bool {
	description = strings.ToLower(strings.TrimSpace(description))
	title = strings.ToLower(pagemeta.CleanPageTitle(strings.TrimSpace(title)))
	if description == "" || title == "" {
		return false
	}
	title = strings.Trim(title, ".!?:;,-—– ")
	if utf8.RuneCountInString(title) < 12 {
		return false
	}
	return strings.Contains(description, title) || strings.Contains(title, strings.Trim(description, ".!?:;,-—– "))
}

// Merge combines enrichment layers without degrading a good card.
func Merge(base, patch domain.BookmarkEnrichment) domain.BookmarkEnrichment {
	patch = gemini.NormalizeEnrichment(patch)
	out := base

	if title := pickTitle(base.Title, patch.Title); title != "" {
		out.Title = title
	}
	if desc := pickDescription(base.Description, patch.Description); desc != "" {
		out.Description = desc
	}
	if cat := pickCategory(base.Category, patch.Category); cat != "" {
		out.Category = cat
	}
	if tags := pickTags(base.Tags, patch.Tags); len(tags) > 0 {
		out.Tags = tags
	}

	return out
}

func pickTitle(base, patch string) string {
	base = pagemeta.CleanPageTitle(strings.TrimSpace(base))
	patch = pagemeta.CleanPageTitle(strings.TrimSpace(patch))

	if patch == "" {
		return base
	}
	if base == "" {
		return patch
	}
	if GoodTitle(base) && !GoodTitle(patch) {
		return base
	}
	if !GoodTitle(base) && GoodTitle(patch) {
		return patch
	}
	if titleScore(base) >= titleScore(patch) {
		return base
	}
	return patch
}

func titleScore(title string) int {
	score := 0
	if GoodTitle(title) {
		score += 3
	}
	// Prefer shorter, cleaner titles.
	score += max(0, 40-utf8.RuneCountInString(title)/2)
	return score
}

func pickDescription(base, patch string) string {
	base = strings.TrimSpace(base)
	patch = strings.TrimSpace(patch)

	if patch == "" || !GoodDescription(patch) {
		if GoodDescription(base) {
			return base
		}
		return ""
	}
	if base == "" || !GoodDescription(base) {
		return patch
	}
	if descScore(base) >= descScore(patch) {
		return base
	}
	return patch
}

func descScore(description string) int {
	score := 0
	if GoodDescription(description) {
		score += 4
	}
	runes := utf8.RuneCountInString(description)
	if runes <= targetDescription {
		score += 2
	}
	if strings.HasSuffix(description, ".") || strings.HasSuffix(description, "!") || strings.HasSuffix(description, "?") {
		score += 1
	}
	return score
}

func pickCategory(base, patch string) string {
	base = strings.TrimSpace(base)
	patch = strings.TrimSpace(patch)

	if patch != "" && patch != "other" {
		return patch
	}
	if base != "" {
		return base
	}
	return patch
}

func pickTags(base, patch []string) []string {
	if len(patch) >= 2 {
		return patch
	}
	if len(base) >= 2 {
		return base
	}
	if len(patch) > len(base) {
		return patch
	}
	return base
}
