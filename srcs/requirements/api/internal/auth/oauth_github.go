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
	githubProvider     = "github"
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubMeURL        = "https://api.github.com/user"
	githubEmailsURL    = "https://api.github.com/user/emails"
)

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

	firstName, lastName := splitDisplayName(profile.Name)
	return OAuthIdentity{
		Provider:       githubProvider,
		ProviderUserID: strconv.FormatInt(profile.ID, 10),
		Email:          email,
		Username:       profile.Login,
		FirstName:      firstName,
		LastName:       lastName,
		ProfilePicture: strings.TrimSpace(profile.AvatarURL),
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

type githubTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type githubProfile struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}
