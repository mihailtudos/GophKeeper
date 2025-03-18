package services

import (
	"context"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
	"github.com/mihailtudos/gophkeeper/internal/client/dto"
	"log/slog"
)

type AuthServiceProvider interface {
	Login(ctx context.Context, username, password string) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error)
	StoreTokens(ctx context.Context, response *dto.LoginResponse) error
	GetAccessToken(ctx context.Context) (string, error)
	StoreBackupKey(ctx context.Context, key string) error
}

type SecretsServiceProvider interface {
	CreateSecret(ctx context.Context, message dto.SecretMessage) error
}

type Services struct {
	Logger         *slog.Logger
	Config         *config.Config
	AuthService    AuthServiceProvider
	SecretsService SecretsServiceProvider
}

func NewServices(ctx context.Context, l *slog.Logger, cfg *config.Config) *Services {
	as := NewAuthService(ctx, l, cfg)
	return &Services{
		Logger:         l,
		Config:         cfg,
		AuthService:    as,
		SecretsService: NewSecretsService(ctx, as, l, cfg),
	}
}
