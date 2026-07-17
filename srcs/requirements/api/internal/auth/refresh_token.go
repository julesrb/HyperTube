package auth

import (
	"net/http"
	"strings"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/requestjson"
	"hypertube/api/internal/respond"
)

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	setTokenResponseHeaders(w)

	if h.tokens == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}

	var req refreshTokenRequest
	if !requestjson.DecodeJSON(w, r, &req) {
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "refresh_token", i18n.MsgRefreshTokenRequired)
		return
	}

	claims, err := h.tokens.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", i18n.MsgInvalidRefreshToken)
		return
	}

	if h.store == nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}

	exists, err := h.store.UserExists(r.Context(), claims.UserID)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgAuthServiceUnavailable)
		return
	}
	if !exists {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", i18n.MsgInvalidRefreshToken)
		return
	}

	accessToken, expiresIn, err := h.issueAccessToken(claims.UserID)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreateAccessToken)
		return
	}

	respond.Data(w, http.StatusOK, refreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	})
}
