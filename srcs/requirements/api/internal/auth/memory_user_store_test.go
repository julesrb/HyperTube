package auth

import (
	"context"
	"errors"
	"testing"

	"hypertube/api/internal/models"
)

func TestMemoryUserStoreFindMissing(t *testing.T) {
	_, err := newMemoryUserStore().FindUserByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMemoryUserStoreUserExists(t *testing.T) {
	store := newMemoryUserStore()
	exists, err := store.UserExists(context.Background(), 42)
	if err != nil {
		t.Fatalf("user exists: %v", err)
	}
	if exists {
		t.Fatal("expected missing user id 42")
	}

	user := createPasswordUser(t, store, "alice@example.com", "alice_1", "right-password")
	exists, err = store.UserExists(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("user exists: %v", err)
	}
	if !exists {
		t.Fatalf("expected user id %d to exist", user.ID)
	}

	store.userExistsErr = errors.New("db down")
	exists, err = store.UserExists(context.Background(), user.ID)
	if err == nil {
		t.Fatal("expected lookup error")
	}
	if exists {
		t.Fatal("expected false when lookup fails")
	}
}

func TestMemoryUserStoreAssignsValidColor(t *testing.T) {
	user, err := newMemoryUserStore().CreateUser(context.Background(), CreateUserParams{
		Email:        "color@example.com",
		Username:     "color_user",
		FirstName:    "Color",
		LastName:     "User",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if !models.IsValidUserColor(user.Color) {
		t.Fatalf("expected valid color, got %q", user.Color)
	}
}

func TestMemoryUserStorePreservesExplicitColor(t *testing.T) {
	user, err := newMemoryUserStore().CreateUser(context.Background(), CreateUserParams{
		Email:        "blue@example.com",
		Username:     "blue_user",
		FirstName:    "Blue",
		LastName:     "User",
		PasswordHash: "hash",
		Color:        models.UserColorBlue,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.Color != models.UserColorBlue {
		t.Fatalf("expected blue, got %q", user.Color)
	}
}

func TestOAuthUsersWithSameEmailStaySeparateByProvider(t *testing.T) {
	store := newMemoryUserStore()

	fortyTwoUser, err := store.FindOrCreateOAuthUser(context.Background(), OAuthUserParams{
		Provider:       fortyTwoProvider,
		ProviderUserID: "42-id",
		Email:          "same@example.com",
		Username:       "forty_two",
		FirstName:      "Forty",
		LastName:       "Two",
	})
	if err != nil {
		t.Fatalf("create 42 user: %v", err)
	}

	githubUser, err := store.FindOrCreateOAuthUser(context.Background(), OAuthUserParams{
		Provider:       githubProvider,
		ProviderUserID: "github-id",
		Email:          "same@example.com",
		Username:       "github_user",
		FirstName:      "Git",
		LastName:       "Hub",
	})
	if err != nil {
		t.Fatalf("create GitHub user: %v", err)
	}

	if githubUser.ID == fortyTwoUser.ID {
		t.Fatalf("expected separate OAuth users for matching provider email, got id %d", githubUser.ID)
	}
	if githubUser.FirstName != "Git" || githubUser.LastName != "Hub" {
		t.Fatalf("expected GitHub profile fields, got %+v", githubUser)
	}
}

func TestMemoryUserStoreOAuthProfileRefreshPreservesColor(t *testing.T) {
	store := newMemoryUserStore()

	firstUser, err := store.FindOrCreateOAuthUser(context.Background(), OAuthUserParams{
		Provider:       githubProvider,
		ProviderUserID: "github-id",
		Email:          "oauth@example.com",
		Username:       "oauth_user",
		FirstName:      "Old",
		LastName:       "Name",
	})
	if err != nil {
		t.Fatalf("create OAuth user: %v", err)
	}
	if !models.IsValidUserColor(firstUser.Color) {
		t.Fatalf("expected valid OAuth color, got %q", firstUser.Color)
	}

	refreshedUser, err := store.FindOrCreateOAuthUser(context.Background(), OAuthUserParams{
		Provider:       githubProvider,
		ProviderUserID: "github-id",
		Email:          "oauth@example.com",
		Username:       "oauth_user",
		FirstName:      "New",
		LastName:       "Profile",
	})
	if err != nil {
		t.Fatalf("refresh OAuth user: %v", err)
	}
	if refreshedUser.Color != firstUser.Color {
		t.Fatalf("expected OAuth refresh to preserve color %q, got %q", firstUser.Color, refreshedUser.Color)
	}
	if refreshedUser.FirstName != "New" || refreshedUser.LastName != "Profile" {
		t.Fatalf("expected refreshed profile fields, got %+v", refreshedUser)
	}
}

func TestMemoryUserStoreOAuthProfileRefreshPreservesProfilePicture(t *testing.T) {
	tests := []struct {
		name                string
		localProfilePicture string
	}{
		{
			name:                "custom picture",
			localProfilePicture: "https://hypertube.example/custom.png",
		},
		{
			name:                "removed picture",
			localProfilePicture: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMemoryUserStore()

			firstUser, err := store.FindOrCreateOAuthUser(context.Background(), OAuthUserParams{
				Provider:       githubProvider,
				ProviderUserID: "github-id",
				Email:          "oauth@example.com",
				Username:       "oauth_user",
				FirstName:      "Old",
				LastName:       "Name",
				ProfilePicture: "https://provider.example/avatar-old.png",
			})
			if err != nil {
				t.Fatalf("create OAuth user: %v", err)
			}

			localUser := firstUser
			localUser.ProfilePicture = tt.localProfilePicture
			store.usersByID[localUser.ID] = localUser
			store.usersByUsername[localUser.Username] = localUser

			refreshedUser, err := store.FindOrCreateOAuthUser(context.Background(), OAuthUserParams{
				Provider:       githubProvider,
				ProviderUserID: "github-id",
				Email:          "oauth@example.com",
				Username:       "oauth_user",
				FirstName:      "New",
				LastName:       "Profile",
				ProfilePicture: "https://provider.example/avatar-new.png",
			})
			if err != nil {
				t.Fatalf("refresh OAuth user: %v", err)
			}

			if refreshedUser.ID != firstUser.ID {
				t.Fatalf("expected OAuth refresh to reuse user id %d, got %d", firstUser.ID, refreshedUser.ID)
			}
			if refreshedUser.ProfilePicture != tt.localProfilePicture {
				t.Fatalf("expected profile picture %q, got %q", tt.localProfilePicture, refreshedUser.ProfilePicture)
			}
			if refreshedUser.Color != firstUser.Color {
				t.Fatalf("expected OAuth refresh to preserve color %q, got %q", firstUser.Color, refreshedUser.Color)
			}
		})
	}
}
