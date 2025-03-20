package security

import (
	"github.com/mihailtudos/gophkeeper/pkg/keyring"
)

type KeyManagerProvider interface {
	GetKey(key, user string) (string, error)
	StoreKey(key, user, value string) error
	RemoveKey(key, user string) error
}

type KeyManager struct{}

func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

func (k *KeyManager) GetKey(key, user string) (string, error) {
	return keyring.GetSecretKey(key, user)
}

func (k *KeyManager) StoreKey(key, user, value string) error {
	return keyring.StoreAuthKey(key, user, value)
}

func (k *KeyManager) RemoveKey(key, user string) error {
	return keyring.RemoveAuthKey(key, user)
}
