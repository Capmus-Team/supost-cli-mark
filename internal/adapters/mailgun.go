package adapters

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const mailgunAPIBase = "https://api.mailgun.net/v3"

// EmailSender sends transactional emails. Implementations may use Mailgun, SendGrid, etc.
type EmailSender interface {
	SendPostMessage(ctx context.Context, toEmail, fromEmail, postName, message string) error
}

// MailgunSender sends emails via Mailgun REST API.
type MailgunSender struct {
	apiKey string
	domain string
	from   string
	client *http.Client
}

// NewMailgunSender creates a Mailgun email sender. All params required.
func NewMailgunSender(apiKey, domain, from string) *MailgunSender {
	return &MailgunSender{
		apiKey: strings.TrimSpace(apiKey),
		domain:  strings.TrimSpace(domain),
		from:   strings.TrimSpace(from),
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// SendPostMessage emails the poster (seller) when someone sends a message about their post.
func (m *MailgunSender) SendPostMessage(ctx context.Context, toEmail, fromEmail, postName, message string) error {
	if m.apiKey == "" || m.domain == "" || m.from == "" {
		return fmt.Errorf("mailgun not configured")
	}
	toEmail = strings.TrimSpace(toEmail)
	if toEmail == "" {
		return fmt.Errorf("recipient email is required")
	}

	subject := "Message about your SUpost listing"
	if postName != "" {
		subject = fmt.Sprintf("Message about: %s", truncateForSubject(postName, 50))
	}

	body := fmt.Sprintf(`Someone sent you a message about your post on SUpost.

Post: %s

Message from %s:
---
%s
---

Reply directly to %s to respond.
`, postName, fromEmail, message, fromEmail)

	form := url.Values{}
	form.Set("from", m.from)
	form.Set("to", toEmail)
	form.Set("subject", subject)
	form.Set("text", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s/messages", mailgunAPIBase, m.domain),
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("api", m.apiKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		detail := strings.TrimSpace(string(body))
		if detail != "" {
			return fmt.Errorf("mailgun %d: %s", resp.StatusCode, detail)
		}
		return fmt.Errorf("mailgun returned %d", resp.StatusCode)
	}
	return nil
}

func truncateForSubject(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
