package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidOTP        = errors.New("invalid or expired otp")
	ErrBookmarkNotFound      = errors.New("bookmark not found")
	ErrBookmarkAlreadyExists = errors.New("bookmark already exists")
)
