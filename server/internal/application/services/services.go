package services

import (
	"context"
	"log/slog"

	"github.com/mihailtudos/gophkeeper/server/internal/application/services/auth"
	"github.com/mihailtudos/gophkeeper/server/internal/config"
	"github.com/mihailtudos/gophkeeper/server/internal/domain"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/repositories"
)

type AuthService interface {
	CreateToken(ctx context.Context, email string) (string, error)
	Register(ctx context.Context, user domain.User) (string, error)
	Authenticate(ctx context.Context, username, password string) (string, error)
	GetAuthToken(ctx context.Context, token string) (string, error)
	GetUserByToken(ctx context.Context, token string) (string, error)
}

type Services struct {
	AuthService AuthService
	
}

func NewServices(ctx context.Context, cfg *config.Config, logger *slog.Logger, repo *repositories.Repository) *Services {
	return &Services{
		AuthService: auth.NewAuthService(ctx, cfg.Auth, logger, repo.UserRepository),
	}
}
