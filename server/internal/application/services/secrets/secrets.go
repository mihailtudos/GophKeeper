package secrets

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mihailtudos/gophkeeper/server/internal/domain"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/repositories"
	"github.com/mihailtudos/gophkeeper/server/internal/pkg"
	"golang.org/x/crypto/argon2"
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

// DeriveKey generates a 32-byte encryption key using Argon2id
func DeriveKey(masterPassword string, userSalt []byte) []byte {
	return argon2.IDKey([]byte(masterPassword), userSalt, 1, 64*1024, 4, 32)
}

// encrypt takes a master password, user salt, and plaintext data and returns:
// - ciphertext: the encrypted data as a byte slice
// - nonce: the initialization vector (IV) used for encryption as a byte slice
// - error: any error encountered during encryption
// The function uses AES-GCM for encryption with a key derived from the master password
// and user salt using Argon2id. A random nonce is generated for each encryption operation.
// Example:
//
//	masterPassword := "user-secure-password"
//	userSalt := []byte{...} // Salt retrieved from user's record
//	plaintext := []byte("sensitive data to encrypt")
//
//	ciphertext, nonce, err := encrypt(masterPassword, userSalt, plaintext)
//	if err != nil {
//	    // Handle encryption error
//	}
//
//	// Store ciphertext and nonce in the database
func encrypt(masterPassword string, userSalt, plaintext []byte) ([]byte, []byte, error) {
	key := DeriveKey(masterPassword, userSalt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// Generate a new nonce (IV)
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	// Encrypt data
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func decrypt(masterPassword string, userSalt, ciphertext, nonce []byte) ([]byte, error) {
	key := DeriveKey(masterPassword, userSalt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// ComputeChecksum calculates a SHA-256 hash of the encrypted data
func ComputeChecksum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (s *Service) StoreSecret(ctx context.Context, userID, secretType, secretName, masterPassword string, secret json.RawMessage) error {
	op := "services.SecretsService.Store"

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt", pkg.ErrAttr(err))
		return fmt.Errorf("%s.GetUserSalt: %w", op, err)
	}

	var encData []byte
	var nonce []byte
	var sumcheck []byte
	switch secretType {
	case "login":
		var loginRecord domain.LoginSecret
		if err = json.Unmarshal(secret, &loginRecord); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal login secret", pkg.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalLoginSecret: %w", op, err)
		}

		if loginRecord.Login == "" || loginRecord.Password == "" {
			return fmt.Errorf("%s.LoginOrPasswordIsEmpty: %w", op, ErrInvalidSecretType)
		}

		sumcheck = ComputeChecksum(secret)

		encData, nonce, err = encrypt(masterPassword, userSalt, secret)
	case "card":
		var card domain.CardDetails
		if err = json.Unmarshal(secret, &card); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal card secret", pkg.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalCardSecret: %w", op, err)
		}

		if card.CardHolder == "" || card.CardNumber == "" || card.ExpirationDate == "" || card.CVV == "" {
			return fmt.Errorf("%s.CardHolderIsEmpty: %w", op, ErrInvalidSecretType)
		}

		sumcheck = ComputeChecksum(secret)

		encData, nonce, err = encrypt(masterPassword, userSalt, secret)
	case "text":
		var text domain.PlainText
		if err = json.Unmarshal(secret, &text); err != nil {
			s.logger.ErrorContext(ctx, "failed to unmarshal text secret", pkg.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalTextSecret: %w", op, err)
		}

		if text.Value == "" {
			return fmt.Errorf("%s.TextIsEmpty: %w", op, ErrInvalidSecretType)
		}

		sumcheck = ComputeChecksum(secret)

		encData, nonce, err = encrypt(masterPassword, userSalt, secret)
	case "binary":
		var base64String string
		var binaryData domain.BinaryData
		// Decode Base64 string from JSON
		if err = json.Unmarshal(secret, &base64String); err != nil {
			s.logger.Error("failed to unmarshal base64 string", pkg.ErrAttr(err))
			return fmt.Errorf("%s.UnmarshalBase64String: %w", op, err)
		}

		// Convert Base64 to raw bytes
		binaryData.Data, err = base64.StdEncoding.DecodeString(base64String)
		if err != nil {
			s.logger.Error("failed to decode base64 data", pkg.ErrAttr(err))
			return fmt.Errorf("%s.DecodeBase64: %w", op, err)
		}

		// Compute checksum for raw bytes
		sumcheck = ComputeChecksum(binaryData.Data)

		encData, nonce, err = encrypt(masterPassword, userSalt, binaryData.Data)
	default:
		return fmt.Errorf("%s: %w", op, ErrInvalidSecretType)
	}

	if err != nil {
		s.logger.ErrorContext(ctx, "failed to encrypt secret", pkg.ErrAttr(err))
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
		s.logger.ErrorContext(ctx, "failed to create secret", pkg.ErrAttr(err))
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

		s.logger.ErrorContext(ctx, "failed to get secret by id", pkg.ErrAttr(err))
		return nil, fmt.Errorf("%s.GetSecretByID: %w", op, err)
	}

	userSalt, err := s.repository.UserRepository.GetUserSalt(ctx, secret.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user salt", pkg.ErrAttr(err))
		return nil, fmt.Errorf("%s.GetUserSalt: %w", op, err)
	}

	decSecret, err := decrypt(masterPassword, userSalt, secret.Data, secret.IV)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to decrypt secret", pkg.ErrAttr(err))
		return nil, fmt.Errorf("%s.DecryptSecret: %w", op, err)
	}

	if decSecret == nil {
		s.logger.ErrorContext(ctx, "decrypted secret is null")
		return nil, fmt.Errorf("%s: %w", op, ErrDecryptionFailed)
	}

	secret.Data = decSecret

	return secret, nil
}

func (s *Service) GetUserSecrets(ctx context.Context, id string) ([]*domain.Secret, error) {
	return nil, nil
}
