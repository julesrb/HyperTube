package email

import (
	"strings"
	"testing"

	"hypertube/api/internal/i18n"
)

func TestPasswordResetEmailContentUsesLocale(t *testing.T) {
	subject, textBody, htmlBody := passwordResetEmailContent(
		i18n.German,
		"Alice Example",
		"https://frontend.test/de/password-reset/set-new-password?token=abc",
		15,
	)

	if subject != "Setze dein Hypertube-Passwort zurück" {
		t.Fatalf("expected German subject, got %q", subject)
	}
	if !strings.Contains(textBody, "Hallo Alice Example,") {
		t.Fatalf("expected German greeting, got %q", textBody)
	}
	if !strings.Contains(textBody, "Dieser Link läuft in 15 Minuten ab.") {
		t.Fatalf("expected German expiry copy, got %q", textBody)
	}
	if !strings.Contains(htmlBody, "Passwort zurücksetzen") {
		t.Fatalf("expected German reset link text, got %q", htmlBody)
	}
}
