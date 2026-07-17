package userinput

import (
	"strings"
	"testing"

	"hypertube/api/internal/i18n"
)

func TestValidateEmailNormalizesInput(t *testing.T) {
	email, message, ok := ValidateEmail(" Alice@Example.COM ")
	if !ok {
		t.Fatalf("expected valid email, got %q", message)
	}
	if email != "alice@example.com" {
		t.Fatalf("expected normalized email, got %q", email)
	}
}

func TestValidateEmailRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		wantMessage i18n.Message
	}{
		{name: "empty", email: " ", wantMessage: i18n.MsgEmailRequired},
		{name: "invalid", email: "not-an-email", wantMessage: i18n.MsgValidEmailRequired},
		{name: "bad prefix", email: ".alice@example.com", wantMessage: i18n.MsgValidEmailRequired},
		{name: "bad tld", email: "alice@example.c0m", wantMessage: i18n.MsgValidEmailRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, message, ok := ValidateEmail(tt.email); ok || message != tt.wantMessage {
				t.Fatalf("expected message %q and ok=false, got message=%q ok=%v", tt.wantMessage, message, ok)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	username, message, ok := ValidateUsername(" alice_1 ")
	if !ok {
		t.Fatalf("expected valid username, got %q", message)
	}
	if username != "alice_1" {
		t.Fatalf("expected trimmed username, got %q", username)
	}

	tests := []struct {
		name        string
		username    string
		wantMessage i18n.Message
	}{
		{name: "empty", username: " ", wantMessage: i18n.MsgUsernameRequired},
		{name: "short", username: "ab", wantMessage: i18n.MsgUsernameTooShort},
		{name: "long", username: strings.Repeat("a", maxUsernameLength+1), wantMessage: i18n.MsgUsernameTooLong},
		{name: "invalid chars", username: "alice!", wantMessage: i18n.MsgUsernameInvalidChars},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, message, ok := ValidateUsername(tt.username); ok || message != tt.wantMessage {
				t.Fatalf("expected message %q and ok=false, got message=%q ok=%v", tt.wantMessage, message, ok)
			}
		})
	}
}

func TestValidateLoginIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		want        string
		wantMessage i18n.Message
		wantOK      bool
	}{
		{name: "email", raw: " Alice@Example.COM ", want: "alice@example.com", wantOK: true},
		{name: "username", raw: " alice_1 ", want: "alice_1", wantOK: true},
		{name: "invalid email", raw: "alice@example", wantMessage: i18n.MsgValidEmailRequired},
		{name: "short username", raw: "ab", wantMessage: i18n.MsgUsernameTooShort},
		{name: "invalid username chars", raw: "alice!", wantMessage: i18n.MsgLoginInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, message, ok := ValidateLoginIdentifier(tt.raw)
			if ok != tt.wantOK || got != tt.want || message != tt.wantMessage {
				t.Fatalf("expected value=%q message=%q ok=%v, got value=%q message=%q ok=%v", tt.want, tt.wantMessage, tt.wantOK, got, message, ok)
			}
		})
	}
}

func TestValidatePasswordVariants(t *testing.T) {
	if message, ok := ValidateRequiredPassword(""); ok || message != i18n.MsgPasswordRequired {
		t.Fatalf("expected required password message, got message=%q ok=%v", message, ok)
	}
	if message, ok := ValidateRequiredPassword("short"); ok || message != i18n.MsgPasswordTooShort {
		t.Fatalf("expected short password message, got message=%q ok=%v", message, ok)
	}
	if message, ok := ValidateRequiredPassword(strings.Repeat("a", maxPasswordBytes+1)); ok || message != i18n.MsgPasswordTooLong {
		t.Fatalf("expected long password message, got message=%q ok=%v", message, ok)
	}

	if message, ok := ValidateLoginPassword("short"); !ok || message != "" {
		t.Fatalf("expected short login password to pass validation, got message=%q ok=%v", message, ok)
	}
	if message, ok := ValidateLoginPassword(""); ok || message != i18n.MsgPasswordRequired {
		t.Fatalf("expected login password required, got message=%q ok=%v", message, ok)
	}

	if message, ok := ValidateUpdatePassword(""); ok || message != i18n.MsgPasswordTooShort {
		t.Fatalf("expected empty update password to be too short, got message=%q ok=%v", message, ok)
	}
	if message, ok := ValidateUpdatePassword("correct-horse-battery"); !ok || message != "" {
		t.Fatalf("expected valid update password, got message=%q ok=%v", message, ok)
	}
}

func TestValidateRequiredPasswordRejectsCommonPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOK   bool
	}{
		{name: "plain common", password: "password"},
		{name: "case and suffix", password: "Password123!"},
		{name: "hyphen and digits suffix", password: "computer-91"},
		{name: "digit suffix", password: "football1"},
		{name: "passphrase", password: "correct-horse-battery", wantOK: true},
		{name: "non-common trimmed suffix", password: "blue-cinema-91", wantOK: true},
		{name: "camel case phrase", password: "MovieLampRiver42", wantOK: true},
		{name: "non-common phrase", password: "not-a-common-passphrase", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, ok := ValidateRequiredPassword(tt.password)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got message=%q ok=%v", tt.wantOK, message, ok)
			}
			if !tt.wantOK && message != i18n.MsgPasswordTooCommon {
				t.Fatalf("expected common password message, got %q", message)
			}
			if tt.wantOK && message != "" {
				t.Fatalf("expected no validation message, got %q", message)
			}
		})
	}
}

func TestValidateUpdatePasswordRejectsCommonPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantOK   bool
	}{
		{name: "plain common", password: "password"},
		{name: "case and suffix", password: "Password123!"},
		{name: "hyphen and digits suffix", password: "computer-91"},
		{name: "digit suffix", password: "football1"},
		{name: "passphrase", password: "correct-horse-battery", wantOK: true},
		{name: "non-common trimmed suffix", password: "blue-cinema-91", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, ok := ValidateUpdatePassword(tt.password)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got message=%q ok=%v", tt.wantOK, message, ok)
			}
			if !tt.wantOK && message != i18n.MsgPasswordTooCommon {
				t.Fatalf("expected common password message, got %q", message)
			}
			if tt.wantOK && message != "" {
				t.Fatalf("expected no validation message, got %q", message)
			}
		})
	}
}

func TestValidateLoginPasswordAllowsExistingCommonPassword(t *testing.T) {
	if message, ok := ValidateLoginPassword("password"); !ok || message != "" {
		t.Fatalf("expected common login password to pass validation, got message=%q ok=%v", message, ok)
	}
}

func TestValidateName(t *testing.T) {
	name, message, ok := ValidateName(" Alice-Louise ", i18n.MsgFirstNameRequired, i18n.MsgFirstNameTooLong, i18n.MsgFirstNameInvalid)
	if !ok {
		t.Fatalf("expected valid name, got %q", message)
	}
	if name != "Alice-Louise" {
		t.Fatalf("expected trimmed name, got %q", name)
	}

	tests := []struct {
		name        string
		raw         string
		wantMessage i18n.Message
	}{
		{name: "empty", raw: " ", wantMessage: i18n.MsgFirstNameRequired},
		{name: "long", raw: strings.Repeat("a", maxNameLength+1), wantMessage: i18n.MsgFirstNameTooLong},
		{name: "invalid", raw: "Alice7", wantMessage: i18n.MsgFirstNameInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, message, ok := ValidateName(tt.raw, i18n.MsgFirstNameRequired, i18n.MsgFirstNameTooLong, i18n.MsgFirstNameInvalid); ok || message != tt.wantMessage {
				t.Fatalf("expected message %q and ok=false, got message=%q ok=%v", tt.wantMessage, message, ok)
			}
		})
	}
}
