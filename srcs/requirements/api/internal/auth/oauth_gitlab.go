package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	gitlabProvider     = "gitlab"
	gitlabAuthorizeURL = "https://gitlab.com/oauth/authorize"
	gitlabTokenURL     = "https://gitlab.com/oauth/token"
	gitlabMeURL        = "https://gitlab.com/api/v4/user"
)

type GitLabOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HTTPClient   *http.Client
}

type GitLabOAuth struct {
	clientID     string
	clientSecret string
	redirectURL  string
	httpClient   *http.Client
}

func NewGitLabOAuth(config GitLabOAuthConfig) *GitLabOAuth {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &GitLabOAuth{
		clientID:     strings.TrimSpace(config.ClientID),
		clientSecret: strings.TrimSpace(config.ClientSecret),
		redirectURL:  strings.TrimSpace(config.RedirectURL),
		httpClient:   httpClient,
	}
}

func (c *GitLabOAuth) AuthCodeURL(state string) (string, error) {
	if !c.configured() {
		return "", ErrOAuthNotConfigured
	}

	authURL, err := url.Parse(gitlabAuthorizeURL)
	if err != nil {
		return "", err
	}

	query := authURL.Query()
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", c.redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "read_user")
	query.Set("state", state)
	authURL.RawQuery = query.Encode()

	return authURL.String(), nil
}

func (c *GitLabOAuth) Exchange(ctx context.Context, code string) (OAuthIdentity, error) {
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
	if profile.ID == 0 || profile.Username == "" {
		return OAuthIdentity{}, errors.New("GitLab profile is missing required user fields")
	}

	firstName, lastName := splitDisplayName(profile.Name)
	return OAuthIdentity{
		Provider:       gitlabProvider,
		ProviderUserID: strconv.FormatInt(profile.ID, 10),
		Email:          firstNonEmpty(profile.Email, profile.PublicEmail),
		Username:       profile.Username,
		FirstName:      firstName,
		LastName:       lastName,
		ProfilePicture: strings.TrimSpace(profile.AvatarURL),
	}, nil
}

func (c *GitLabOAuth) configured() bool {
	return c != nil && c.clientID != "" && c.clientSecret != "" && c.redirectURL != ""
}

func (c *GitLabOAuth) exchangeCode(ctx context.Context, code string) (gitlabTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", c.redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gitlabTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return gitlabTokenResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return gitlabTokenResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return gitlabTokenResponse{}, fmt.Errorf("GitLab token exchange failed: %s", limitedResponseBody(resp.Body))
	}

	var token gitlabTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return gitlabTokenResponse{}, err
	}
	if token.Error != "" {
		return gitlabTokenResponse{}, fmt.Errorf("GitLab token exchange failed: %s", gitlabOAuthErrorMessage(token))
	}
	if token.AccessToken == "" {
		return gitlabTokenResponse{}, errors.New("GitLab token response is missing access_token")
	}

	return token, nil
}

func (c *GitLabOAuth) fetchProfile(ctx context.Context, accessToken string) (gitlabProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, gitlabMeURL, nil)
	if err != nil {
		return gitlabProfile{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "hypertube-api")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return gitlabProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return gitlabProfile{}, fmt.Errorf("GitLab profile request failed: %s", limitedResponseBody(resp.Body))
	}

	var profile gitlabProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return gitlabProfile{}, err
	}

	return profile, nil
}

func gitlabOAuthErrorMessage(token gitlabTokenResponse) string {
	if token.ErrorDescription != "" {
		return token.ErrorDescription
	}
	return token.Error
}

type gitlabTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	CreatedAt        int64  `json:"created_at"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type gitlabProfile struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	PublicEmail string `json:"public_email"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
}
