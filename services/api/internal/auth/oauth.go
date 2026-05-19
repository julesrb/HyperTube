package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hypertube/api/internal/models"
	"hypertube/api/internal/respond"
)

const (
	fortyTwoProvider     = "42"
	fortyTwoAuthorizeURL = "https://api.intra.42.fr/oauth/authorize"
	fortyTwoTokenURL     = "https://api.intra.42.fr/oauth/token"
	fortyTwoMeURL        = "https://api.intra.42.fr/v2/me"

	githubProvider     = "github"
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubMeURL        = "https://api.github.com/user"
	githubEmailsURL    = "https://api.github.com/user/emails"

	oauthStateCookieName       = "hypertube_oauth_42_state"
	githubOAuthStateCookieName = "hypertube_oauth_github_state"
	oauthStateTTL              = 10 * time.Minute
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
}

type FortyTwoOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HTTPClient   *http.Client
}

type FortyTwoOAuth struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

func NewFortyTwoOAuth(config FortyTwoOAuthConfig) *FortyTwoOAuth {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &FortyTwoOAuth{
		clientID:     strings.TrimSpace(config.ClientID),
		clientSecret: strings.TrimSpace(config.ClientSecret),
		redirectURL:  strings.TrimSpace(config.RedirectURL),
		httpClient:   httpClient,
	}
}

func (c *FortyTwoOAuth) AuthCodeURL(state string) (string, error) {
	if !c.configured() {
		return "", ErrOAuthNotConfigured
	}

	authURL, err := url.Parse(fortyTwoAuthorizeURL)
	if err != nil {
		return "", err
	}

	query := authURL.Query()
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", c.redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "public")
	query.Set("state", state)
	authURL.RawQuery = query.Encode()

	return authURL.String(), nil
}

func (c *FortyTwoOAuth) Exchange(ctx context.Context, code string) (OAuthIdentity, error) {
	if !c.configured() {
		return OAuthIdentity{}, ErrOAuthNotConfigured
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return OAuthIdentity{}, errors.New("missing authorization code")
	}

	token, err := c.exchangeCode(ctx, code)
	if err != nil {
		return OAuthIdentity{}, err
	}

	profile, err := c.fetchProfile(ctx, token.AccessToken)
	if err != nil {
		return OAuthIdentity{}, err
	}
	if profile.ID == 0 || profile.Login == "" {
		return OAuthIdentity{}, errors.New("42 profile is missing required user fields")
	}

	return OAuthIdentity{
		Provider:       fortyTwoProvider,
		ProviderUserID: strconv.FormatInt(profile.ID, 10),
		Email:          profile.Email,
		Username:       profile.Login,
		FirstName:      profile.FirstName,
		LastName:       profile.LastName,
	}, nil
}

func (c *FortyTwoOAuth) configured() bool {
	return c != nil && c.clientID != "" && c.clientSecret != "" && c.redirectURL != ""
}

func (c *FortyTwoOAuth) exchangeCode(ctx context.Context, code string) (fortyTwoTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", c.redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fortyTwoTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fortyTwoTokenResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fortyTwoTokenResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fortyTwoTokenResponse{}, fmt.Errorf("42 token exchange failed: %s", limitedResponseBody(resp.Body))
	}

	var token fortyTwoTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fortyTwoTokenResponse{}, err
	}
	if token.AccessToken == "" {
		return fortyTwoTokenResponse{}, errors.New("42 token response is missing access_token")
	}

	return token, nil
}

func (c *FortyTwoOAuth) fetchProfile(ctx context.Context, accessToken string) (fortyTwoProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fortyTwoMeURL, nil)
	if err != nil {
		return fortyTwoProfile{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fortyTwoProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fortyTwoProfile{}, fmt.Errorf("42 profile request failed: %s", limitedResponseBody(resp.Body))
	}

	var profile fortyTwoProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return fortyTwoProfile{}, err
	}

	return profile, nil
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

type fortyTwoTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type fortyTwoProfile struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Login     string `json:"login"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type GitHubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HTTPClient   *http.Client
}

type GitHubOAuth struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

func NewGitHubOAuth(config GitHubOAuthConfig) *GitHubOAuth {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &GitHubOAuth{
		clientID:     strings.TrimSpace(config.ClientID),
		clientSecret: strings.TrimSpace(config.ClientSecret),
		redirectURL:  strings.TrimSpace(config.RedirectURL),
		httpClient:   httpClient,
	}
}

func (c *GitHubOAuth) AuthCodeURL(state string) (string, error) {
	if !c.configured() {
		return "", ErrOAuthNotConfigured
	}

	authURL, err := url.Parse(githubAuthorizeURL)
	if err != nil {
		return "", err
	}

	query := authURL.Query()
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", c.redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "read:user user:email")
	query.Set("state", state)
	authURL.RawQuery = query.Encode()

	return authURL.String(), nil
}

func (c *GitHubOAuth) Exchange(ctx context.Context, code string) (OAuthIdentity, error) {
	if !c.configured() {
		return OAuthIdentity{}, ErrOAuthNotConfigured
	}

	code = strings.TrimSpace(code)
	if code == "" {
		return OAuthIdentity{}, errors.New("missing authorization code")
	}

	token, err := c.exchangeCode(ctx, code)
	if err != nil {
		return OAuthIdentity{}, err
	}

	profile, err := c.fetchProfile(ctx, token.AccessToken)
	if err != nil {
		return OAuthIdentity{}, err
	}
	if profile.ID == 0 || profile.Login == "" {
		return OAuthIdentity{}, errors.New("GitHub profile is missing required user fields")
	}

	email, err := c.fetchVerifiedEmail(ctx, token.AccessToken)
	if err != nil {
		return OAuthIdentity{}, err
	}

	firstName, lastName := splitGitHubName(profile.Name)
	return OAuthIdentity{
		Provider:       githubProvider,
		ProviderUserID: strconv.FormatInt(profile.ID, 10),
		Email:          email,
		Username:       profile.Login,
		FirstName:      firstName,
		LastName:       lastName,
	}, nil
}

func (c *GitHubOAuth) configured() bool {
	return c != nil && c.clientID != "" && c.clientSecret != "" && c.redirectURL != ""
}

func (c *GitHubOAuth) exchangeCode(ctx context.Context, code string) (githubTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", c.redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return githubTokenResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return githubTokenResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return githubTokenResponse{}, fmt.Errorf("GitHub token exchange failed: %s", limitedResponseBody(resp.Body))
	}

	var token githubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return githubTokenResponse{}, err
	}
	if token.Error != "" {
		return githubTokenResponse{}, fmt.Errorf("GitHub token exchange failed: %s", githubOAuthErrorMessage(token))
	}
	if token.AccessToken == "" {
		return githubTokenResponse{}, errors.New("GitHub token response is missing access_token")
	}

	return token, nil
}

func (c *GitHubOAuth) fetchProfile(ctx context.Context, accessToken string) (githubProfile, error) {
	req, err := githubAPIRequest(ctx, githubMeURL, accessToken)
	if err != nil {
		return githubProfile{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return githubProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return githubProfile{}, fmt.Errorf("GitHub profile request failed: %s", limitedResponseBody(resp.Body))
	}

	var profile githubProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return githubProfile{}, err
	}

	return profile, nil
}

func (c *GitHubOAuth) fetchVerifiedEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := githubAPIRequest(ctx, githubEmailsURL, accessToken)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("GitHub email request failed: %s", limitedResponseBody(resp.Body))
	}

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return strings.TrimSpace(email.Email), nil
		}
	}
	for _, email := range emails {
		if email.Verified {
			return strings.TrimSpace(email.Email), nil
		}
	}
	return "", nil
}

func githubAPIRequest(ctx context.Context, endpoint string, accessToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "hypertube-api")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}

func githubOAuthErrorMessage(token githubTokenResponse) string {
	if token.ErrorDescription != "" {
		return token.ErrorDescription
	}
	return token.Error
}

func splitGitHubName(name string) (string, string) {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

type githubTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type githubProfile struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

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

func (h *Handler) loginOAuth(w http.ResponseWriter, r *http.Request, provider oauthProvider, providerName string, stateCookieName string) {
	if provider == nil {
		respond.Error(w, http.StatusServiceUnavailable, "OAUTH_NOT_CONFIGURED", providerName+" OAuth is not configured")
		return
	}

	state, err := newOAuthState()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create OAuth state")
		return
	}

	authURL, err := provider.AuthCodeURL(state)
	if err != nil {
		if errors.Is(err, ErrOAuthNotConfigured) {
			respond.Error(w, http.StatusServiceUnavailable, "OAUTH_NOT_CONFIGURED", providerName+" OAuth is not configured")
			return
		}
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to start "+providerName+" OAuth")
		return
	}

	http.SetCookie(w, oauthStateCookie(r, stateCookieName, state, int(oauthStateTTL.Seconds())))
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handler) callbackOAuth(w http.ResponseWriter, r *http.Request, provider oauthProvider, providerName string, stateCookieName string) {
	if provider == nil {
		h.redirectOAuthError(w, r, http.StatusServiceUnavailable, "OAUTH_NOT_CONFIGURED", providerName+" OAuth is not configured")
		return
	}

	if providerError := strings.TrimSpace(r.URL.Query().Get("error")); providerError != "" {
		state := strings.TrimSpace(r.URL.Query().Get("state"))
		h.clearOAuthState(w, r, stateCookieName)
		if state == "" || !validOAuthState(r, stateCookieName, state) {
			h.redirectOAuthError(w, r, http.StatusBadRequest, "INVALID_OAUTH_STATE", "invalid OAuth state")
			return
		}
		h.redirectOAuthError(w, r, http.StatusUnauthorized, "OAUTH_DENIED", providerError)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" || !validOAuthState(r, stateCookieName, state) {
		h.clearOAuthState(w, r, stateCookieName)
		h.redirectOAuthError(w, r, http.StatusBadRequest, "INVALID_OAUTH_STATE", "invalid OAuth state")
		return
	}
	h.clearOAuthState(w, r, stateCookieName)

	identity, err := provider.Exchange(r.Context(), code)
	if err != nil {
		h.redirectOAuthError(w, r, http.StatusBadGateway, "OAUTH_EXCHANGE_FAILED", "failed to exchange "+providerName+" authorization code")
		return
	}

	user, err := h.store.FindOrCreateOAuthUser(r.Context(), OAuthUserParams{
		Provider:       identity.Provider,
		ProviderUserID: identity.ProviderUserID,
		Email:          identity.Email,
		Username:       identity.Username,
		FirstName:      identity.FirstName,
		LastName:       identity.LastName,
	})
	if err != nil {
		h.redirectOAuthError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create OAuth user")
		return
	}

	h.writeOAuthSuccess(w, r, user)
}

func (h *Handler) writeOAuthSuccess(w http.ResponseWriter, r *http.Request, user models.User) {
	token, _, err := h.tokens.CreateAccessToken(user.ID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create token")
		return
	}

	response := authResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(AccessTokenTTL.Seconds()),
		User:        toUserResponse(user),
	}

	if h.frontendAuthCallbackURL == "" {
		respond.Data(w, http.StatusOK, response)
		return
	}

	callbackURL, err := url.Parse(h.frontendAuthCallbackURL)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "invalid frontend auth callback URL")
		return
	}

	userJSON, err := json.Marshal(response.User)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create auth response")
		return
	}

	fragment := url.Values{}
	fragment.Set("access_token", response.AccessToken)
	fragment.Set("token_type", response.TokenType)
	fragment.Set("expires_in", strconv.FormatInt(response.ExpiresIn, 10))
	fragment.Set("user", string(userJSON))
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

func validOAuthState(r *http.Request, cookieName string, state string) bool {
	cookie, err := r.Cookie(cookieName)
	return err == nil && cookie.Value != "" && cookie.Value == state
}

func (h *Handler) clearOAuthState(w http.ResponseWriter, r *http.Request, cookieName string) {
	http.SetCookie(w, oauthStateCookie(r, cookieName, "", -1))
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
