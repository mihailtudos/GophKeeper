package secrets

import (
	"encoding/base64"
	"encoding/json"
	"github.com/mihailtudos/gophkeeper/pkg/encrypt"
	"github.com/mihailtudos/gophkeeper/pkg/errors"
)

// BinarySecretStrategy handles binary secrets
type BinarySecretStrategy struct{}

func (b *BinarySecretStrategy) Validate(secret json.RawMessage) error {
	op := "secret.strategy.validation"

	var base64String string
	if err := json.Unmarshal(secret, &base64String); err != nil {
		return errors.WrapStandardError(op, "failed json unmarshal", err)
	}

	_, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return errors.WrapStandardError(op, "invalid base64 encoding", ErrInvalidSecretType)
	}

	return nil
}

func (b *BinarySecretStrategy) CalculateCheckSum(secret json.RawMessage) []byte {
	data, _ := base64.StdEncoding.DecodeString(string(secret))
	return encrypt.ComputeChecksum(data)
}

func (b *BinarySecretStrategy) EncryptSecret(masterPassword string, userSalt []byte, secret json.RawMessage) ([]byte, []byte, error) {
	op := "secret.strategy.encrypt"
	data, err := base64.StdEncoding.DecodeString(string(secret))
	if err != nil {
		return nil, nil, errors.WrapStandardError(op, "invalid base64 encoding", ErrInvalidSecretType)
	}

	return encrypt.Encrypt(masterPassword, userSalt, data)
}
