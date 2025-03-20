package keyring

import "github.com/zalando/go-keyring"

func GetSecretKey(key, user string) (string, error) {
	token, err := keyring.Get(key, user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func StoreAuthKey(key, user, value string) error {
	err := keyring.Set(key, user, value)
	if err != nil {
		return err
	}

	return nil
}

func RemoveAuthKey(key, user string) error {
	err := keyring.Delete(key, user)
	if err != nil {
		return err
	}
	return nil
}
