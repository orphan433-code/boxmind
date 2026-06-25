package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"pet-link/internal/domain"
	"pet-link/internal/pkg/jwt"
	"pet-link/internal/pkg/otp"
)

type OTPRepository interface {
	InvalidateActiveByEmail(ctx context.Context, email string) error
	Create(ctx context.Context, email, codeHash string, expiresAt time.Time) error
	GetLatestActive(ctx context.Context, email string) (domain.LoginOTP, error)
	MarkUsed(ctx context.Context, id string) error
}

type EmailSender interface {
	SendLoginCode(ctx context.Context, email, code string) error
}

type TokenProvider interface {
	Generate(userID, email string) (string, error)
	Parse(tokenString string) (jwt.Claims, error)
}

type AuthService interface {
	RequestLogin(ctx context.Context, email string) error
	VerifyLogin(ctx context.Context, email, code string) (domain.VerifyLoginResult, error)
}

type authService struct {
	otpRepo     OTPRepository
	userService UserService
	tokens      TokenProvider
	emailSender EmailSender
	otpSecret   string
	otpTTL      time.Duration
}

func NewAuthService(
	otpRepo OTPRepository,
	userService UserService,
	tokens TokenProvider,
	emailSender EmailSender,
	otpSecret string,
	otpTTL time.Duration,
) AuthService {
	return &authService{
		otpRepo:     otpRepo,
		userService: userService,
		tokens:      tokens,
		emailSender: emailSender,
		otpSecret:   otpSecret,
		otpTTL:      otpTTL,
	}
}

func (s *authService) RequestLogin(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return err
	}

	code, err := otp.Generate6Digit()
	if err != nil {
		return err
	}

	codeHash := otp.Hash(code, s.otpSecret)
	expiresAt := time.Now().Add(s.otpTTL)

	if err := s.otpRepo.Create(ctx, email, codeHash, expiresAt); err != nil {
		return err
	}

	return s.emailSender.SendLoginCode(ctx, email, code)
}

func (s *authService) VerifyLogin(ctx context.Context, email, code string) (domain.VerifyLoginResult, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return domain.VerifyLoginResult{}, err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.VerifyLoginResult{}, fmt.Errorf("code is required")
	}

	loginOTP, err := s.otpRepo.GetLatestActive(ctx, email)
	if err != nil {
		return domain.VerifyLoginResult{}, err
	}

	if loginOTP.CodeHash != otp.Hash(code, s.otpSecret) {
		return domain.VerifyLoginResult{}, domain.ErrInvalidOTP
	}

	if err := s.otpRepo.MarkUsed(ctx, loginOTP.ID); err != nil {
		return domain.VerifyLoginResult{}, err
	}

	user, err := s.userService.GetOrCreate(ctx, email)
	if err != nil {
		return domain.VerifyLoginResult{}, err
	}

	accessToken, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return domain.VerifyLoginResult{}, err
	}

	return domain.VerifyLoginResult{
		Tokens: domain.AuthTokens{AccessToken: accessToken},
		User:   user,
	}, nil
}

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return fmt.Errorf("invalid email")
	}
	return nil
}

type ConsoleEmailSender struct{}

func NewConsoleEmailSender() *ConsoleEmailSender {
	return &ConsoleEmailSender{}
}

func (s *ConsoleEmailSender) SendLoginCode(ctx context.Context, email, code string) error {
	_ = ctx
	log.Printf("[DEV] login code for %s: %s", email, code)
	return nil
}
