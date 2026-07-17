package models

import (
	"testing"
	"time"
)

func TestIsValidUserColor(t *testing.T) {
	for _, color := range AllowedUserColors {
		if !IsValidUserColor(color) {
			t.Fatalf("expected %q to be valid", color)
		}
	}

	invalidColors := []string{
		"",
		"   ",
		"orange",
		"#747AF5",
		"var(--color-purple)",
		"purple-hover",
		"Purple",
	}
	for _, color := range invalidColors {
		if IsValidUserColor(color) {
			t.Fatalf("expected %q to be invalid", color)
		}
	}
}

func TestRandomUserColorAlwaysReturnsAllowedColor(t *testing.T) {
	for i := 0; i < 100; i++ {
		color := RandomUserColor()
		if !IsValidUserColor(color) {
			t.Fatalf("random color %q is not allowed", color)
		}
	}
}

func TestToUserSmallPrivateProfilePictureNullability(t *testing.T) {
	withoutPicture := ToUserSmallPrivate(User{})
	if withoutPicture.ProfilePicture != nil {
		t.Fatalf("expected nil profile picture, got %+v", withoutPicture.ProfilePicture)
	}

	withPicture := ToUserSmallPrivate(User{ProfilePicture: "https://example.com/avatar.png"})
	if withPicture.ProfilePicture == nil || *withPicture.ProfilePicture != "https://example.com/avatar.png" {
		t.Fatalf("expected profile picture URL, got %+v", withPicture.ProfilePicture)
	}
}

func TestToUserResponseProfilePictureNullability(t *testing.T) {
	withoutPicture := ToUserResponse(User{})
	if withoutPicture.ProfilePicture != nil {
		t.Fatalf("expected nil profile picture, got %+v", withoutPicture.ProfilePicture)
	}

	withPicture := ToUserResponse(User{ProfilePicture: "https://example.com/avatar.png"})
	if withPicture.ProfilePicture == nil || *withPicture.ProfilePicture != "https://example.com/avatar.png" {
		t.Fatalf("expected profile picture URL, got %+v", withPicture.ProfilePicture)
	}

	whitespacePicture := ToUserResponse(User{ProfilePicture: "   "})
	if whitespacePicture.ProfilePicture != nil {
		t.Fatalf("expected whitespace-only profile picture to be nil, got %+v", whitespacePicture.ProfilePicture)
	}
}

func TestToUserProfilePrivate(t *testing.T) {
	createdAt := time.Date(2026, time.June, 20, 12, 0, 0, 0, time.UTC)
	profile := ToUserProfilePrivate(User{
		ID:             7,
		Username:       "alice",
		FirstName:      "Alice",
		LastName:       "Liddell",
		ProfilePicture: "https://example.test/avatar.png",
		Color:          UserColorGreen,
		CreatedAt:      createdAt,
	})

	if profile.CreatedAt != createdAt {
		t.Fatalf("expected created_at %v, got %v", createdAt, profile.CreatedAt)
	}
	if profile.FirstName != "A" || profile.LastName != "L" {
		t.Fatalf("expected private initials A L, got %q %q", profile.FirstName, profile.LastName)
	}
	if profile.ProfilePicture == nil || *profile.ProfilePicture != "https://example.test/avatar.png" {
		t.Fatalf("unexpected profile picture: %+v", profile.ProfilePicture)
	}
}
