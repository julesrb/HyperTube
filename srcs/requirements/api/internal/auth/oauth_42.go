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
	fortyTwoProvider     = "42"
	fortyTwoAuthorizeURL = "https://api.intra.42.fr/oauth/authorize"
	fortyTwoTokenURL     = "https://api.intra.42.fr/oauth/token"
	fortyTwoMeURL        = "https://api.intra.42.fr/v2/me"
)

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
		FirstName:      firstNonEmpty(profile.UsualFirstName, profile.FirstName),
		LastName:       profile.LastName,
		ProfilePicture: profileImageURL(profile.Image),
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

func profileImageURL(image fortyTwoProfileImage) string {
	return firstNonEmpty(
		image.Versions.Medium,
		image.Versions.Large,
		image.Link,
		image.Versions.Small,
		image.Versions.Micro,
	)
}

type fortyTwoTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type fortyTwoProfile struct {
	ID             int64                `json:"id"`
	Email          string               `json:"email"`
	Login          string               `json:"login"`
	FirstName      string               `json:"first_name"`
	LastName       string               `json:"last_name"`
	UsualFirstName string               `json:"usual_first_name"`
	DisplayName    string               `json:"displayname"`
	Image          fortyTwoProfileImage `json:"image"`
}

type fortyTwoProfileImage struct {
	Link     string `json:"link"`
	Versions struct {
		Large  string `json:"large"`
		Medium string `json:"medium"`
		Small  string `json:"small"`
		Micro  string `json:"micro"`
	} `json:"versions"`
}
