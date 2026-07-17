package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"hypertube/api/internal/models"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicateUser   = errors.New("duplicate user")
	ErrPasswordChanged = errors.New("password changed concurrently")
)

type DuplicateUserError struct {
	Fields []string
}

func (e *DuplicateUserError) Error() string {
	if len(e.Fields) == 0 {
		return ErrDuplicateUser.Error()
	}
	return fmt.Sprintf("%s: %s", ErrDuplicateUser, strings.Join(e.Fields, ", "))
}

func (e *DuplicateUserError) Unwrap() error {
	return ErrDuplicateUser
}

func duplicateUserError(fields ...string) error {
	return &DuplicateUserError{Fields: fields}
}

func duplicateUserFields(err error) []string {
	var duplicateErr *DuplicateUserError
	if errors.As(err, &duplicateErr) {
		return duplicateErr.Fields
	}
	return nil
}

type UpdateUserParams struct {
	Email               *string
	Username            *string
	FirstName           *string
	LastName            *string
	ClearProfilePicture bool
	Color               *string
}

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, email, username, first_name, last_name, COALESCE(profile_picture, ''), COALESCE(password_hash, ''), color, created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.PasswordHash, &u.Color, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var total int
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	return total, err
}

func (s *Store) FindUserByID(ctx context.Context, id int64) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, username, first_name, last_name, COALESCE(profile_picture, ''), COALESCE(password_hash, ''), color, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.PasswordHash, &u.Color, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

func (s *Store) UserHasOAuthAccount(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM oauth_accounts
			WHERE user_id = $1
		)
	`, id).Scan(&exists)
	return exists, err
}

func (s *Store) UpdateUser(ctx context.Context, id int64, params UpdateUserParams) (models.User, error) {
	var u models.User

	err := s.db.QueryRow(ctx, `
		UPDATE users
		SET
			email = COALESCE($2, email),
			username = COALESCE($3, username),
			first_name = COALESCE($4, first_name),
			last_name = COALESCE($5, last_name),
			profile_picture = CASE WHEN $6 THEN NULL::text ELSE profile_picture END,
			color = COALESCE($7, color),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, username, first_name, last_name, COALESCE(profile_picture, ''), COALESCE(password_hash, ''), color, created_at, updated_at
	`, id, params.Email, params.Username, params.FirstName, params.LastName, params.ClearProfilePicture, params.Color).
		Scan(&u.ID, &u.Email, &u.Username, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.PasswordHash, &u.Color, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		if isUniqueViolation(err) {
			return models.User{}, duplicateUpdateUserError(err)
		}
		return models.User{}, err
	}
	return u, nil
}

func (s *Store) UpdatePasswordHash(ctx context.Context, id int64, expectedHash, newHash string) error {
	result, err := s.db.Exec(ctx, `
		UPDATE users
		SET password_hash = $3, updated_at = NOW()
		WHERE id = $1 AND password_hash = $2
	`, id, expectedHash, newHash)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPasswordChanged
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func duplicateUpdateUserError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ErrDuplicateUser
	}

	switch pgErr.ConstraintName {
	case "users_username_key":
		return duplicateUserError("username")
	case "users_password_email_key":
		return duplicateUserError("email")
	default:
		return ErrDuplicateUser
	}
}
