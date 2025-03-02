package auth

import (
	"context"
	"crypto/bcrypt"
	"fmt"
	"log/slog"
	"time"

	"github.com/mihailtudos/gophkeeper/server/internal/config"
	"github.com/mihailtudos/gophkeeper/server/internal/domain"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/go-playground/validator/v10"
)

type AuthService struct {
	cfg    config.Auth
	logger *slog.Logger
	repo   repositories.UserRepository
}

func NewAuthService(ctx context.Context, cfg config.Auth, logger *slog.Logger, ur repositories.UserRepository) *AuthService {
	return &AuthService{
		cfg:    cfg,
		logger: logger,
		repo:   ur,
	}
}

func (s *AuthService) CreateToken(ctx context.Context, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) Register(ctx context.Context, user domain.User) (string, error) {
	op := "auth.Register"
	s.logger.Info("registering user", slog.String("email", user.Email))

	v := validator.New()
	v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= 8 && len(fl.Field().String()) <= 32
	})

	v.StructCtx(ctx, user)


	user.Password = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	err := s.repo.Create(ctx, user)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return s.CreateToken(ctx, user.Email)
}

func (s *AuthService) GetUserByToken(ctx context.Context, token string) (string, error) {
	return "", nil
}

func (s *AuthService) GetAuthToken(ctx context.Context, token string) (string, error) {
	return "", nil
}

func (s *AuthService) Authenticate(ctx context.Context, username, password string) (string, error) {
	return "", nil
}
