package secrets

import (
	"encoding/json"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"github.com/mihailtudos/gophkeeper/pkg/encrypt"
	"github.com/mihailtudos/gophkeeper/pkg/errors"
)

type LoginSecretStrategy struct{}

func (s *LoginSecretStrategy) Validate(secret json.RawMessage) error {
	op := "secret.strategy.validation"

	var loginSecret domain.LoginSecret
	if err := json.Unmarshal(secret, &loginSecret); err != nil {
		return errors.WrapStandardError(op, "failed json unmarshal", err)
	}

	if loginSecret.Login == "" || loginSecret.Password == "" {
		return errors.WrapStandardError(op, "failed json unmarshal", domain.ErrInvalidSecret)
	}

	return nil
}

func (s *LoginSecretStrategy) CalculateCheckSum(secret json.RawMessage) []byte {
	return encrypt.ComputeChecksum(secret)
}

func (s *LoginSecretStrategy) EncryptSecret(masterPassword string, userSalt []byte, secret json.RawMessage) ([]byte, []byte, error) {
	return encrypt.Encrypt(masterPassword, userSalt, secret)
}
