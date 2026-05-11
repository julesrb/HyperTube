package auth

import (
	"strings"
	"testing"
)

func TestNormalizeOAuthUserParamsFillsFallbackFields(t *testing.T) {
	params := normalizeOAuthUserParams(OAuthUserParams{
		Provider:       " 42 ",
		ProviderUserID: " 123 ",
		Email:          "not-an-email",
		Username:       " ",
		FirstName:      " ",
		LastName:       " ",
	})

	if params.Provider != "42" || params.ProviderUserID != "123" {
		t.Fatalf("expected trimmed provider fields, got %+v", params)
	}
	if params.Email != "42-123@oauth.local" {
		t.Fatalf("expected fallback email, got %q", params.Email)
	}
	if params.Username != "user_123" {
		t.Fatalf("expected fallback username, got %q", params.Username)
	}
	if params.FirstName != "user_123" || params.LastName != "42" {
		t.Fatalf("expected fallback names, got %+v", params)
	}
}

func TestNormalizeOAuthUserParamsKeepsValidProfileFields(t *testing.T) {
	params := normalizeOAuthUserParams(OAuthUserParams{
		Provider:       " 42 ",
		ProviderUserID: " 123 ",
		Email:          " FT.User@Example.COM ",
		Username:       " ft_user ",
		FirstName:      " Forty ",
		LastName:       " Two ",
	})

	if params.Email != "ft.user@example.com" {
		t.Fatalf("expected normalized email, got %q", params.Email)
	}
	if params.Username != "ft_user" {
		t.Fatalf("expected trimmed username, got %q", params.Username)
	}
	if params.FirstName != "Forty" || params.LastName != "Two" {
		t.Fatalf("expected trimmed names, got %+v", params)
	}
}

func TestOAuthUsernameBaseSanitizesAndBoundsUsername(t *testing.T) {
	if got := oauthUsernameBase(" John.Doe--42 ", "42", "123"); got != "john_doe_42" {
		t.Fatalf("expected sanitized username, got %q", got)
	}

	if got := oauthUsernameBase("ab", "42", "user-123"); got != "user_42_user123" {
		t.Fatalf("expected fallback username base, got %q", got)
	}

	got := oauthUsernameBase(strings.Repeat("a", 40), "42", "123")
	if len(got) != 32 {
		t.Fatalf("expected username base capped at 32 chars, got len=%d value=%q", len(got), got)
	}
}

func TestUsernameWithSuffixKeepsUsernameWithinLimit(t *testing.T) {
	got := usernameWithSuffix(strings.Repeat("a", 32), "_42_provider_user")

	if len(got) != 32 {
		t.Fatalf("expected username capped at 32 chars, got len=%d value=%q", len(got), got)
	}
	if !strings.HasSuffix(got, "_42_provider_user") {
		t.Fatalf("expected suffix to be preserved, got %q", got)
	}
}
