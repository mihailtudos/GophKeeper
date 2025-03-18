package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
	"github.com/mihailtudos/gophkeeper/internal/client/dto"
	"github.com/mihailtudos/gophkeeper/pkg/keyring"
	"io"
	"log/slog"
	"net/http"
)

type AuthService struct {
	logger *slog.Logger
	config *config.Config
}

func NewAuthService(ctx context.Context, logger *slog.Logger, cfg *config.Config) *AuthService {
	return &AuthService{
		logger: logger,
		config: cfg,
	}
}

func (a *AuthService) refreshToken(refreshToken string) (*dto.RefreshTokenResponse, error) {
	data, _ := json.Marshal(struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: refreshToken,
	})

	req, err := http.NewRequest(http.MethodPost, a.config.HTTPServer.HostUrl()+"/api/refresh", bytes.NewReader(data))
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

	var refreshTokenResp dto.RefreshTokenResponse
	if err = json.Unmarshal(body, &refreshTokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &refreshTokenResp, nil
}

func (a *AuthService) cleanLoginCreds() {
	_ = keyring.StoreAuthCreds(a.config.ServiceAccessTokenKey, a.config.AppName, "")
	_ = keyring.StoreAuthCreds(a.config.ServiceRefreshTokenKey, a.config.AppName, "")
}

func (a *AuthService) setNewAccessToken() error {
	rt, err := keyring.GetAuthCreds(a.config.ServiceRefreshTokenKey, a.config.AppName)
	a.logger.Debug("exising refresh token", slog.String("refresh_token", rt))
	if err != nil || rt == "" {
		a.cleanLoginCreds()
		return fmt.Errorf("failed to get refresh token: %w", err)
	}

	accessTokenResp, errAuth := a.refreshToken(rt)
	if errAuth != nil {
		a.cleanLoginCreds()
		return fmt.Errorf("failed to refresh token: %w", errAuth)
	} else {
		if err = keyring.StoreAuthCreds(a.config.ServiceAccessTokenKey, a.config.AppName, accessTokenResp.AccessToken); err != nil {
			a.cleanLoginCreds()
			return fmt.Errorf("failed to store access token: %w", err)
		}
	}

	return nil
}

func (a *AuthService) StoreTokens(ctx context.Context, response *dto.LoginResponse) error {
	if err := keyring.StoreAuthCreds(a.config.ServiceAccessTokenKey, a.config.AppName, response.AccessToken); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}
	if err := keyring.StoreAuthCreds(a.config.ServiceRefreshTokenKey, a.config.AppName, response.RefreshToken); err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

func (a *AuthService) GetAccessToken(ctx context.Context) (string, error) {
	token, err := keyring.GetAuthCreds(a.config.ServiceAccessTokenKey, a.config.AppName)
	if err != nil || token == "" {
		if err = a.setNewAccessToken(); err != nil {
			return "", fmt.Errorf("failed to set new access token: %w", err)
		}
	}

	token, err = keyring.GetAuthCreds(a.config.ServiceAccessTokenKey, a.config.AppName)
	if err != nil || token == "" {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	return token, nil
}

func (a *AuthService) Login(ctx context.Context, username, password string) (*dto.LoginResponse, error) {
	op := "client.services.AuthService.Login"
	log := a.logger.With(
		slog.String("op", op),
	)

	data, _ := json.Marshal(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	})

	req, err := http.NewRequest(http.MethodPost, a.config.HTTPServer.HostUrl()+"/api/login", bytes.NewReader(data))
	if err != nil {
		log.Error("failed to create login http request form", slog.String("err", err.Error()))
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

	var loginResp dto.LoginResponse
	if err = json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &loginResp, nil
}

func (a *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	op := "client.services.AuthService.Register"
	log := a.logger.With(
		slog.String("op", op),
	)

	log.Debug("refreshing token")

	return nil, nil
}

func (a *AuthService) StoreBackupKey(ctx context.Context, key string) error {
	op := "client.services.AuthService.StoreBackupKey"
	log := a.logger.With(
		slog.String("op", op),
	)

	log.Debug("storing backup key")

	return keyring.StoreAuthCreds(a.config.ServiceMasterPasswordKey, a.config.AppName, key)
}
