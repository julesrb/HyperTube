package auth

import (
	"net/mail"
	"regexp"
	"strings"
)

const (
	minPasswordBytes = 8
	maxPasswordBytes = 72
	maxNameLength    = 100
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

func validateRegisterRequest(req registerRequest) (CreateUserParams, string, bool) {
	email, ok := normalizeEmail(req.Email)
	if !ok {
		return CreateUserParams{}, "valid email is required", false
	}

	username := strings.TrimSpace(req.Username)
	if !usernamePattern.MatchString(username) {
		return CreateUserParams{}, "username must be 3-32 characters and contain only letters, numbers, or underscores", false
	}

	firstName := strings.TrimSpace(req.FirstName)
	if firstName == "" || len(firstName) > maxNameLength {
		return CreateUserParams{}, "first_name is required and must be at most 100 characters", false
	}

	lastName := strings.TrimSpace(req.LastName)
	if lastName == "" || len(lastName) > maxNameLength {
		return CreateUserParams{}, "last_name is required and must be at most 100 characters", false
	}

	if len(req.Password) < minPasswordBytes || len(req.Password) > maxPasswordBytes {
		return CreateUserParams{}, "password must be between 8 and 72 bytes", false
	}

	return CreateUserParams{
		Email:     email,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
	}, "", true
}

func validateLoginRequest(req loginRequest) (string, string, bool) {
	email, ok := normalizeEmail(req.Email)
	if !ok {
		return "", "valid email is required", false
	}
	if req.Password == "" || len(req.Password) > maxPasswordBytes {
		return "", "password is required", false
	}
	return email, "", true
}

func normalizeEmail(raw string) (string, bool) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", false
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", false
	}

	return email, true
}
