// Package mailer sends transactional email via Resend (https://resend.com).
// When no API key is configured it falls back to logging the message, so local
// development still works without external credentials.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const resendEndpoint = "https://api.resend.com/emails"

type Mailer interface {
	Send(ctx context.Context, to, subject, html string) error
}

// resendMailer posts to the Resend API.
type resendMailer struct {
	apiKey string
	from   string
	client *http.Client
}

// logMailer just prints emails to the log (dev fallback when no API key).
type logMailer struct{ from string }

// New returns a Resend-backed mailer when apiKey is set, otherwise a log mailer.
func New(apiKey, from string) Mailer {
	if apiKey == "" {
		log.Println("[mailer] RESEND_API_KEY not set — emails will be logged, not sent")
		return &logMailer{from: from}
	}
	return &resendMailer{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *resendMailer) Send(ctx context.Context, to, subject, html string) error {
	body, err := json.Marshal(map[string]interface{}{
		"from":    m.from,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (m *logMailer) Send(_ context.Context, to, subject, html string) error {
	log.Printf("\n[mailer:log] To: %s | Subject: %s\n%s\n", to, subject, html)
	return nil
}
