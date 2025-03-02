package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/mihailtudos/gophkeeper/server/internal/domain"
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

func (ur *UserRepo) Create(ctx context.Context, user domain.User) error {
	fmt.Println("create user called")
	op := "repositories.UserRepo.Create"

	query := `INSERT INTO users (email, password) VALUES ($1, $2)`

	_, err := ur.db.Exec(query, user.Email, user.Password)
	if err != nil {
		ur.logger.ErrorContext(ctx, "failed to create user", "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
