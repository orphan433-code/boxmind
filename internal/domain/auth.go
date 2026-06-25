package domain

import "time"

type LoginOTP struct {
	ID        string
	Email     string
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type AuthTokens struct {
	AccessToken string `json:"access_token"`
}

type VerifyLoginResult struct {
	Tokens AuthTokens `json:"tokens"`
	User   User       `json:"user"`
}
