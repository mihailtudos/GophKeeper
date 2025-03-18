package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/client/application/security"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
	"github.com/mihailtudos/gophkeeper/internal/client/dto"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Secret struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	SType     string    `json:"s_type"`
	SName     string    `json:"s_name"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SecretService struct {
	logger     *slog.Logger
	cfg        *config.Config
	keyManager security.KeyManagerProvider
}

func NewSecretsService(ctx context.Context, km security.KeyManagerProvider, l *slog.Logger, cfg *config.Config) *SecretService {
	return &SecretService{
		logger:     l,
		cfg:        cfg,
		keyManager: km,
	}
}

func (s *SecretService) CreateSecret(ctx context.Context, message dto.SecretMessage) error {
	accessToken, err := s.keyManager.GetKey(s.cfg.ServiceAccessTokenKey, s.cfg.AppName)
	if err != nil {
		return fmt.Errorf("failed to retrieve access token: %w", err)
	}
	masterPassword, err := s.keyManager.GetKey(s.cfg.ServiceMasterPasswordKey, s.cfg.AppName)
	if err != nil {
		return fmt.Errorf("failed to retrieve master password: %w", err)
	}

	s.logger.Debug("access token", slog.String("token", accessToken), slog.String("master password", masterPassword))

	message.MasterPassword = masterPassword

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	s.logger.Debug("message", slog.Any("message", string(data)))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.HTTPServer.HostUrl()+"/api/secrets", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create secret: %w", errors.New(resp.Status))
	}

	return nil
}

func (s *SecretService) fetchSecretsFromServer(accessToken, masterPassword string) ([]Secret, error) {
	data, _ := json.Marshal(struct {
		MasterPassword string `json:"master_password"`
	}{
		MasterPassword: masterPassword,
	})

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/secrets", s.cfg.HTTPServer.HostUrl()), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch secrets: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var secrets []Secret
	if err = json.Unmarshal(body, &secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}
