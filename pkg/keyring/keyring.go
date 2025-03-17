package keyring

import "github.com/zalando/go-keyring"

func GetAuthCreds(key, user string) (string, error) {
	token, err := keyring.Get(key, user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func StoreAuthCreds(key, user, value string) error {
	err := keyring.Set(key, user, value)
	if err != nil {
		return err
	}

	return nil
}
