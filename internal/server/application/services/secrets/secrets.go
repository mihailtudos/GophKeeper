package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"github.com/mihailtudos/gophkeeper/internal/server/infrastructure/repositories"
	errorsHandler "github.com/mihailtudos/gophkeeper/pkg/errors"
	"github.com/mihailtudos/gophkeeper/pkg/logger"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mihailtudos/gophkeeper/pkg/encrypt"
)

type Service struct {
	logger     *slog.Logger
	repository *repositories.Repository
}

type ShowSecretRequest struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	MasterPassword string `json:"master_password"`
}

var (
	ErrInvalidSecretType = errors.New("invalid secret type")
	ErrSecretNotFound    = errors.New("secret not found")
	ErrDecryptionFailed  = errors.New("decryption failed")
	ErrSecretExists      = errors.New("secret already exists")
)

func NewSecretsService(ctx context.Context, logger *slog.Logger, repository *repositories.Repository) *Service {
	return &Service{
		logger:     logger,
		repository: repository,
	}
}

func GetSecretStrategy(secretType string) (SecretStrategy, error) {
	switch secretType {
	case "login":
		return &LoginSecretStrategy{}, nil
	case "card":
		return &CardSecretStrategy{}, nil
	case "text":
		return &TextSecretStrategy{}, nil
	case "binary":
		return &BinarySecretStrategy{}, nil
	default:
		return nil, errors.New("invalid secret type")
	}
}

func (s *Service) StoreSecret(ctx context.Context, userID, secretType, secretName, masterPassword string, secret json.RawMessage) error {
	op := "services.SecretsService.Store"

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt",
			slog.String("user_id", userID), logger.ErrAttr(err))
		return errorsHandler.WrapStandardError(op, "failed to get user salt", err)
	}

	var encData []byte
	var nonce []byte
	var sumcheck []byte
	var strategy SecretStrategy

	strategy, err = GetSecretStrategy(secretType)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get secret strategy",
			slog.String("secret_type", secretType), logger.ErrAttr(err))
		return errorsHandler.WrapStandardError(op, "strategy not found", err)
	}

	if err = strategy.Validate(secret); err != nil {
		s.logger.DebugContext(ctx, "failed to validate secret", logger.ErrAttr(err))
		return errorsHandler.WrapStandardError(op, "validation failed", err)
	}

	encData, nonce, err = strategy.EncryptSecret(masterPassword, userSalt, secret)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to encrypt secret", logger.ErrAttr(err))
		return errorsHandler.WrapStandardError(op, "encryption failed", err)
	}

	sumcheck = encrypt.ComputeChecksum(secret)

	secretData := domain.Secret{
		ID:        uuid.New().String(),
		UserID:    userID,
		SType:     secretType,
		SName:     secretName,
		Data:      encData,
		IV:        nonce,
		SumCheck:  sumcheck,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = s.repository.SecretRepository.Create(ctx, secretData); err != nil {
		s.logger.ErrorContext(ctx, "failed to create secret",
			slog.String("secret_type", secretType),
			slog.String("secret_name", secretName),
			slog.String("secret_checksum", fmt.Sprintf("%x", sumcheck)),
			logger.ErrAttr(err))
		if errors.Is(err, repositories.ErrUniqueConstraintViolation) {
			return errorsHandler.WrapStandardError(op, "failed to store secret", ErrSecretExists)
		}

		return errorsHandler.WrapStandardError(op, "failed to store secret", err)
	}

	return nil
}

func (s *Service) GetSecretByID(ctx context.Context, secretID, masterPassword string) (*domain.Secret, error) {
	op := "services.SecretsService.GetSecretByID"

	secret, err := s.repository.SecretRepository.GetSecretByID(ctx, secretID)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return nil, errorsHandler.WrapStandardError(op, "missing secret", ErrSecretNotFound)
		}

		s.logger.ErrorContext(ctx, "failed to get secret by id",
			slog.String("secret_id", secretID),
			logger.ErrAttr(err))
		return nil, errorsHandler.WrapStandardError(op, "failed to find secret", err)
	}

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, secret.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt",
			slog.String("user_id", secret.UserID),
			logger.ErrAttr(err))
		return nil, errorsHandler.WrapStandardError(op, "failed to get user salt", err)
	}

	decSecret, err := encrypt.Decrypt(masterPassword, userSalt, secret.Data, secret.IV)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to decrypt secret",
			slog.String("secret_id", secretID),
			logger.ErrAttr(err))
		return nil, errorsHandler.WrapStandardError(op, "failed secret decryption", err)
	}

	if decSecret == nil {
		s.logger.ErrorContext(ctx, "decrypted secret is null")
		return nil, errorsHandler.WrapStandardError(op, "decrypted secret is nil", ErrDecryptionFailed)
	}

	secret.Data = decSecret

	return secret, nil
}

func (s *Service) GetUserSecrets(ctx context.Context, userID, masterPassword string) (*[]domain.Secret, error) {
	op := "services.SecretsService.GetUserSecrets"

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt",
			slog.String("user_id", userID),
			logger.ErrAttr(err))
		return nil, errorsHandler.WrapStandardError(op, "failed to get user salt", err)
	}

	secrets, err := s.repository.SecretRepository.GetUserSecrets(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user secrets",
			slog.String("user_id", userID),
			logger.ErrAttr(err))
		return nil, errorsHandler.WrapStandardError(op, "failed to get user secrets", err)
	}

	for i, secret := range *secrets {
		var decSecret []byte
		decSecret, err = encrypt.Decrypt(masterPassword, userSalt, secret.Data, secret.IV)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to decrypt secret",
				slog.String("secret_id", secret.ID),
				slog.String("user_id", userID),
				logger.ErrAttr(err))
			continue
		}

		if decSecret == nil {
			s.logger.ErrorContext(ctx, "decrypted secret is null",
				slog.String("secret_id", secret.ID),
				slog.String("user_id", userID))
			continue
		}

		(*secrets)[i].Data = decSecret
	}

	return secrets, nil
}
