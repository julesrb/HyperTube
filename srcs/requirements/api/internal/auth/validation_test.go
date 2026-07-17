package auth

import (
	"strings"
	"testing"
)

const (
	testMinPasswordBytes = 8
	testMaxPasswordBytes = 72
	testMaxNameLength    = 100
)

func TestValidateRegisterRequestNormalizesInput(t *testing.T) {
	params, fields, ok := validateRegisterRequest(registerRequest{
		Email:     " Alice@Example.COM ",
		Username:  " alice_1 ",
		FirstName: " Alice ",
		LastName:  " Example ",
		Password:  "correct-horse-battery",
	})

	if !ok {
		t.Fatalf("expected valid request, got fields %+v", fields)
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

func TestValidateRegisterRequestAcceptsFrontendNameAliases(t *testing.T) {
	params, fields, ok := validateRegisterRequest(registerRequest{
		Email:             " Alice@Example.COM ",
		Username:          " alice_1 ",
		FrontendFirstName: " Alice ",
		FrontendLastName:  " Example ",
		Password:          "correct-horse-battery",
	})

	if !ok {
		t.Fatalf("expected valid request, got fields %+v", fields)
	}
	if params.FirstName != "Alice" || params.LastName != "Example" {
		t.Fatalf("expected names from frontend aliases, got %+v", params)
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
		field  string
		mutate func(*registerRequest)
	}{
		{
			name:  "invalid email",
			field: "email",
			mutate: func(req *registerRequest) {
				req.Email = "not-an-email"
			},
		},
		{
			name:  "invalid username",
			field: "username",
			mutate: func(req *registerRequest) {
				req.Username = "al"
			},
		},
		{
			name:  "missing first name",
			field: "first_name",
			mutate: func(req *registerRequest) {
				req.FirstName = " "
			},
		},
		{
			name:  "missing last name",
			field: "last_name",
			mutate: func(req *registerRequest) {
				req.LastName = " "
			},
		},
		{
			name:  "short password",
			field: "password",
			mutate: func(req *registerRequest) {
				req.Password = strings.Repeat("a", testMinPasswordBytes-1)
			},
		},
		{
			name:  "long password",
			field: "password",
			mutate: func(req *registerRequest) {
				req.Password = strings.Repeat("a", testMaxPasswordBytes+1)
			},
		},
		{
			name:  "long first name",
			field: "first_name",
			mutate: func(req *registerRequest) {
				req.FirstName = strings.Repeat("a", testMaxNameLength+1)
			},
		},
		{
			name:  "invalid first name characters",
			field: "first_name",
			mutate: func(req *registerRequest) {
				req.FirstName = "Alice🙂"
			},
		},
		{
			name:  "invalid last name characters",
			field: "last_name",
			mutate: func(req *registerRequest) {
				req.LastName = "Example7"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			tt.mutate(&req)

			_, fields, ok := validateRegisterRequest(req)
			if ok {
				t.Fatalf("expected invalid request for %s", tt.name)
			}
			if fields[tt.field] == "" {
				t.Fatalf("expected validation message for %q, got %+v", tt.field, fields)
			}
		})
	}
}

func TestValidateRegisterRequestRejectsInvalidEmailRules(t *testing.T) {
	valid := registerRequest{
		Email:     "alice@example.com",
		Username:  "alice_1",
		FirstName: "Alice",
		LastName:  "Example",
		Password:  "correct-horse-battery",
	}

	tests := []struct {
		name  string
		email string
	}{
		{name: "prefix too long", email: strings.Repeat("a", 65) + "@example.com"},
		{name: "prefix invalid character", email: "ali!ce@example.com"},
		{name: "prefix starts with period", email: ".alice@example.com"},
		{name: "prefix ends with period", email: "alice.@example.com"},
		{name: "prefix consecutive periods", email: "ali..ce@example.com"},
		{name: "domain too long", email: "alice@" + strings.Repeat("a", 250) + ".com"},
		{name: "domain invalid character", email: "alice@exa_mple.com"},
		{name: "domain missing tld", email: "alice@example"},
		{name: "domain leading hyphen", email: "alice@-example.com"},
		{name: "domain trailing hyphen", email: "alice@example-.com"},
		{name: "tld too short", email: "alice@example.c"},
		{name: "tld too long", email: "alice@example." + strings.Repeat("a", 64)},
		{name: "tld non-letter", email: "alice@example.c0m"},
		{name: "multiple at symbols", email: "alice@@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			req.Email = tt.email

			_, fields, ok := validateRegisterRequest(req)
			if ok {
				t.Fatalf("expected invalid request for %s", tt.email)
			}
			if fields["email"] == "" {
				t.Fatalf("expected email validation message, got %+v", fields)
			}
		})
	}
}

func TestValidateRegisterRequestCollectsMultipleInvalidFields(t *testing.T) {
	_, fields, ok := validateRegisterRequest(registerRequest{
		Email:     "not-an-email",
		Username:  "ab",
		FirstName: "",
		LastName:  "",
		Password:  "short",
	})

	if ok {
		t.Fatal("expected invalid request")
	}
	for _, field := range []string{"email", "username", "first_name", "last_name", "password"} {
		if fields[field] == "" {
			t.Fatalf("expected validation message for %q, got %+v", field, fields)
		}
	}
}

func TestValidateLoginRequestNormalizesEmail(t *testing.T) {
	login, fields, ok := validateLoginRequest(loginRequest{
		Login:    " Alice@Example.COM ",
		Password: "correct-horse-battery",
	})

	if !ok {
		t.Fatalf("expected valid login, got fields %+v", fields)
	}
	if login != "alice@example.com" {
		t.Fatalf("expected normalized email, got %q", login)
	}
}

func TestValidateLoginRequestAcceptsUsername(t *testing.T) {
	login, fields, ok := validateLoginRequest(loginRequest{
		Login:    "alice_1",
		Password: "correct-horse-battery",
	})

	if !ok {
		t.Fatalf("expected valid login, got fields %+v", fields)
	}
	if login != "alice_1" {
		t.Fatalf("expected username login, got %q", login)
	}
}

func TestValidateLoginRequestRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		req   loginRequest
		field string
	}{
		{req: loginRequest{Login: "not-an-email!", Password: "correct-horse-battery"}, field: "login"},
		{req: loginRequest{Login: "alice@example.com", Password: ""}, field: "password"},
		{req: loginRequest{Login: "alice@example.com", Password: strings.Repeat("a", testMaxPasswordBytes+1)}, field: "password"},
		{req: loginRequest{Password: "correct-horse-battery"}, field: "login"},
	}

	for _, tt := range tests {
		_, fields, ok := validateLoginRequest(tt.req)
		if ok {
			t.Fatalf("expected invalid login request: %+v", tt.req)
		}
		if fields[tt.field] == "" {
			t.Fatalf("expected validation message for %q, got %+v", tt.field, fields)
		}
	}
}
