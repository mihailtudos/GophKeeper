package services

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/mihailtudos/gophkeeper/server/internal/application/services/auth"
	"github.com/mihailtudos/gophkeeper/server/internal/application/services/secrets"
	"github.com/mihailtudos/gophkeeper/server/internal/config"
	"github.com/mihailtudos/gophkeeper/server/internal/domain"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/repositories"
)

type AuthService interface {
	CreateAccessToken(ctx context.Context, email string) (string, error)
	CreateAndStoreRefreshToken(ctx context.Context, userID string) (*repositories.RefreshToken, error)
	Register(ctx context.Context, registrationReq auth.RegisterRequest) (*auth.JWTAuthTokens, error)
	Login(ctx context.Context, lr auth.LoginRequest) (*auth.JWTAuthTokens, error)
	GetAuthToken(ctx context.Context, token string) (string, error)
	GetUserByToken(ctx context.Context, token string) (string, error)
}

type SecretsService interface {
	StoreSecret(ctx context.Context, userID, secretType, secretName, masterPassword string, secret json.RawMessage) error
	GetSecretByID(ctx context.Context, secretID, masterPassword string) (*domain.Secret, error)
	GetUserSecrets(ctx context.Context, id string) ([]*domain.Secret, error)
}
type Services struct {
	AuthService    AuthService
	SecretsService SecretsService
}

func NewServices(ctx context.Context, cfg *config.Config, logger *slog.Logger, repo *repositories.Repository) *Services {
	return &Services{
		AuthService:    auth.NewAuthService(ctx, cfg.Auth, logger, repo),
		SecretsService: secrets.NewSecretsService(ctx, logger, repo),
	}
}
