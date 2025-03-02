package repositories

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/mihailtudos/gophkeeper/server/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
}

type Repository struct {
	UserRepository
}

func NewRepository(ctx context.Context, db *sql.DB, logger *slog.Logger) (*Repository, error) {
	return &Repository{
		UserRepository: NewUserRepository(ctx, db, logger),
	}, nil
}
