package secrets

import "encoding/json"

type SecretStrategy interface {
	Validate(secret json.RawMessage) error
	CalculateCheckSum(secret json.RawMessage) []byte
	EncryptSecret(masterPassword string, userSalt []byte, secret json.RawMessage) ([]byte, []byte, error)
}
