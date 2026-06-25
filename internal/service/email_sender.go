package service

import (
	"log"
	"strings"

	"pet-link/internal/config"
	"pet-link/internal/pkg/mail"
)

func NewEmailSender(cfg config.SMTPConfig) EmailSender {
	if !cfg.Enabled() {
		log.Println("SMTP is not configured; login codes are printed to server logs")
		return NewConsoleEmailSender()
	}

	from := strings.TrimSpace(cfg.From)
	if from == "" {
		from = cfg.Username
	}

	return mail.NewSMTPSender(mail.SMTPConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		From:     from,
	})
}
