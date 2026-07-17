package auth

import (
	"net/http"

	"hypertube/api/internal/models"
)

func (h *Handler) issueAccessToken(userID int64) (string, int64, error) {
	token, _, err := h.tokens.CreateAccessToken(userID)
	if err != nil {
		return "", 0, err
	}
	return token, int64(AccessTokenTTL.Seconds()), nil
}

func (h *Handler) issueRefreshToken(userID int64) (string, error) {
	token, _, err := h.tokens.CreateRefreshToken(userID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (h *Handler) newAuthResponse(user models.User, oauthMethod *string) (authResponse, error) {
	token, expiresIn, err := h.issueAccessToken(user.ID)
	if err != nil {
		return authResponse{}, err
	}

	return authResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        toUserResponse(user, oauthMethod),
	}, nil
}

func (h *Handler) newLoginResponse(user models.User) (authResponse, error) {
	response, err := h.newAuthResponse(user, nil)
	if err != nil {
		return authResponse{}, err
	}

	refreshToken, err := h.issueRefreshToken(user.ID)
	if err != nil {
		return authResponse{}, err
	}
	response.RefreshToken = refreshToken
	return response, nil
}

func setTokenResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}
