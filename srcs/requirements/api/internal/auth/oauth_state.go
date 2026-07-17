package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hypertube/api/internal/i18n"
)

const (
	oauthStateCookieName       = "hypertube_oauth_42_state"
	githubOAuthStateCookieName = "hypertube_oauth_github_state"
	gitlabOAuthStateCookieName = "hypertube_oauth_gitlab_state"
	oauthStateTTL              = 10 * time.Minute
)

func validOAuthState(r *http.Request, cookieName string, state string) bool {
	cookie, err := r.Cookie(cookieName)
	return err == nil && cookie.Value != "" && cookie.Value == state
}

func (h *Handler) clearOAuthState(w http.ResponseWriter, r *http.Request, cookieName string) {
	http.SetCookie(w, oauthStateCookie(r, cookieName, "", -1))
	http.SetCookie(w, oauthLocaleCookie(r, cookieName, "", -1))
	http.SetCookie(w, oauthRedirectCookie(r, cookieName, "", -1))
}

func oauthStateCookie(r *http.Request, name string, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
	}
}

func oauthLocaleCookie(r *http.Request, stateCookieName string, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     oauthLocaleCookieName(stateCookieName),
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
	}
}

func oauthRedirectCookie(r *http.Request, stateCookieName string, value string, maxAge int) *http.Cookie {
	cookieValue := ""
	if maxAge >= 0 && value != "" {
		cookieValue = url.QueryEscape(value)
	}
	return &http.Cookie{
		Name:     oauthRedirectCookieName(stateCookieName),
		Value:    cookieValue,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
	}
}

func oauthRedirectFromRequest(r *http.Request) string {
	raw := strings.TrimSpace(r.URL.Query().Get("redirect"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("href"))
	}
	if raw == "" || !validOAuthRedirectPath(raw) {
		return ""
	}
	return raw
}

func oauthRedirectFromCookie(r *http.Request, stateCookieName string) string {
	cookie, err := r.Cookie(oauthRedirectCookieName(stateCookieName))
	if err != nil || cookie.Value == "" {
		return ""
	}
	value, err := url.QueryUnescape(cookie.Value)
	if err != nil || !validOAuthRedirectPath(value) {
		return ""
	}
	return value
}

func validOAuthRedirectPath(value string) bool {
	if strings.ContainsAny(value, "\r\n\t\\") {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return parsed.Scheme == "" &&
		parsed.Host == "" &&
		strings.HasPrefix(value, "/") &&
		!strings.HasPrefix(value, "//")
}

func oauthLocaleFromRequest(r *http.Request, stateCookieName string) i18n.Locale {
	if locale := strings.TrimSpace(r.URL.Query().Get("locale")); locale != "" {
		return i18n.FromValue(locale)
	}
	if cookie, err := r.Cookie(oauthLocaleCookieName(stateCookieName)); err == nil && cookie.Value != "" {
		return i18n.FromValue(cookie.Value)
	}
	return i18n.FromRequest(r)
}

func oauthLocaleCookieName(stateCookieName string) string {
	return stateCookieName + "_locale"
}

func oauthRedirectCookieName(stateCookieName string) string {
	return stateCookieName + "_redirect"
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func newOAuthState() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}
