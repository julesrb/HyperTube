package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"hypertube/api/internal/models"
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
		ProfilePicture: " https://cdn.intra.42.fr/users/123/medium_ft_user.jpg ",
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
	if params.ProfilePicture != "https://cdn.intra.42.fr/users/123/medium_ft_user.jpg" {
		t.Fatalf("expected trimmed profile picture, got %q", params.ProfilePicture)
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

func TestApplyOAuthProfilePreservesProfilePictureWhenProviderPictureChanges(t *testing.T) {
	now := time.Now().UTC()
	user := models.User{
		ID:             42,
		Email:          "oauth@example.com",
		Username:       "oauth_user",
		FirstName:      "Old",
		LastName:       "Name",
		ProfilePicture: "https://hypertube.example/custom.png",
		Color:          models.UserColorGreen,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	returnedUser := user
	returnedUser.FirstName = "New"
	q := &fakeOAuthProfileQuerier{returnedUser: returnedUser}

	updated, err := applyOAuthProfile(context.Background(), q, user, OAuthUserParams{
		Provider:       githubProvider,
		ProviderUserID: "123",
		Username:       "oauth_user",
		FirstName:      "New",
		LastName:       "Name",
		ProfilePicture: "https://provider.example/avatar-new.png",
	})
	if err != nil {
		t.Fatalf("apply OAuth profile: %v", err)
	}

	if updated.ProfilePicture != "https://hypertube.example/custom.png" {
		t.Fatalf("expected local profile picture to be preserved, got %q", updated.ProfilePicture)
	}
	if !q.called {
		t.Fatal("expected profile refresh update query")
	}
	if strings.Contains(q.query, "profile_picture =") {
		t.Fatalf("update query must not write profile_picture: %s", q.query)
	}
	if len(q.args) != 4 {
		t.Fatalf("expected 4 query args, got %d: %+v", len(q.args), q.args)
	}
	if q.args[0] != "oauth_user" || q.args[1] != "New" || q.args[2] != "Name" || q.args[3] != int64(42) {
		t.Fatalf("unexpected query args: %+v", q.args)
	}
}

func TestApplyOAuthProfileDoesNotWriteWhenOnlyProviderPictureChanges(t *testing.T) {
	now := time.Now().UTC()
	user := models.User{
		ID:             42,
		Email:          "oauth@example.com",
		Username:       "oauth_user",
		FirstName:      "Old",
		LastName:       "Name",
		ProfilePicture: "https://hypertube.example/custom.png",
		Color:          models.UserColorGreen,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	q := &fakeOAuthProfileQuerier{returnedUser: user}

	updated, err := applyOAuthProfile(context.Background(), q, user, OAuthUserParams{
		Provider:       githubProvider,
		ProviderUserID: "123",
		Username:       "oauth_user",
		FirstName:      "Old",
		LastName:       "Name",
		ProfilePicture: "https://provider.example/avatar-new.png",
	})
	if err != nil {
		t.Fatalf("apply OAuth profile: %v", err)
	}

	if updated != user {
		t.Fatalf("expected unchanged user, got %+v", updated)
	}
	if q.called {
		t.Fatalf("expected no query when only provider picture changes, got %s", q.query)
	}
}

type fakeOAuthProfileQuerier struct {
	called       bool
	query        string
	args         []any
	returnedUser models.User
}

func (q *fakeOAuthProfileQuerier) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	q.called = true
	q.query = query
	q.args = args
	return fakeOAuthProfileRow{user: q.returnedUser}
}

type fakeOAuthProfileRow struct {
	user models.User
}

func (r fakeOAuthProfileRow) Scan(dest ...any) error {
	if len(dest) != 10 {
		return fmt.Errorf("expected 10 scan destinations, got %d", len(dest))
	}

	*(dest[0].(*int64)) = r.user.ID
	*(dest[1].(*string)) = r.user.Email
	*(dest[2].(*string)) = r.user.Username
	*(dest[3].(*string)) = r.user.FirstName
	*(dest[4].(*string)) = r.user.LastName
	*(dest[5].(*sql.NullString)) = sql.NullString{
		String: r.user.ProfilePicture,
		Valid:  r.user.ProfilePicture != "",
	}
	*(dest[6].(*string)) = r.user.PasswordHash
	*(dest[7].(*string)) = r.user.Color
	*(dest[8].(*time.Time)) = r.user.CreatedAt
	*(dest[9].(*time.Time)) = r.user.UpdatedAt
	return nil
}
