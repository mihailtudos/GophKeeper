package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mihailtudos/gophkeeper/server/internal/pkg"
)

type TokenRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

type RefreshToken struct {
	ID        string    `json:"id,omitempty"`
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

func NewTokenRepository(ctx context.Context, db *sql.DB, logger *slog.Logger) *TokenRepo {
	return &TokenRepo{
		db:     db,
		logger: logger,
	}
}

func (tr *TokenRepo) Create(ctx context.Context, token RefreshToken) error {
	op := "repositories.TokenRepo.Create"

	query := `INSERT INTO refresh_tokens (id, user_id, token, expires_at) VALUES ($1, $2, $3, $4)`

	if (token.ID == "") {
		token.ID = uuid.New().String()
	}

	_, err := tr.db.Exec(query, token.ID, token.UserID, token.Token, token.ExpiresAt)
	if err != nil {
		tr.logger.ErrorContext(ctx, "failed to store refresh token", slog.String("user_id", token.UserID), slog.String("token", token.Token), pkg.ErrAttr(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}