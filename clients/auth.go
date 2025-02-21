package main

import "github.com/zalando/go-keyring"

const (
	service = "my-cli-tool"
	user    = "api-key"
)

func storeAuthCreds(token string) error {
	err := keyring.Set(service, user, token)
	if err != nil {
		return err
	}

	return nil
}


func getAuthCreds() (string, error) {
	token, err := keyring.Get(service, user)
	if err != nil {
		return "", err
	}

	return token, nil 
}

