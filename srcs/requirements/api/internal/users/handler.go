package users

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/i18n"
	"hypertube/api/internal/models"
	"hypertube/api/internal/requestjson"
	"hypertube/api/internal/respond"
	"hypertube/api/internal/userinput"
)

type userStore interface {
	ListUsers(ctx context.Context, limit, offset int) ([]models.User, error)
	CountUsers(ctx context.Context) (int, error)
	FindUserByID(ctx context.Context, id int64) (models.User, error)
	UserHasOAuthAccount(ctx context.Context, id int64) (bool, error)
	UpdateUser(ctx context.Context, id int64, params UpdateUserParams) (models.User, error)
	UpdatePasswordHash(ctx context.Context, id int64, expectedHash, newHash string) error
}

type Handler struct {
	store userStore
}

func NewHandler(store userStore) *Handler {
	return &Handler{store: store}
}

const (
	userPageLimit = 12
)

type validationErrors map[string]i18n.Message

type updateUserParams struct {
	UpdateUserParams
}

type setPasswordRequest struct {
	CurrentPassword string
	NewPassword     string
	PasswordConfirm *string
}

type setPasswordResponse struct {
	Message string `json:"message"`
}

// ListUsers returns a paginated UserSmall list.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)
	offset := page * userPageLimit

	total, err := h.store.CountUsers(r.Context())
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadUser)
		return
	}

	users, err := h.store.ListUsers(r.Context(), userPageLimit, offset)
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadUser)
		return
	}

	result := make([]models.UserSmall, 0, len(users))
	for _, u := range users {
		result = append(result, models.ToUserSmallPrivate(u))
	}

	respond.ListPaginated(w, http.StatusOK, result, total, page, userPageLimit)
}

// GetUser returns the public profile (UserSmall) for the user with the given id.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgUserNotFound)
		return
	}

	user, err := h.store.FindUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgUserNotFound)
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadUser)
		return
	}

	respond.Data(w, http.StatusOK, models.ToUserProfilePrivate(user))
}

// UpdateUser applies a partial profile update for the authenticated user's own profile.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	authenticatedUserID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgUserNotFound)
		return
	}

	if authenticatedUserID != id {
		respond.LocalizedError(w, r, http.StatusForbidden, "FORBIDDEN", i18n.MsgUserUpdateForbidden)
		return
	}

	params, ok := decodeUpdateUserParams(w, r)
	if !ok {
		return
	}

	if hasOAuthRestrictedUpdate(params) {
		if ok := h.ensureOAuthRestrictedUpdateAllowed(w, r, id, params); !ok {
			return
		}
	}

	user, err := h.store.UpdateUser(r.Context(), id, params.UpdateUserParams)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgUserNotFound)
			return
		}
		if errors.Is(err, ErrDuplicateUser) {
			writeDuplicateUserError(w, r, duplicateUserFields(err))
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedUpdateUser)
		return
	}

	respond.Data(w, http.StatusOK, models.ToUserResponse(user))
}

func hasOAuthRestrictedUpdate(params updateUserParams) bool {
	return params.Email != nil ||
		params.Username != nil ||
		params.FirstName != nil ||
		params.LastName != nil
}

func (h *Handler) ensureOAuthRestrictedUpdateAllowed(w http.ResponseWriter, r *http.Request, id int64, params updateUserParams) bool {
	hasOAuthAccount, err := h.store.UserHasOAuthAccount(r.Context(), id)
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedUpdateUser)
		return false
	}
	if !hasOAuthAccount {
		return true
	}

	fields := validationErrors{}
	if params.Email != nil {
		fields["email"] = i18n.MsgOAuthEmailUpdateForbidden
	}
	if params.Username != nil {
		fields["username"] = i18n.MsgOAuthUsernameUpdateForbidden
	}
	if params.FirstName != nil {
		fields["first_name"] = i18n.MsgOAuthFirstNameUpdateForbidden
	}
	if params.LastName != nil {
		fields["last_name"] = i18n.MsgOAuthLastNameUpdateForbidden
	}
	writeValidationError(w, r, fields)
	return false
}

