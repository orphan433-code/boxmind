package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	port := strings.TrimSpace(cfg.Port)
	if port == "" {
		port = "587"
	}

	from := strings.TrimSpace(cfg.From)
	if from == "" {
		from = cfg.Username
	}

	return &SMTPSender{
		cfg: SMTPConfig{
			Host:     strings.TrimSpace(cfg.Host),
			Port:     port,
			Username: strings.TrimSpace(cfg.Username),
			Password: cfg.Password,
			From:     from,
		},
	}
}

func (s *SMTPSender) SendLoginCode(ctx context.Context, to, code string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	to = strings.TrimSpace(strings.ToLower(to))
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}

	subject := "Код для входа в Boxmind"
	body := fmt.Sprintf(
		"Привет!\n\nТвой код для входа: %s\n\nКод действует ограниченное время. Если ты не запрашивал вход — просто проигнорируй это письмо.\n\n— Boxmind\n",
		code,
	)

	return s.send(ctx, to, subject, body)
}

func (s *SMTPSender) send(ctx context.Context, to, subject, body string) error {
	fromAddr, err := parseAddress(s.cfg.From)
	if err != nil {
		return fmt.Errorf("invalid smtp from address: %w", err)
	}

	msg := buildMessage(s.cfg.From, to, subject, body)
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	if s.cfg.Port == "465" {
		return sendImplicitTLS(ctx, addr, s.cfg.Host, auth, fromAddr, []string{to}, msg)
	}

	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(addr, auth, fromAddr, []string{to}, msg)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return fmt.Errorf("send smtp mail: %w", err)
		}
		return nil
	}
}

func sendImplicitTLS(
	ctx context.Context,
	addr, host string,
	auth smtp.Auth,
	from string,
	to []string,
	msg []byte,
) error {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return fmt.Errorf("smtp tls handshake: %w", err)
	}

	client, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt to: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}

func buildMessage(from, to, subject, body string) []byte {
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	return []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s",
		from,
		to,
		encodedSubject,
		body,
	))
}

func parseAddress(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty address")
	}

	if strings.Contains(raw, "<") {
		start := strings.Index(raw, "<")
		end := strings.Index(raw, ">")
		if start >= 0 && end > start {
			raw = raw[start+1 : end]
		}
	}

	if !strings.Contains(raw, "@") {
		return "", fmt.Errorf("invalid address %q", raw)
	}

	return raw, nil
}
