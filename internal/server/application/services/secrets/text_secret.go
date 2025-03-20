package secrets

import (
	"encoding/json"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"github.com/mihailtudos/gophkeeper/pkg/encrypt"
	"github.com/mihailtudos/gophkeeper/pkg/errors"
)

type TextSecretStrategy struct{}

func (t *TextSecretStrategy) Validate(secret json.RawMessage) error {
	op := "secret.strategy.validation"

	var text domain.PlainText
	if err := json.Unmarshal(secret, &text); err != nil {
		return errors.WrapStandardError(op, "failed json unmarshal", err)
	}

	if text.Value == "" {
		return errors.WrapStandardError(op, "failed validation", domain.ErrInvalidSecret)
	}

	return nil
}

func (t *TextSecretStrategy) CalculateCheckSum(secret json.RawMessage) []byte {
	return encrypt.ComputeChecksum(secret)
}

func (t *TextSecretStrategy) EncryptSecret(masterPassword string, userSalt []byte, secret json.RawMessage) ([]byte, []byte, error) {
	return encrypt.Encrypt(masterPassword, userSalt, secret)
}
