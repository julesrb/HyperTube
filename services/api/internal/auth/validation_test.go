package auth

import (
	"strings"
	"testing"
)

func TestValidateRegisterRequestNormalizesInput(t *testing.T) {
	params, message, ok := validateRegisterRequest(registerRequest{
		Email:     " Alice@Example.COM ",
		Username:  " alice_1 ",
		FirstName: " Alice ",
		LastName:  " Example ",
		Password:  "correct-horse-battery",
	})

	if !ok {
		t.Fatalf("expected valid request, got message %q", message)
	}
	if params.Email != "alice@example.com" {
		t.Fatalf("expected normalized email, got %q", params.Email)
	}
	if params.Username != "alice_1" {
		t.Fatalf("expected trimmed username, got %q", params.Username)
	}
	if params.FirstName != "Alice" || params.LastName != "Example" {
		t.Fatalf("expected trimmed names, got %+v", params)
	}
}

func TestValidateRegisterRequestRejectsInvalidFields(t *testing.T) {
	valid := registerRequest{
		Email:     "alice@example.com",
		Username:  "alice_1",
		FirstName: "Alice",
		LastName:  "Example",
		Password:  "correct-horse-battery",
	}

	tests := []struct {
		name   string
		mutate func(*registerRequest)
	}{
		{
			name: "invalid email",
			mutate: func(req *registerRequest) {
				req.Email = "not-an-email"
			},
		},
		{
			name: "invalid username",
			mutate: func(req *registerRequest) {
				req.Username = "al"
			},
		},
		{
			name: "missing first name",
			mutate: func(req *registerRequest) {
				req.FirstName = " "
			},
		},
		{
			name: "missing last name",
			mutate: func(req *registerRequest) {
				req.LastName = " "
			},
		},
		{
			name: "short password",
			mutate: func(req *registerRequest) {
				req.Password = strings.Repeat("a", minPasswordBytes-1)
			},
		},
		{
			name: "long password",
			mutate: func(req *registerRequest) {
				req.Password = strings.Repeat("a", maxPasswordBytes+1)
			},
		},
		{
			name: "long first name",
			mutate: func(req *registerRequest) {
				req.FirstName = strings.Repeat("a", maxNameLength+1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			tt.mutate(&req)

			_, message, ok := validateRegisterRequest(req)
			if ok {
				t.Fatalf("expected invalid request for %s", tt.name)
			}
			if message == "" {
				t.Fatal("expected validation message")
			}
		})
	}
}

func TestValidateLoginRequestNormalizesEmail(t *testing.T) {
	email, message, ok := validateLoginRequest(loginRequest{
		Email:    " Alice@Example.COM ",
		Password: "correct-horse-battery",
	})

	if !ok {
		t.Fatalf("expected valid login, got message %q", message)
	}
	if email != "alice@example.com" {
		t.Fatalf("expected normalized email, got %q", email)
	}
}

func TestValidateLoginRequestRejectsInvalidInput(t *testing.T) {
	tests := []loginRequest{
		{Email: "not-an-email", Password: "correct-horse-battery"},
		{Email: "alice@example.com", Password: ""},
		{Email: "alice@example.com", Password: strings.Repeat("a", maxPasswordBytes+1)},
	}

	for _, req := range tests {
		_, message, ok := validateLoginRequest(req)
		if ok {
			t.Fatalf("expected invalid login request: %+v", req)
		}
		if message == "" {
			t.Fatal("expected validation message")
		}
	}
}
