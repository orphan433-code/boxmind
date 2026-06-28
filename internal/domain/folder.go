package domain

import "time"

type Folder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateFolderInput struct {
	Name string
}

type UpdateFolderInput struct {
	Name string
}
