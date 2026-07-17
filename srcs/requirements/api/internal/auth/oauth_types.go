package auth

import (
	"context"
	"errors"
	"io"
	"strings"
)

var ErrOAuthNotConfigured = errors.New("oauth provider is not configured")

type oauthProvider interface {
	AuthCodeURL(state string) (string, error)
	Exchange(ctx context.Context, code string) (OAuthIdentity, error)
}

type OAuthIdentity struct {
	Provider       string
	ProviderUserID string
	Email          string
	Username       string
	FirstName      string
	LastName       string
	ProfilePicture string
}

func limitedResponseBody(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 4096))
	if err != nil {
		return "unreadable response body"
	}
	message := strings.TrimSpace(string(data))
	if message == "" {
		return "empty response body"
	}
	return message
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func splitDisplayName(name string) (string, string) {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}
