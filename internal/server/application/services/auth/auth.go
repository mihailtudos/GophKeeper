package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/mihailtudos/gophkeeper/internal/domain"
	"github.com/mihailtudos/gophkeeper/internal/server/config"
	repositories2 "github.com/mihailtudos/gophkeeper/internal/server/infrastructure/repositories"
	"github.com/mihailtudos/gophkeeper/pkg/logger"
	"log/slog"
	"time"

	"github.com/google/uuid"
	tokens "github.com/mihailtudos/gophkeeper/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenExpired       = errors.New("token is expired")
)

// LoginRequest is the login payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest is the register payload
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JWTAuthTokens is the response from authentication operations
type JWTAuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
type AuthService struct {
	cfg    config.Auth
	logger *slog.Logger
	repo   *repositories2.Repository
}

func NewAuthService(ctx context.Context, cfg config.Auth, logger *slog.Logger, repo *repositories2.Repository) *AuthService {
	return &AuthService{
		cfg:    cfg,
		logger: logger,
		repo:   repo,
	}
}

func (s *AuthService) CreateAccessToken(ctx context.Context, userID string) (string, error) {
	op := "auth.CreateAccessToken"

	token, err := tokens.NewToken(ctx, s.cfg.SecretKey, userID, s.cfg.AccessTokenExpiry)
	if err != nil {
		s.logger.ErrorContext(ctx, op, logger.ErrAttr(err))
		return "", fmt.Errorf("%s.NewToken: %w", op, err)
	}

	return token, nil
}

func (s *AuthService) CreateAndStoreRefreshToken(ctx context.Context, userID string) (*repositories2.RefreshToken, error) {
	op := "auth.CreateAndStoreRefreshToken"

	token, err := tokens.NewToken(ctx, s.cfg.SecretKey, userID, s.cfg.RefreshTokenExpiry)
	if err != nil {
		s.logger.Error("failed to create refresh token", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.sign: %w", op, err)
	}

	refreshToken := repositories2.RefreshToken{
		ID:        uuid.New().String(),
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiry),
	}

	if err := s.repo.TokenRepository.Create(ctx, refreshToken); err != nil {
		s.logger.Error("failed to create refresh token", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.create: %w", op, err)
	}

	return &refreshToken, nil
}

func (s *AuthService) Register(ctx context.Context, regReq RegisterRequest) (*JWTAuthTokens, error) {
	op := "auth.Register"
	s.logger.Info("registering user", slog.String("email", regReq.Username))

	// TODO: validate user input
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(regReq.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user := domain.User{
		ID:       uuid.New().String(),
		Email:    regReq.Username,
		Password: string(hashedPassword),
	}

	err = s.repo.UserRepository.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repositories2.ErrUniqueConstraintViolation) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserAlreadyExists)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	token, err := s.CreateAccessToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s.accessToken: %w", op, err)
	}

	refreshToken, err := s.CreateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s.refreshToken: %w", op, err)
	}

	return &JWTAuthTokens{
		AccessToken:  token,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    int(s.cfg.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) GetUserByToken(ctx context.Context, token string) (string, error) {
	return "", nil
}

func (s *AuthService) GetAuthToken(ctx context.Context, token string) (string, error) {
	return "", nil
}

func (s *AuthService) Login(ctx context.Context, lr LoginRequest) (*JWTAuthTokens, error) {
	op := "auth.Login"
	s.logger.Debug("logging in user", slog.String("username", lr.Username))

	user, err := s.repo.UserRepository.GetByUsername(ctx, lr.Username)
	if err != nil {
		if errors.Is(err, repositories2.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s.get: %w", op, ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s.get: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(lr.Password)); err != nil {
		return nil, fmt.Errorf("%s.compare: %w", op, ErrInvalidCredentials)
	}

	accessToken, err := s.CreateAccessToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s.accessToken: %w", op, err)
	}

	refreshToken, err := s.CreateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%s.refreshToken: %w", op, err)
	}

	tokens := &JWTAuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    int(s.cfg.AccessTokenExpiry.Seconds()),
	}

	return tokens, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*JWTAuthTokens, error) {
	op := "auth.Refresh"

	rt, err := s.repo.TokenRepository.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, repositories2.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s.get: %w", op, ErrTokenNotFound)
		}

		return nil, fmt.Errorf("%s.get: %w", op, err)
	}

	// Check if token is expired or revoked
	if rt.ExpiresAt.Before(time.Now()) || rt.Revoked {
		return nil, fmt.Errorf("%s.expired: %w", op, ErrTokenExpired)
	}

	// Generate new access token
	accessToken, err := s.CreateAccessToken(ctx, rt.UserID)
	if err != nil {
		s.logger.Error("failed to create access token", logger.ErrAttr(err))
		return nil, fmt.Errorf("%s.accessToken: %w", op, err)
	}

	return &JWTAuthTokens{
		AccessToken:  accessToken,
		RefreshToken: rt.Token,
		ExpiresIn:    int(s.cfg.AccessTokenExpiry.Seconds()),
	}, nil
}
