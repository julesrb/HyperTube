package auth

import (
	"context"
	"errors"

	"hypertube/api/internal/models"
)

var errInvalidCredentials = errors.New("invalid credentials")

func (h *Handler) authenticatePassword(ctx context.Context, login string, password string) (models.User, error) {
	user, err := h.store.FindUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return models.User{}, errInvalidCredentials
		}
		return models.User{}, err
	}

	if user.PasswordHash == "" || !CheckPassword(user.PasswordHash, password) {
		return models.User{}, errInvalidCredentials
	}

	return user, nil
}