func parsePage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		return 0
	}
	return page
}

func decodeUpdateUserParams(w http.ResponseWriter, r *http.Request) (updateUserParams, bool) {
	body, ok := requestjson.DecodeJSONObject(w, r, map[string]struct{}{
		"email":           {},
		"username":        {},
		"profile_picture": {},
		"first_name":      {},
		"last_name":       {},
		"color":           {},
	})
	if !ok {
		return updateUserParams{}, false
	}
	if len(body) == 0 {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "body", i18n.MsgInvalidRequestBody)
		return updateUserParams{}, false
	}

	params := updateUserParams{}
	fields := validationErrors{}

	if raw, ok := body["email"]; ok {
		value, ok := decodeStringField(raw, "email", fields)
		if ok {
			if email, message, ok := userinput.ValidateEmail(value); ok {
				params.Email = &email
			} else {
				fields["email"] = message
			}
		}
	}

	if raw, ok := body["username"]; ok {
		value, ok := decodeStringField(raw, "username", fields)
		if ok {
			if username, message, ok := userinput.ValidateUsername(value); ok {
				params.Username = &username
			} else {
				fields["username"] = message
			}
		}
	}

	if raw, ok := body["first_name"]; ok {
		value, ok := decodeStringField(raw, "first_name", fields)
		if ok {
			if firstName, message, ok := userinput.ValidateName(value, i18n.MsgFirstNameRequired, i18n.MsgFirstNameTooLong, i18n.MsgFirstNameInvalid); ok {
				params.FirstName = &firstName
			} else {
				fields["first_name"] = message
			}
		}
	}

	if raw, ok := body["last_name"]; ok {
		value, ok := decodeStringField(raw, "last_name", fields)
		if ok {
			if lastName, message, ok := userinput.ValidateName(value, i18n.MsgLastNameRequired, i18n.MsgLastNameTooLong, i18n.MsgLastNameInvalid); ok {
				params.LastName = &lastName
			} else {
				fields["last_name"] = message
			}
		}
	}

	if raw, ok := body["profile_picture"]; ok {
		if requestjson.IsNull(raw) {
			params.ClearProfilePicture = true
		} else if value, ok := decodeStringField(raw, "profile_picture", fields); ok {
			profilePicture := strings.TrimSpace(value)
			if profilePicture == "" {
				params.ClearProfilePicture = true
			} else {
				fields["profile_picture"] = i18n.MsgProfilePictureUpdateForbidden
			}
		}
	}

	if raw, ok := body["color"]; ok {
		value, ok := decodeStringField(raw, "color", fields)
		if ok {
			color := strings.TrimSpace(value)
			if models.IsValidUserColor(color) {
				params.Color = &color
			} else {
				fields["color"] = i18n.MsgInvalidUserColor
			}
		}
	}

	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return updateUserParams{}, false
	}

	return params, true
}

func decodeSetPasswordRequest(w http.ResponseWriter, r *http.Request) (setPasswordRequest, bool) {
	body, ok := requestjson.DecodeJSONObject(w, r, map[string]struct{}{
		"current_password":     {},
		"new_password":         {},
		"new_password_confirm": {},
	})
	if !ok {
		return setPasswordRequest{}, false
	}

	req := setPasswordRequest{}
	fields := validationErrors{}

	if raw, ok := body["current_password"]; ok {
		if value, validType := decodeStringField(raw, "current-password", fields); validType {
			if message, valid := userinput.ValidateLoginPassword(value); valid {
				req.CurrentPassword = value
			} else {
				fields["current-password"] = message
			}
		}
	} else {
		fields["current-password"] = i18n.MsgPasswordRequired
	}

	if raw, ok := body["new_password"]; ok {
		if value, validType := decodeStringField(raw, "new-password", fields); validType {
			if message, valid := userinput.ValidateLoginPassword(value); valid {
				req.NewPassword = value
			} else {
				fields["new-password"] = message
			}
		}
	} else {
		fields["new-password"] = i18n.MsgPasswordRequired
	}

	if raw, ok := body["new_password_confirm"]; ok {
		if value, validType := decodeStringField(raw, "confirm-new-password", fields); validType {
			req.PasswordConfirm = &value
		}
	}

	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return setPasswordRequest{}, false
	}
	if req.PasswordConfirm != nil && *req.PasswordConfirm != req.NewPassword {
		writeValidationError(w, r, validationErrors{
			"confirm-new-password": i18n.MsgPasswordConfirmationMismatch,
		})
		return setPasswordRequest{}, false
	}

	return req, true
}

