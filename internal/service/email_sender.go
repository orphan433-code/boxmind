package service

import (
	"log"
	"strings"

	"pet-link/internal/config"
	"pet-link/internal/pkg/mail"
)

func NewEmailSender(cfg config.MailConfig) EmailSender {
	from := strings.TrimSpace(cfg.From)

	if cfg.Resend.Enabled() {
		log.Println("email provider: resend (https)")
		return mail.NewResendSender(mail.ResendConfig{
			APIKey: cfg.Resend.APIKey,
			From:   from,
		})
	}

	if cfg.SMTP.Enabled() {
		log.Println("email provider: smtp")
		smtpFrom := from
		if smtpFrom == "" {
			smtpFrom = cfg.SMTP.From
		}
		if smtpFrom == "" {
			smtpFrom = cfg.SMTP.Username
		}
		return mail.NewSMTPSender(mail.SMTPConfig{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     smtpFrom,
		})
	}

	log.Println("email provider: console (login codes printed to server logs)")
	return NewConsoleEmailSender()
}
