package domain

import "time"

type Bookmark struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Enriched    bool      `json:"enriched"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateBookmarkInput struct {
	URL         string
	Title       string
	Description string
	ImageURL    string
	Category    string
	Tags        []string
}
