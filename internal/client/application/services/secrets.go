package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
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
	logger *slog.Logger
	cfg    *config.Config
}

func NewSecretsService(ctx context.Context, l *slog.Logger, cfg *config.Config) *SecretService {
	return &SecretService{
		logger: l,
		cfg:    cfg,
	}
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
