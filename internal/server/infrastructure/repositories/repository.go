package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"log/slog"
)

var (
	ErrUniqueConstraintViolation = errors.New("unique constraint violation")
	ErrRecordNotFound            = errors.New("record not found")
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByUsername(ctx context.Context, email string) (*domain.User, error)
	GetUserSalt(ctx context.Context, userID string) ([]byte, error)
}

type TokenRepository interface {
	Create(ctx context.Context, token RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenID string) (*RefreshToken, error)
}

type SecretRepository interface {
	Create(ctx context.Context, secret domain.Secret) error
	GetSecretByID(ctx context.Context, secretID string) (*domain.Secret, error)
	GetUserSecrets(ctx context.Context, userID string) (*[]domain.Secret, error)
}

type Repository struct {
	UserRepository   UserRepository
	TokenRepository  TokenRepository
	SecretRepository SecretRepository
}

func NewRepository(ctx context.Context, db *sql.DB, logger *slog.Logger) (*Repository, error) {
	return &Repository{
		UserRepository:   NewUserRepository(ctx, db, logger),
		TokenRepository:  NewTokenRepository(ctx, db, logger),
		SecretRepository: NewSecretRepository(ctx, db, logger),
	}, nil
}
