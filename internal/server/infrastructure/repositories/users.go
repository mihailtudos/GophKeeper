package repositories

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"log/slog"
	"strings"
)

type UserRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(ctx context.Context, db *sql.DB, logger *slog.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logger,
	}
}

// GenerateSalt creates a new random 16-byte salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

func (ur *UserRepo) Create(ctx context.Context, user domain.User) error {
	op := "repositories.UserRepo.Create"

	query := `INSERT INTO users (id, username, password_hash, salt) VALUES ($1, $2, $3, $4)`

	salt, err := GenerateSalt()
	if err != nil {
		ur.logger.ErrorContext(ctx, "failed to generate salt", "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = ur.db.Exec(query, user.ID, user.Email, user.Password, salt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("%s: %w", op, ErrUniqueConstraintViolation)
		}

		ur.logger.ErrorContext(ctx, "failed to create user", "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (ur *UserRepo) GetByUsername(ctx context.Context, email string) (*domain.User, error) {
	op := "repositories.UserRepo.GetByUsername"

	query := `SELECT id, username, password_hash FROM users WHERE username = $1`
	row := ur.db.QueryRowContext(ctx, query, email)
	var user domain.User

	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrRecordNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (ur *UserRepo) GetUserSalt(ctx context.Context, userID string) ([]byte, error) {
	op := "repositories.UserRepo.GetUserSalt"

	query := `SELECT salt FROM users WHERE id = $1`
	row := ur.db.QueryRowContext(ctx, query, userID)
	var salt []byte
	err := row.Scan(&salt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrRecordNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return salt, nil
}
