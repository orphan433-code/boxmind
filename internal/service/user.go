package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pet-link/internal/domain"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
}

type UserService interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetOrCreate(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return domain.User{}, fmt.Errorf("email is required")
	}

	return s.repo.GetByEmail(ctx, email)
}

func (s *userService) GetOrCreate(ctx context.Context, email string) (domain.User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return domain.User{}, fmt.Errorf("email is required")
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, err
	}

	return s.repo.Create(ctx, email)
}

func (s *userService) GetByID(ctx context.Context, id string) (domain.User, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.User{}, fmt.Errorf("user id is required")
	}
	return s.repo.GetByID(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
