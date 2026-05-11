package auth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestFortyTwoOAuthAuthCodeURL(t *testing.T) {
	provider := NewFortyTwoOAuth(FortyTwoOAuthConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost:8080/api/v1/auth/42/callback",
	})

	authURL, err := provider.AuthCodeURL("state-123")
	if err != nil {
		t.Fatalf("auth code url: %v", err)
	}

	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "api.intra.42.fr" || parsed.Path != "/oauth/authorize" {
		t.Fatalf("unexpected authorize URL: %q", authURL)
	}

	query := parsed.Query()
	if query.Get("client_id") != "client-id" {
		t.Fatalf("expected client_id, got %q", query.Get("client_id"))
	}
	if query.Get("redirect_uri") != "http://localhost:8080/api/v1/auth/42/callback" {
		t.Fatalf("expected redirect_uri, got %q", query.Get("redirect_uri"))
	}
	if query.Get("response_type") != "code" {
		t.Fatalf("expected response_type=code, got %q", query.Get("response_type"))
	}
	if query.Get("scope") != "public" {
		t.Fatalf("expected scope=public, got %q", query.Get("scope"))
	}
	if query.Get("state") != "state-123" {
		t.Fatalf("expected state, got %q", query.Get("state"))
	}
}

func TestFortyTwoOAuthRequiresConfiguration(t *testing.T) {
	provider := NewFortyTwoOAuth(FortyTwoOAuthConfig{ClientID: "client-id"})

	if _, err := provider.AuthCodeURL("state"); !errors.Is(err, ErrOAuthNotConfigured) {
		t.Fatalf("expected ErrOAuthNotConfigured from AuthCodeURL, got %v", err)
	}
	if _, err := provider.Exchange(context.Background(), "code"); !errors.Is(err, ErrOAuthNotConfigured) {
		t.Fatalf("expected ErrOAuthNotConfigured from Exchange, got %v", err)
	}
}

func TestFortyTwoOAuthExchangeFetchesProfile(t *testing.T) {
	var tokenRequested bool
	var profileRequested bool
	provider := NewFortyTwoOAuth(FortyTwoOAuthConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost:8080/api/v1/auth/42/callback",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				switch req.URL.String() {
				case fortyTwoTokenURL:
					tokenRequested = true
					if req.Method != http.MethodPost {
						t.Fatalf("expected token request POST, got %s", req.Method)
					}
					if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
						t.Fatalf("expected form content type, got %q", got)
					}
					if err := req.ParseForm(); err != nil {
						t.Fatalf("parse token form: %v", err)
					}
					if req.PostForm.Get("grant_type") != "authorization_code" {
						t.Fatalf("expected grant_type authorization_code, got %q", req.PostForm.Get("grant_type"))
					}
					if req.PostForm.Get("client_id") != "client-id" {
						t.Fatalf("expected client_id, got %q", req.PostForm.Get("client_id"))
					}
					if req.PostForm.Get("client_secret") != "client-secret" {
						t.Fatalf("expected client_secret, got %q", req.PostForm.Get("client_secret"))
					}
					if req.PostForm.Get("code") != "valid-code" {
						t.Fatalf("expected trimmed code, got %q", req.PostForm.Get("code"))
					}
					return jsonResponse(req, http.StatusOK, `{"access_token":"token-123","token_type":"bearer","expires_in":7200}`), nil

				case fortyTwoMeURL:
					profileRequested = true
					if req.Method != http.MethodGet {
						t.Fatalf("expected profile request GET, got %s", req.Method)
					}
					if got := req.Header.Get("Authorization"); got != "Bearer token-123" {
						t.Fatalf("expected bearer token header, got %q", got)
					}
					return jsonResponse(req, http.StatusOK, `{"id":123,"email":"ft.user@example.com","login":"ft_user","first_name":"Forty","last_name":"Two"}`), nil

				default:
					t.Fatalf("unexpected request URL: %s", req.URL.String())
					return nil, nil
				}
			}),
		},
	})

	identity, err := provider.Exchange(context.Background(), " valid-code ")
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}

	if !tokenRequested {
		t.Fatal("expected token request")
	}
	if !profileRequested {
		t.Fatal("expected profile request")
	}
	if identity.Provider != fortyTwoProvider {
		t.Fatalf("expected provider %q, got %q", fortyTwoProvider, identity.Provider)
	}
	if identity.ProviderUserID != "123" {
		t.Fatalf("expected provider user id 123, got %q", identity.ProviderUserID)
	}
	if identity.Email != "ft.user@example.com" || identity.Username != "ft_user" {
		t.Fatalf("unexpected identity: %+v", identity)
	}
}

func TestFortyTwoOAuthExchangeRejectsIncompleteProfile(t *testing.T) {
	provider := NewFortyTwoOAuth(FortyTwoOAuthConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost:8080/api/v1/auth/42/callback",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				switch req.URL.String() {
				case fortyTwoTokenURL:
					return jsonResponse(req, http.StatusOK, `{"access_token":"token-123"}`), nil
				case fortyTwoMeURL:
					return jsonResponse(req, http.StatusOK, `{"id":0,"login":""}`), nil
				default:
					t.Fatalf("unexpected request URL: %s", req.URL.String())
					return nil, nil
				}
			}),
		},
	})

	if _, err := provider.Exchange(context.Background(), "valid-code"); err == nil {
		t.Fatal("expected incomplete profile error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
