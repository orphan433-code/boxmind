package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port           string
	Env            string
	AllowedOrigins []string
	DatabaseURL    string
	JWTSecret      string
	JWTTTL         time.Duration
	OTPTTL         time.Duration
	GeminiAPIKey   string
	GeminiModel    string
	SMTP           SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func (c SMTPConfig) Enabled() bool {
	return strings.TrimSpace(c.Host) != ""
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_POSTGRE_USER"),
		os.Getenv("DB_POSTGRE_PASSWORD"),
		os.Getenv("DB_POSTGRE_HOST"),
		os.Getenv("DB_POSTGRE_PORT"),
		os.Getenv("DB_POSTGRE_NAME"),
	)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}

	jwtTTL, err := time.ParseDuration(os.Getenv("JWT_TTL"))
	if err != nil || jwtTTL == 0 {
		jwtTTL = 24 * time.Hour
	}

	otpTTL, err := time.ParseDuration(os.Getenv("OTP_TTL"))
	if err != nil || otpTTL == 0 {
		otpTTL = 10 * time.Minute
	}

	return Config{
		Port:           port,
		Env:            env,
		AllowedOrigins: allowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
		DatabaseURL:    databaseURL,
		JWTSecret:      jwtSecret,
		JWTTTL:         jwtTTL,
		OTPTTL:         otpTTL,
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		GeminiModel:    geminiModel(os.Getenv("GEMINI_MODEL")),
		SMTP: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     os.Getenv("SMTP_PORT"),
			Username: os.Getenv("SMTP_USER"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
	}
}

func allowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}

func geminiModel(model string) string {
	if model == "" {
		return "gemini-2.5-flash"
	}
	return model
}
