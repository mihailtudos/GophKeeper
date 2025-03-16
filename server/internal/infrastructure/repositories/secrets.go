package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mihailtudos/gophkeeper/server/internal/domain"
	"github.com/mihailtudos/gophkeeper/server/internal/pkg"
)

type SecretsRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSecretRepository(ctx context.Context, db *sql.DB, logger *slog.Logger) *SecretsRepo {
	return &SecretsRepo{
		db:     db,
		logger: logger,
	}
}

func (sr *SecretsRepo) Create(ctx context.Context, secret domain.Secret) error {
	op := "repositories.SecretsRepo.Create"

	query := `INSERT INTO user_secrets (id, user_id, s_type, s_name, encrypted_data, iv, checksum, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := sr.db.ExecContext(ctx, query, secret.ID, secret.UserID, secret.SType, secret.SName, secret.Data, secret.IV, secret.SumCheck, secret.CreatedAt, secret.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("%s: %w", op, ErrUniqueConstraintViolation)
		}

		sr.logger.ErrorContext(ctx, "failed to create secret", pkg.ErrAttr(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (sr *SecretsRepo) GetUserSecrets(ctx context.Context, userID string) (*[]domain.Secret, error) {
	op := "repositories.SecretsRepo.GetUserSecrets"

	query := `SELECT id, user_id, s_type, s_name, encrypted_data, iv, created_at, updated_at FROM user_secrets WHERE user_id = $1`

	rows, err := sr.db.QueryContext(ctx, query, userID)
	if err != nil {
		sr.logger.ErrorContext(ctx, "failed to get user secrets", pkg.ErrAttr(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	secrets := make([]domain.Secret, 0)
	for rows.Next() {
		var secret domain.Secret
		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secret.SType,
			&secret.SName,
			&secret.Data,
			&secret.IV,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		)
		if err != nil {
			sr.logger.ErrorContext(ctx, "failed to scan secret", pkg.ErrAttr(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		secrets = append(secrets, secret)
	}

	if rows.Err() != nil {
		sr.logger.ErrorContext(ctx, "failed to iterate over secrets", pkg.ErrAttr(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	return &secrets, nil
}

func (sr *SecretsRepo) GetSecretByID(ctx context.Context, secretID string) (*domain.Secret, error) {
	op := "repositories.SecretsRepo.GetByID"

	query := `SELECT id, user_id, s_type, s_name, encrypted_data, iv, created_at, updated_at FROM user_secrets WHERE id = $1`

	var secret domain.Secret
	err := sr.db.QueryRowContext(ctx, query, secretID).Scan(&secret.ID, &secret.UserID, &secret.SType, &secret.SName, &secret.Data, &secret.IV, &secret.CreatedAt, &secret.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, ErrRecordNotFound)
		}

		sr.logger.ErrorContext(ctx, "failed to get secret by id", pkg.ErrAttr(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &secret, nil
}
