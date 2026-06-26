package movies

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"pet-link/internal/domain"
)

const (
	tmdbAPIBase     = "https://api.themoviedb.org/3"
	tmdbImageBase   = "https://image.tmdb.org/t/p/w500"
	minTMDBScore    = 0.62
	minTMDBVoteHint = 3
)

type TMDBProvider struct {
	apiKey string
	client *http.Client
}

func NewTMDBProvider(apiKey string) *TMDBProvider {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil
	}
	return &TMDBProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

func (p *TMDBProvider) Lookup(ctx context.Context, query domain.MovieMetadataQuery) (domain.MovieMetadata, bool) {
	title := cleanMovieQuery(query.Title)
	if title == "" {
		return domain.MovieMetadata{}, false
	}

	results, err := p.search(ctx, title)
	if err != nil {
		return domain.MovieMetadata{}, false
	}

	best, ok := pickBestResult(title, query, results)
	if !ok {
		return domain.MovieMetadata{}, false
	}

	meta := domain.MovieMetadata{
		Enrichment: domain.BookmarkEnrichment{
			Title:       best.DisplayTitle(),
			Description: normalizeOverview(best.Overview),
			Category:    "movies",
			Tags:        movieTags(best),
		},
		ImageURL:   best.ImageURL(),
		Confidence: best.score,
		Source:     "tmdb",
	}
	return meta, true
}

func (p *TMDBProvider) search(ctx context.Context, title string) ([]tmdbResult, error) {
	u, err := url.Parse(tmdbAPIBase + "/search/multi")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("api_key", p.apiKey)
	q.Set("query", title)
	q.Set("language", "ru-RU")
	q.Set("include_adult", "false")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb status %d", resp.StatusCode)
	}

	var parsed tmdbSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed.Results, nil
}

type tmdbSearchResponse struct {
	Results []tmdbResult `json:"results"`
}

type tmdbResult struct {
	MediaType     string  `json:"media_type"`
	Title         string  `json:"title"`
	Name          string  `json:"name"`
	OriginalTitle string  `json:"original_title"`
	OriginalName  string  `json:"original_name"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	ReleaseDate   string  `json:"release_date"`
	FirstAirDate  string  `json:"first_air_date"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	GenreIDs      []int   `json:"genre_ids"`
	score         float64
}

func (r tmdbResult) DisplayTitle() string {
	for _, candidate := range []string{r.Title, r.Name, r.OriginalTitle, r.OriginalName} {
		if title := strings.TrimSpace(candidate); title != "" {
			return title
		}
	}
	return ""
}

func (r tmdbResult) ImageURL() string {
	if strings.TrimSpace(r.PosterPath) == "" {
		return ""
	}
	return tmdbImageBase + r.PosterPath
}

func (r tmdbResult) Year() int {
	for _, raw := range []string{r.ReleaseDate, r.FirstAirDate} {
		if len(raw) >= 4 {
			if year, err := strconv.Atoi(raw[:4]); err == nil {
				return year
			}
		}
	}
	return 0
}

func pickBestResult(queryTitle string, query domain.MovieMetadataQuery, results []tmdbResult) (tmdbResult, bool) {
	filtered := make([]tmdbResult, 0, len(results))
	for _, result := range results {
		if result.MediaType != "movie" && result.MediaType != "tv" {
			continue
		}
		if !kindMatches(query.Kind, result.MediaType) {
			continue
		}

		score := titleSimilarity(queryTitle, result.DisplayTitle())
		if result.OriginalTitle != "" {
			score = math.Max(score, titleSimilarity(queryTitle, result.OriginalTitle))
		}
		if result.OriginalName != "" {
			score = math.Max(score, titleSimilarity(queryTitle, result.OriginalName))
		}
		if query.Year > 0 && result.Year() > 0 {
			diff := abs(query.Year - result.Year())
			switch {
			case diff == 0:
				score += 0.08
			case diff <= 1:
				score += 0.04
			case diff > 3:
				score -= 0.12
			}
		}
		if result.PosterPath != "" {
			score += 0.04
		}
		if result.VoteCount >= minTMDBVoteHint {
			score += 0.02
		}
		if result.Popularity > 10 {
			score += 0.02
		}
		result.score = score
		filtered = append(filtered, result)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].score > filtered[j].score
	})
	if len(filtered) == 0 || filtered[0].score < minTMDBScore {
		return tmdbResult{}, false
	}
	return filtered[0], true
}

func kindMatches(kind, mediaType string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "movie", "film", "фильм":
		return mediaType == "movie"
	case "series", "tv", "сериал", "anime", "аниме":
		return mediaType == "tv"
	default:
		return true
	}
}

func movieTags(result tmdbResult) []string {
	first := "фильм"
	if result.MediaType == "tv" {
		first = "сериал"
	}
	if hasGenre(result.GenreIDs, 16) {
		first = "аниме"
	}

	second := "видео"
	for _, id := range result.GenreIDs {
		if tag := genreTag(id); tag != "" && tag != first {
			second = tag
			break
		}
	}
	return []string{first, second}
}

func genreTag(id int) string {
	switch id {
	case 16:
		return "аниме"
	case 18:
		return "драма"
	case 35:
		return "комедия"
	case 27:
		return "ужасы"
	case 28, 10759:
		return "боевик"
	case 12:
		return "приключения"
	case 14, 878, 10765:
		return "фантастика"
	case 53:
		return "триллер"
	case 9648:
		return "детектив"
	case 10749:
		return "романтика"
	case 99:
		return "документальное"
	case 10751:
		return "семейное"
	default:
		return ""
	}
}

func hasGenre(ids []int, target int) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func normalizeOverview(raw string) string {
	overview := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if overview == "" {
		return ""
	}
	if utf8.RuneCountInString(overview) <= 100 {
		return overview
	}

	runes := []rune(overview)
	cut := 100
	for cut > 70 && !unicode.IsSpace(runes[cut-1]) {
		cut--
	}
	if cut <= 70 {
		cut = 100
	}
	return strings.TrimSpace(string(runes[:cut]))
}

var movieQueryNoise = regexp.MustCompile(`(?i)(смотреть\s+онлайн|скачать\s+бесплатно|скачать|бесплатно|без\s+регистрации|онлайн|hd|1080p|720p|latest)`)

func cleanMovieQuery(raw string) string {
	title := strings.ToLower(strings.TrimSpace(raw))
	title = strings.ReplaceAll(title, "ё", "е")
	title = movieQueryNoise.ReplaceAllString(title, " ")
	title = strings.NewReplacer("_", " ", "-", " ", "—", " ", "–", " ", "|", " ").Replace(title)
	title = strings.Join(strings.Fields(title), " ")
	return title
}

func titleSimilarity(a, b string) float64 {
	at := tokenSet(cleanMovieQuery(a))
	bt := tokenSet(cleanMovieQuery(b))
	if len(at) == 0 || len(bt) == 0 {
		return 0
	}

	intersection := 0
	for token := range at {
		if _, ok := bt[token]; ok {
			intersection++
		}
	}
	union := len(at) + len(bt) - intersection
	return float64(intersection) / float64(union)
}

func tokenSet(value string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, token := range strings.Fields(value) {
		if utf8.RuneCountInString(token) < 2 {
			continue
		}
		out[token] = struct{}{}
	}
	return out
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
