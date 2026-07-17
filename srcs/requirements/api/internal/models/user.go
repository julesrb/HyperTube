package models

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"
)

const (
	UserColorYellow = "yellow"
	UserColorPink   = "pink"
	UserColorGreen  = "green"
	UserColorPurple = "purple"
	UserColorBlue   = "blue"
	UserColorRed    = "red"
)

var AllowedUserColors = []string{
	UserColorYellow,
	UserColorPink,
	UserColorGreen,
	UserColorPurple,
	UserColorBlue,
	UserColorRed,
}

type User struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	ProfilePicture string    `json:"profile_picture"`
	PasswordHash   string    `json:"-"`
	Color          string    `json:"color"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	ProfilePicture *string   `json:"profile_picture"`
	Color          string    `json:"color"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserSmall struct {
	ID             int64   `json:"id"`
	Username       string  `json:"username"`
	ProfilePicture *string `json:"profile_picture"`
	Color          string  `json:"color"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
}

type UserProfile struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	ProfilePicture *string   `json:"profile_picture"`
	Color          string    `json:"color"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	CreatedAt      time.Time `json:"created_at"`
}

func ToUserResponse(u User) UserResponse {
	return UserResponse{
		ID:             u.ID,
		Email:          u.Email,
		Username:       u.Username,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		ProfilePicture: nullableProfilePicture(u.ProfilePicture),
		Color:          u.Color,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func ToUserSmall(u User) UserSmall {
	return UserSmall{
		ID:             u.ID,
		Username:       u.Username,
		ProfilePicture: nullableProfilePicture(u.ProfilePicture),
		Color:          u.Color,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
	}
}

func ToUserSmallPrivate(u User) UserSmall {
	return UserSmall{
		ID:             u.ID,
		Username:       u.Username,
		ProfilePicture: nullableProfilePicture(u.ProfilePicture),
		Color:          u.Color,
		FirstName:      initial(u.FirstName),
		LastName:       initial(u.LastName),
	}
}

func ToUserProfilePrivate(u User) UserProfile {
	return UserProfile{
		ID:             u.ID,
		Username:       u.Username,
		ProfilePicture: nullableProfilePicture(u.ProfilePicture),
		Color:          u.Color,
		FirstName:      initial(u.FirstName),
		LastName:       initial(u.LastName),
		CreatedAt:      u.CreatedAt,
	}
}

func nullableProfilePicture(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

// initial returns the first character of s, or an empty string if s is empty.
func initial(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	return string(r[0])
}

func IsValidUserColor(color string) bool {
	color = strings.TrimSpace(color)
	for _, allowed := range AllowedUserColors {
		if color == allowed {
			return true
		}
	}
	return false
}

func RandomUserColor() string {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(AllowedUserColors))))
	if err != nil {
		return UserColorPurple
	}
	return AllowedUserColors[index.Int64()]
}