func (h *Handler) SetPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}

	req, ok := decodeSetPasswordRequest(w, r)
	if !ok {
		return
	}

	user, err := h.store.FindUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgUserNotFound)
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadUser)
		return
	}

	hasOAuthAccount, err := h.store.UserHasOAuthAccount(r.Context(), userID)
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedChangePassword)
		return
	}
	if hasOAuthAccount {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "new-password", i18n.MsgOAuthPasswordUpdateForbidden)
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		writePasswordFieldError(w, r, http.StatusUnauthorized, "INVALID_CURRENT_PASSWORD", "current-password", i18n.MsgCurrentPasswordInvalid)
		return
	}
	if auth.CheckPassword(user.PasswordHash, req.NewPassword) {
		writePasswordFieldError(w, r, http.StatusConflict, "PASSWORD_UNCHANGED", "new-password", i18n.MsgNewPasswordSameAsCurrent)
		return
	}
	if message, valid := userinput.ValidateRequiredPassword(req.NewPassword); !valid {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "new-password", message)
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "new-password", i18n.MsgPasswordInvalid)
		return
	}
	if err := h.store.UpdatePasswordHash(r.Context(), userID, user.PasswordHash, newHash); err != nil {
		if errors.Is(err, ErrPasswordChanged) {
			writePasswordFieldError(w, r, http.StatusUnauthorized, "INVALID_CURRENT_PASSWORD", "current-password", i18n.MsgCurrentPasswordInvalid)
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedChangePassword)
		return
	}

	respond.Data(w, http.StatusOK, setPasswordResponse{Message: i18n.T(i18n.FromRequest(r), i18n.MsgPasswordChangeSuccess)})
}

func decodeStringField(raw json.RawMessage, field string, fields validationErrors) (string, bool) {
	value, ok := requestjson.DecodeString(raw)
	if !ok {
		fields[field] = i18n.MsgInvalidRequestBody
		return "", false
	}
	return value, true
}

func writeValidationError(w http.ResponseWriter, r *http.Request, fields validationErrors) {
	locale := i18n.FromRequest(r)
	responseFields := make(respond.FieldErrors, len(fields))
	for field, message := range fields {
		responseFields[field] = respond.FieldError{Message: i18n.T(locale, message)}
	}
	respond.ValidationError(w, http.StatusBadRequest, responseFields)
}

func writePasswordFieldError(w http.ResponseWriter, r *http.Request, status int, code, field string, message i18n.Message) {
	respond.ErrorWithFields(w, status, code, respond.FieldErrors{
		field: {Message: i18n.T(i18n.FromRequest(r), message)},
	})
}

func writeDuplicateUserError(w http.ResponseWriter, r *http.Request, fields []string) {
	locale := i18n.FromRequest(r)
	responseFields := respond.FieldErrors{}
	if len(fields) == 0 {
		fields = []string{"email", "username"}
	}

	for _, field := range fields {
		switch field {
		case "email":
			responseFields[field] = respond.FieldError{Message: i18n.T(locale, i18n.MsgEmailAlreadyInUse)}
		case "username":
			responseFields[field] = respond.FieldError{Message: i18n.T(locale, i18n.MsgUsernameAlreadyInUse)}
		}
	}
	if len(responseFields) == 0 {
		responseFields["email"] = respond.FieldError{Message: i18n.T(locale, i18n.MsgEmailAlreadyInUse)}
		responseFields["username"] = respond.FieldError{Message: i18n.T(locale, i18n.MsgUsernameAlreadyInUse)}
	}

	respond.ErrorWithFields(w, http.StatusConflict, "ALREADY_EXIST_ERROR", responseFields)
}
