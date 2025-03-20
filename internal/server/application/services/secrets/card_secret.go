package secrets

import (
	"encoding/json"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"github.com/mihailtudos/gophkeeper/pkg/encrypt"
	"github.com/mihailtudos/gophkeeper/pkg/errors"
)

type CardSecretStrategy struct{}

func (c *CardSecretStrategy) Validate(secret json.RawMessage) error {
	op := "secret.strategy.validation"

	var creditCard domain.CardDetails
	if err := json.Unmarshal(secret, &creditCard); err != nil {
		return errors.WrapStandardError(op, "failed json unmarshal", err)
	}

	if creditCard.CardHolder == "" || creditCard.CardNumber == "" ||
		creditCard.ExpirationDate == "" || creditCard.CVV == "" {
		return errors.WrapStandardError(op, "failed validation", domain.ErrInvalidSecret)
	}

	return nil
}

func (c *CardSecretStrategy) CalculateCheckSum(secret json.RawMessage) []byte {
	return encrypt.ComputeChecksum(secret)
}

func (c *CardSecretStrategy) EncryptSecret(masterPassword string, userSalt []byte, secret json.RawMessage) ([]byte, []byte, error) {
	return encrypt.Encrypt(masterPassword, userSalt, secret)
}
