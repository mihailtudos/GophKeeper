package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zalando/go-keyring"
	"io"
	"log/slog"
	"net/http"
)

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

const (
	serverAddress            = "http://localhost:8080"
	serviceAccessTokenKey    = "GophKeeperAccessToken"
	serviceRefreshTokenKey   = "GophKeeperRefreshToken"
	serviceMasterPasswordKey = "GophKeeperMasterPassword"
	user                     = "api-key"
)

func storeAuthCreds(key string, token string) error {
	err := keyring.Set(key, user, token)
	if err != nil {
		return err
	}

	return nil
}

func getAuthCreds(key string) (string, error) {
	token, err := keyring.Get(key, user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func refreshToken(refreshToken string) (*RefreshTokenResponse, error) {
	data, _ := json.Marshal(struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: refreshToken,
	})

	req, err := http.NewRequest(http.MethodPost, serverAddress+"/api/refresh", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create login http request form: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var refreshTokenResp RefreshTokenResponse
	if err = json.Unmarshal(body, &refreshTokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &refreshTokenResp, nil
}

func cleanLoginCreds() {
	_ = storeAuthCreds(serviceRefreshTokenKey, "")
	_ = storeAuthCreds(serviceAccessTokenKey, "")
}

func setNewAccessToken() error {
	rt, err := getAuthCreds(serviceRefreshTokenKey)
	logger.Debug("exising refresh token", slog.String("refresh_token", rt))
	if err != nil || rt == "" {
		cleanLoginCreds()
		return fmt.Errorf("failed to get refresh token: %w", err)
	}

	accessTokenResp, errAuth := refreshToken(rt)
	if errAuth != nil {
		cleanLoginCreds()
		return fmt.Errorf("failed to refresh token: %w", errAuth)
	} else {
		if err = storeAuthCreds(serviceAccessTokenKey, accessTokenResp.AccessToken); err != nil {
			cleanLoginCreds()
			return fmt.Errorf("failed to store access token: %w", err)
		}
	}

	return nil
}

func handleLogin(f loginForm) (*LoginResponse, error) {
	data, _ := json.Marshal(f)

	req, err := http.NewRequest(http.MethodPost, serverAddress+"/api/login", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create login http request form: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var loginResp LoginResponse
	if err = json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &loginResp, nil
}
