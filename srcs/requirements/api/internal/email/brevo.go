package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"hypertube/api/internal/i18n"
)

const defaultBrevoAPIURL = "https://api.brevo.com/v3/smtp/email"

type BrevoConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

type BrevoMailer struct {
	apiKey    string
	apiURL    string
	fromEmail string
	fromName  string
	client    *http.Client
}

func NewBrevoMailer(config BrevoConfig) (*BrevoMailer, error) {
	config.APIKey = strings.TrimSpace(config.APIKey)
	config.FromEmail = strings.TrimSpace(config.FromEmail)
	config.FromName = strings.TrimSpace(config.FromName)
	if config.FromName == "" {
		config.FromName = "Hypertube"
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("BREVO_API_KEY is required")
	}
	if _, err := mail.ParseAddress(config.FromEmail); err != nil {
		return nil, fmt.Errorf("MAIL_FROM_EMAIL must be a valid email address")
	}

	return &BrevoMailer{
		apiKey:    config.APIKey,
		apiURL:    defaultBrevoAPIURL,
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
		client:    &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (m *BrevoMailer) SendPasswordReset(ctx context.Context, toEmail string, toName string, resetURL string, expiresIn time.Duration, locale i18n.Locale) error {
	toEmail = strings.TrimSpace(toEmail)
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("recipient email must be valid")
	}

	minutes := int(expiresIn.Minutes())
	if minutes < 1 {
		minutes = 1
	}
	displayName := strings.TrimSpace(toName)

	subject, textBody, htmlBody := passwordResetEmailContent(locale, displayName, resetURL, minutes)

	return m.send(ctx, toEmail, toName, subject, textBody, htmlBody)
}

func passwordResetEmailContent(locale i18n.Locale, displayName string, resetURL string, minutes int) (string, string, string) {
	locale = i18n.FromValue(string(locale))
	greeting := passwordResetGreeting(locale, displayName)
	switch locale {
	case i18n.French:
		return "Réinitialisez votre mot de passe Hypertube",
			fmt.Sprintf(
				"%s,\n\nUtilisez ce lien pour réinitialiser votre mot de passe Hypertube :\n%s\n\nCe lien expire dans %d minutes. Si vous n'êtes pas à l'origine de cette demande, vous pouvez ignorer cet email.\n",
				greeting,
				resetURL,
				minutes,
			),
			fmt.Sprintf(
				`<p>%s,</p><p>Utilisez ce lien pour réinitialiser votre mot de passe Hypertube :</p><p><a href="%s">Réinitialiser le mot de passe</a></p><p>Ce lien expire dans %d minutes. Si vous n'êtes pas à l'origine de cette demande, vous pouvez ignorer cet email.</p>`,
				html.EscapeString(greeting),
				html.EscapeString(resetURL),
				minutes,
			)
	case i18n.German:
		return "Setze dein Hypertube-Passwort zurück",
			fmt.Sprintf(
				"%s,\n\nNutze diesen Link, um dein Hypertube-Passwort zurückzusetzen:\n%s\n\nDieser Link läuft in %d Minuten ab. Wenn du das nicht angefordert hast, kannst du diese E-Mail ignorieren.\n",
				greeting,
				resetURL,
				minutes,
			),
			fmt.Sprintf(
				`<p>%s,</p><p>Nutze diesen Link, um dein Hypertube-Passwort zurückzusetzen:</p><p><a href="%s">Passwort zurücksetzen</a></p><p>Dieser Link läuft in %d Minuten ab. Wenn du das nicht angefordert hast, kannst du diese E-Mail ignorieren.</p>`,
				html.EscapeString(greeting),
				html.EscapeString(resetURL),
				minutes,
			)
	default:
		return "Reset your Hypertube password",
			fmt.Sprintf(
				"%s,\n\nUse this link to reset your Hypertube password:\n%s\n\nThis link expires in %d minutes. If you did not request this, you can ignore this email.\n",
				greeting,
				resetURL,
				minutes,
			),
			fmt.Sprintf(
				`<p>%s,</p><p>Use this link to reset your Hypertube password:</p><p><a href="%s">Reset password</a></p><p>This link expires in %d minutes. If you did not request this, you can ignore this email.</p>`,
				html.EscapeString(greeting),
				html.EscapeString(resetURL),
				minutes,
			)
	}
}

func passwordResetGreeting(locale i18n.Locale, displayName string) string {
	displayName = strings.TrimSpace(displayName)
	switch locale {
	case i18n.French:
		if displayName == "" {
			return "Bonjour"
		}
		return "Bonjour " + displayName
	case i18n.German:
		if displayName == "" {
			return "Hallo"
		}
		return "Hallo " + displayName
	default:
		if displayName == "" {
			return "Hello"
		}
		return "Hello " + displayName
	}
}

func (m *BrevoMailer) send(ctx context.Context, toEmail string, toName string, subject string, textBody string, htmlBody string) error {
	payload := brevoEmailRequest{
		Sender: brevoContact{
			Name:  m.fromName,
			Email: m.fromEmail,
		},
		To: []brevoContact{{
			Name:  strings.TrimSpace(toName),
			Email: toEmail,
		}},
		Subject:     subject,
		TextContent: textBody,
		HTMLContent: htmlBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", m.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("brevo email request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

type brevoEmailRequest struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
	TextContent string         `json:"textContent"`
}

type brevoContact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}
