package secrets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"github.com/mihailtudos/gophkeeper/internal/server/infrastructure/repositories"
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

func (s *Service) StoreSecret(ctx context.Context, userID, secretType, secretName, masterPassword string, secret json.RawMessage) error {
	op := "services.SecretsService.Store"

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt", logger.ErrAttr(err))
		return fmt.Errorf("%s.GetUserSalt: %w", op, err)
	}

	var encData []byte
	var nonce []byte
	var sumcheck []byte
	switch secretType {
	case "login":
		var loginRecord domain.LoginSecret
		if err = json.Unmarshal(secret, &loginRecord); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal login secret", logger.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalLoginSecret: %w", op, err)
		}

		if loginRecord.Login == "" || loginRecord.Password == "" {
			return fmt.Errorf("%s.LoginOrPasswordIsEmpty: %w", op, ErrInvalidSecretType)
		}

		sumcheck = encrypt.ComputeChecksum(secret)

		encData, nonce, err = encrypt.Encrypt(masterPassword, userSalt, secret)
	case "card":
		var card domain.CardDetails
		if err = json.Unmarshal(secret, &card); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal card secret", logger.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalCardSecret: %w", op, err)
		}

		if card.CardHolder == "" || card.CardNumber == "" || card.ExpirationDate == "" || card.CVV == "" {
			return fmt.Errorf("%s.CardHolderIsEmpty: %w", op, ErrInvalidSecretType)
		}

		sumcheck = encrypt.ComputeChecksum(secret)

		encData, nonce, err = encrypt.Encrypt(masterPassword, userSalt, secret)
	case "text":
		var text domain.PlainText
		if err = json.Unmarshal(secret, &text); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal text secret", logger.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalTextSecret: %w", op, err)
		}

		if text.Value == "" {
			return fmt.Errorf("%s.TextIsEmpty: %w", op, ErrInvalidSecretType)
		}

		sumcheck = encrypt.ComputeChecksum(secret)

		encData, nonce, err = encrypt.Encrypt(masterPassword, userSalt, secret)
	case "binary":
		var base64String string
		var binaryData domain.BinaryData
		// Decode Base64 string from JSON
		if err = json.Unmarshal(secret, &base64String); err != nil {
			s.logger.Error("failed to unmarshal base64 string", logger.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalBase64String: %w", op, err)
		}

		// Convert Base64 to raw bytes
		binaryData.Data, err = base64.StdEncoding.DecodeString(base64String)
		if err != nil {
			s.logger.Error("failed to decode base64 data", logger.ErrAttr(err))
			return fmt.Errorf("%s.DecodeBase64: %w", op, err)
		}

		// Compute checksum for raw bytes
		sumcheck = encrypt.ComputeChecksum(binaryData.Data)

		encData, nonce, err = encrypt.Encrypt(masterPassword, userSalt, binaryData.Data)
	default:
		return fmt.Errorf("%s: %w", op, ErrInvalidSecretType)
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "failed to encrypt secret", logger.ErrAttr(err))
		return fmt.Errorf("%s.EncryptSecret: %w", op, err)
	}

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
		s.logger.ErrorContext(ctx, "failed to create secret", logger.ErrAttr(err))
		if errors.Is(err, repositories.ErrUniqueConstraintViolation) {
			return fmt.Errorf("%s.UniqueConstraintViolation: %w", op, ErrSecretExists)
		}

		return fmt.Errorf("%s.CreateSecret: %w", op, err)
	}

	return nil
}

func (s *Service) GetSecretByID(ctx context.Context, secretID, masterPassword string) (*domain.Secret, error) {
	op := "services.SecretsService.GetSecretByID"

	secret, err := s.repository.SecretRepository.GetSecretByID(ctx, secretID)
	if err != nil {
		if errors.Is(err, repositories.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s.GetSecretByID: %w", op, ErrSecretNotFound)
		}

		s.logger.ErrorContext(ctx, "failed to get secret by id", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.GetSecretByID: %w", op, err)
	}

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, secret.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.GetUserSalt: %w", op, err)
	}

	decSecret, err := encrypt.Decrypt(masterPassword, userSalt, secret.Data, secret.IV)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to decrypt secret", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.DecryptSecret: %w", op, err)
	}

	if decSecret == nil {
		s.logger.ErrorContext(ctx, "decrypted secret is null")
		return nil, fmt.Errorf("%s: %w", op, ErrDecryptionFailed)
	}

	secret.Data = decSecret

	return secret, nil
}

func (s *Service) GetUserSecrets(ctx context.Context, userID, masterPassword string) (*[]domain.Secret, error) {
	op := "services.SecretsService.GetUserSecrets"

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.GetUserSalt: %w", op, err)
	}

	secrets, err := s.repository.SecretRepository.GetUserSecrets(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user secrets", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.GetUserSecrets: %w", op, err)
	}

	for i, secret := range *secrets {
		var decSecret []byte
		decSecret, err = encrypt.Decrypt(masterPassword, userSalt, secret.Data, secret.IV)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to decrypt secret", logger.ErrAttr(err))
			return nil, fmt.Errorf("%s.DecryptSecret: %w", op, err)
		}

		if decSecret == nil {
			s.logger.ErrorContext(ctx, "decrypted secret is null")
			return nil, fmt.Errorf("%s: %w", op, ErrDecryptionFailed)
		}

		(*secrets)[i].Data = decSecret
	}

	return secrets, nil
}
