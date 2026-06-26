package domain

type MovieMetadataQuery struct {
	Title string
	Year  int
	Kind  string // movie, series, anime, or empty when unknown
}

type MovieMetadata struct {
	Enrichment BookmarkEnrichment
	ImageURL   string
	Confidence float64
	Source     string
}
