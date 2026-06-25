package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const resendEndpoint = "https://api.resend.com/emails"

type ResendConfig struct {
	APIKey string
	From   string
}

type ResendSender struct {
	apiKey string
	from   string
	client *http.Client
}

func NewResendSender(cfg ResendConfig) *ResendSender {
	from := strings.TrimSpace(cfg.From)
	if from == "" {
		from = "Boxmind <onboarding@resend.dev>"
	}

	return &ResendSender{
		apiKey: strings.TrimSpace(cfg.APIKey),
		from:   from,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *ResendSender) SendLoginCode(ctx context.Context, to, code string) error {
	to = strings.TrimSpace(strings.ToLower(to))
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}

	payload := map[string]any{
		"from":    s.from,
		"to":      []string{to},
		"subject": "Код для входа в Boxmind",
		"text": fmt.Sprintf(
			"Привет!\n\nТвой код для входа: %s\n\nКод действует ограниченное время. Если ты не запрашивал вход — просто проигнорируй это письмо.\n\n— Boxmind\n",
			code,
		),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send resend email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	return fmt.Errorf("resend returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}
