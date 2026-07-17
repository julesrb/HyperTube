package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/models"
	"hypertube/api/internal/respond"
)

func (h *Handler) LoginFortyTwo(w http.ResponseWriter, r *http.Request) {
	h.loginOAuth(w, r, h.fortyTwo, "42", oauthStateCookieName)
}

func (h *Handler) CallbackFortyTwo(w http.ResponseWriter, r *http.Request) {
	h.callbackOAuth(w, r, h.fortyTwo, "42", oauthStateCookieName)
}

func (h *Handler) LoginGitHub(w http.ResponseWriter, r *http.Request) {
	h.loginOAuth(w, r, h.github, "GitHub", githubOAuthStateCookieName)
}

func (h *Handler) CallbackGitHub(w http.ResponseWriter, r *http.Request) {
	h.callbackOAuth(w, r, h.github, "GitHub", githubOAuthStateCookieName)
}

func (h *Handler) LoginGitLab(w http.ResponseWriter, r *http.Request) {
	h.loginOAuth(w, r, h.gitlab, "GitLab", gitlabOAuthStateCookieName)
}

func (h *Handler) CallbackGitLab(w http.ResponseWriter, r *http.Request) {
	h.callbackOAuth(w, r, h.gitlab, "GitLab", gitlabOAuthStateCookieName)
}

func (h *Handler) loginOAuth(w http.ResponseWriter, r *http.Request, provider oauthProvider, providerName string, stateCookieName string) {
	locale := oauthLocaleFromRequest(r, stateCookieName)
	if provider == nil {
		respond.Error(w, http.StatusServiceUnavailable, "OAUTH_NOT_CONFIGURED", i18n.T(locale, i18n.MsgOAuthProviderNotConfigured, providerName))
		return
	}

	redirectPath := oauthRedirectFromRequest(r)
	state, err := newOAuthState()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.T(locale, i18n.MsgFailedCreateOAuthState))
		return
	}

	authURL, err := provider.AuthCodeURL(state)
	if err != nil {
		if errors.Is(err, ErrOAuthNotConfigured) {
			respond.Error(w, http.StatusServiceUnavailable, "OAUTH_NOT_CONFIGURED", i18n.T(locale, i18n.MsgOAuthProviderNotConfigured, providerName))
			return
		}
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.T(locale, i18n.MsgFailedStartOAuth, providerName))
		return
	}

	http.SetCookie(w, oauthStateCookie(r, stateCookieName, state, int(oauthStateTTL.Seconds())))
	http.SetCookie(w, oauthLocaleCookie(r, stateCookieName, locale.String(), int(oauthStateTTL.Seconds())))
	if redirectPath != "" {
		http.SetCookie(w, oauthRedirectCookie(r, stateCookieName, redirectPath, int(oauthStateTTL.Seconds())))
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handler) callbackOAuth(w http.ResponseWriter, r *http.Request, provider oauthProvider, providerName string, stateCookieName string) {
	locale := oauthLocaleFromRequest(r, stateCookieName)
	if provider == nil {
		h.redirectOAuthError(w, r, http.StatusServiceUnavailable, "OAUTH_NOT_CONFIGURED", i18n.T(locale, i18n.MsgOAuthProviderNotConfigured, providerName))
		return
	}

	if providerError := strings.TrimSpace(r.URL.Query().Get("error")); providerError != "" {
		state := strings.TrimSpace(r.URL.Query().Get("state"))
		h.clearOAuthState(w, r, stateCookieName)
		if state == "" || !validOAuthState(r, stateCookieName, state) {
			h.redirectOAuthError(w, r, http.StatusBadRequest, "INVALID_OAUTH_STATE", i18n.T(locale, i18n.MsgInvalidOAuthState))
			return
		}
		h.redirectOAuthError(w, r, http.StatusUnauthorized, "OAUTH_DENIED", i18n.T(locale, i18n.MsgOAuthDenied, providerName))
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" || !validOAuthState(r, stateCookieName, state) {
		h.clearOAuthState(w, r, stateCookieName)
		h.redirectOAuthError(w, r, http.StatusBadRequest, "INVALID_OAUTH_STATE", i18n.T(locale, i18n.MsgInvalidOAuthState))
		return
	}
	redirectPath := oauthRedirectFromCookie(r, stateCookieName)
	h.clearOAuthState(w, r, stateCookieName)

	identity, err := provider.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("oauth callback: %s exchange failed: %v", providerName, err)
		h.redirectOAuthError(w, r, http.StatusBadGateway, "OAUTH_EXCHANGE_FAILED", i18n.T(locale, i18n.MsgFailedExchangeOAuthCode, providerName))
		return
	}

	user, err := h.store.FindOrCreateOAuthUser(r.Context(), OAuthUserParams{
		Provider:       identity.Provider,
		ProviderUserID: identity.ProviderUserID,
		Email:          identity.Email,
		Username:       identity.Username,
		FirstName:      identity.FirstName,
		LastName:       identity.LastName,
		ProfilePicture: identity.ProfilePicture,
	})
	if err != nil {
		log.Printf(
			"oauth callback: failed to create %s OAuth user: provider_user_id=%q username=%q error=%v",
			providerName,
			identity.ProviderUserID,
			identity.Username,
			err,
		)
		h.redirectOAuthError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.T(locale, i18n.MsgFailedCreateOAuthUser))
		return
	}

	h.writeOAuthSuccess(w, r, user, locale, redirectPath, identity.Provider)
}

func (h *Handler) writeOAuthSuccess(w http.ResponseWriter, r *http.Request, user models.User, locale i18n.Locale, redirectPath string, oauthMethod string) {
	setTokenResponseHeaders(w)

	response, err := h.newAuthResponse(user, &oauthMethod)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.T(locale, i18n.MsgFailedCreateToken))
		return
	}

	refreshToken, err := h.issueRefreshToken(user.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.T(locale, i18n.MsgFailedCreateToken))
		return
	}
	response.RefreshToken = refreshToken

	if h.frontendAuthCallbackURL == "" {
		respond.Data(w, http.StatusOK, response)
		return
	}

	callbackURL, err := url.Parse(h.frontendAuthCallbackURL)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.T(locale, i18n.MsgInvalidFrontendAuthCallbackURL))
		return
	}

	userJSON, err := json.Marshal(response.User)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.T(locale, i18n.MsgFailedCreateAuthResponse))
		return
	}

	fragment := url.Values{}
	fragment.Set("access_token", response.AccessToken)
	fragment.Set("refresh_token", response.RefreshToken)
	fragment.Set("token_type", response.TokenType)
	fragment.Set("expires_in", strconv.FormatInt(response.ExpiresIn, 10))
	fragment.Set("user", string(userJSON))
	if redirectPath != "" {
		fragment.Set("redirect", redirectPath)
	}
	callbackURL.Fragment = fragment.Encode()

	http.Redirect(w, r, callbackURL.String(), http.StatusSeeOther)
}

func (h *Handler) redirectOAuthError(w http.ResponseWriter, r *http.Request, status int, code string, message string) {
	if h.frontendAuthCallbackURL == "" {
		respond.Error(w, status, code, message)
		return
	}

	callbackURL, err := url.Parse(h.frontendAuthCallbackURL)
	if err != nil {
		respond.Error(w, status, code, message)
		return
	}

	query := callbackURL.Query()
	query.Set("error", code)
	query.Set("error_description", message)
	callbackURL.RawQuery = query.Encode()

	http.Redirect(w, r, callbackURL.String(), http.StatusSeeOther)
}
