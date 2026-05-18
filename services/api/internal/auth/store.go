package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"hypertube/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrDuplicateUser = errors.New("duplicate user")
)

type userStore interface {
	CreateUser(ctx context.Context, params CreateUserParams) (models.User, error)
	FindUserByEmail(ctx context.Context, email string) (models.User, error)
	FindUserByLogin(ctx context.Context, login string) (models.User, error)
	FindOrCreateOAuthUser(ctx context.Context, params OAuthUserParams) (models.User, error)
}

type Store struct {
	db *pgxpool.Pool
}

type CreateUserParams struct {
	Email        string
	Username     string
	FirstName    string
	LastName     string
	PasswordHash string
}

type OAuthUserParams struct {
	Provider       string
	ProviderUserID string
	Email          string
	Username       string
	FirstName      string
	LastName       string
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, params CreateUserParams) (models.User, error) {
	user, err := scanUser(s.db.QueryRow(ctx, `
		INSERT INTO users (email, username, first_name, last_name, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, username, first_name, last_name, COALESCE(password_hash, ''), created_at, updated_at
	`, params.Email, params.Username, params.FirstName, params.LastName, params.PasswordHash))
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrDuplicateUser
		}
		return models.User{}, err
	}
	return user, nil
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (models.User, error) {
	user, err := scanUser(s.db.QueryRow(ctx, `
		SELECT id, email, username, first_name, last_name, COALESCE(password_hash, ''), created_at, updated_at
		FROM users
		WHERE email = $1
	`, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (s *Store) FindUserByLogin(ctx context.Context, login string) (models.User, error) {
	login = strings.TrimSpace(login)
	email := ""
	if normalizedEmail, ok := normalizeEmail(login); ok {
		email = normalizedEmail
	}

	user, err := scanUser(s.db.QueryRow(ctx, `
		SELECT id, email, username, first_name, last_name, COALESCE(password_hash, ''), created_at, updated_at
		FROM users
		WHERE email = $1 OR username = $2
	`, email, login))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (s *Store) FindOrCreateOAuthUser(ctx context.Context, params OAuthUserParams) (models.User, error) {
	params = normalizeOAuthUserParams(params)
	if params.Provider == "" || params.ProviderUserID == "" {
		return models.User{}, ErrUserNotFound
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.User{}, err
	}
	defer tx.Rollback(ctx)

	user, err := findOAuthUserByAccount(ctx, tx, params.Provider, params.ProviderUserID)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return models.User{}, err
		}
		return user, nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return models.User{}, err
	}

	if params.Email != "" {
		user, err = findUserByEmail(ctx, tx, params.Email)
		if err == nil {
			if err := insertOAuthAccount(ctx, tx, user.ID, params); err != nil {
				return models.User{}, err
			}
			if err := tx.Commit(ctx); err != nil {
				return models.User{}, err
			}
			return user, nil
		}
		if !errors.Is(err, ErrUserNotFound) {
			return models.User{}, err
		}
	}

	username, err := availableOAuthUsername(ctx, tx, params.Username, params.Provider, params.ProviderUserID)
	if err != nil {
		return models.User{}, err
	}

	user, err = scanUser(tx.QueryRow(ctx, `
		INSERT INTO users (email, username, first_name, last_name, password_hash)
		VALUES ($1, $2, $3, $4, '')
		RETURNING id, email, username, first_name, last_name, COALESCE(password_hash, ''), created_at, updated_at
	`, params.Email, username, params.FirstName, params.LastName))
	if err != nil {
		if isUniqueViolation(err) {
			return models.User{}, ErrDuplicateUser
		}
		return models.User{}, err
	}

	if err := insertOAuthAccount(ctx, tx, user.ID, params); err != nil {
		return models.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func scanUser(row pgx.Row) (models.User, error) {
	var user models.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func findUserByEmail(ctx context.Context, q rowQuerier, email string) (models.User, error) {
	user, err := scanUser(q.QueryRow(ctx, `
		SELECT id, email, username, first_name, last_name, COALESCE(password_hash, ''), created_at, updated_at
		FROM users
		WHERE email = $1
	`, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func findOAuthUserByAccount(ctx context.Context, q rowQuerier, provider string, providerUserID string) (models.User, error) {
	user, err := scanUser(q.QueryRow(ctx, `
		SELECT u.id, u.email, u.username, u.first_name, u.last_name, COALESCE(u.password_hash, ''), u.created_at, u.updated_at
		FROM users u
		JOIN oauth_accounts oa ON oa.user_id = u.id
		WHERE oa.provider = $1 AND oa.provider_user_id = $2
	`, provider, providerUserID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func insertOAuthAccount(ctx context.Context, q execer, userID int64, params OAuthUserParams) error {
	_, err := q.Exec(ctx, `
		INSERT INTO oauth_accounts (user_id, provider, provider_user_id, provider_email)
		VALUES ($1, $2, $3, $4)
	`, userID, params.Provider, params.ProviderUserID, nullableString(params.Email))
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateUser
		}
		return err
	}
	return nil
}

func availableOAuthUsername(ctx context.Context, q rowQuerier, rawUsername string, provider string, providerUserID string) (string, error) {
	base := oauthUsernameBase(rawUsername, provider, providerUserID)
	candidates := []string{
		base,
		usernameWithSuffix(base, "_"+provider),
		usernameWithSuffix(base, "_"+providerUserID),
	}

	for i := 2; i <= 100; i++ {
		candidates = append(candidates, usernameWithSuffix(base, fmt.Sprintf("_%s_%d", provider, i)))
	}

	for _, candidate := range candidates {
		exists, err := usernameExists(ctx, q, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", ErrDuplicateUser
}

func usernameExists(ctx context.Context, q rowQuerier, username string) (bool, error) {
	var exists bool
	err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}

func normalizeOAuthUserParams(params OAuthUserParams) OAuthUserParams {
	params.Provider = strings.TrimSpace(params.Provider)
	params.ProviderUserID = strings.TrimSpace(params.ProviderUserID)
	if email, ok := normalizeEmail(params.Email); ok {
		params.Email = email
	} else {
		params.Email = oauthFallbackEmail(params.Provider, params.ProviderUserID)
	}

	params.Username = strings.TrimSpace(params.Username)
	if params.Username == "" {
		params.Username = "user_" + params.ProviderUserID
	}

	params.FirstName = strings.TrimSpace(params.FirstName)
	if params.FirstName == "" {
		params.FirstName = params.Username
	}

	params.LastName = strings.TrimSpace(params.LastName)
	if params.LastName == "" {
		params.LastName = params.Provider
	}

	return params
}

func oauthFallbackEmail(provider string, providerUserID string) string {
	provider = compactIdentifier(provider)
	providerUserID = compactIdentifier(providerUserID)
	if provider == "" {
		provider = "oauth"
	}
	if providerUserID == "" {
		providerUserID = "user"
	}
	return fmt.Sprintf("%s-%s@oauth.local", provider, providerUserID)
}

func oauthUsernameBase(rawUsername string, provider string, providerUserID string) string {
	rawUsername = strings.ToLower(strings.TrimSpace(rawUsername))
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range rawUsername {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'
		if valid {
			builder.WriteRune(r)
			lastUnderscore = r == '_'
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}

	base := strings.Trim(builder.String(), "_")
	if len(base) < 3 {
		base = "user_" + compactIdentifier(provider) + "_" + compactIdentifier(providerUserID)
	}
	if len(base) < 3 {
		base = "oauth_user"
	}
	if len(base) > 32 {
		base = base[:32]
	}
	return base
}

func usernameWithSuffix(base string, suffix string) string {
	suffix = compactUsernameSuffix(suffix)
	if suffix == "" {
		return base
	}
	if !strings.HasPrefix(suffix, "_") {
		suffix = "_" + suffix
	}
	if len(suffix) > 31 {
		suffix = suffix[:31]
	}
	limit := 32 - len(suffix)
	if limit < 1 {
		limit = 1
	}
	if len(base) > limit {
		base = strings.TrimRight(base[:limit], "_")
		if base == "" {
			base = "u"
		}
	}
	return base + suffix
}

func compactIdentifier(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func compactUsernameSuffix(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range value {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'
		if valid {
			builder.WriteRune(r)
			lastUnderscore = r == '_'
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
