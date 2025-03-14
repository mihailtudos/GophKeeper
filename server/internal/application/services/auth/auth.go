package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mihailtudos/gophkeeper/server/internal/config"
	"github.com/mihailtudos/gophkeeper/server/internal/domain"
	"github.com/mihailtudos/gophkeeper/server/internal/infrastructure/repositories"
	"github.com/mihailtudos/gophkeeper/server/internal/pkg"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type Claims struct {
    UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

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
	repo   *repositories.Repository
}

func NewAuthService(ctx context.Context, cfg config.Auth, logger *slog.Logger, repo *repositories.Repository) *AuthService {
	return &AuthService{
		cfg:    cfg,
		logger: logger,
		repo:   repo,
	}
}

func (s *AuthService) CreateAccessToken(ctx context.Context, userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.cfg.AccessTokenExpiry).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) CreateAndStoreRefreshToken(ctx context.Context, userID string) (*repositories.RefreshToken, error) {
	op := "auth.CreateAndStoreRefreshToken"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(s.cfg.RefreshTokenExpiry).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		s.logger.Error("failed to sign refresh token", pkg.ErrAttr(err))
		return nil, fmt.Errorf("%s.sign: %w", op, err)
	}

	refreshToken := repositories.RefreshToken{
		ID:        uuid.New().String(),
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiry),
	}

	if err := s.repo.TokenRepository.Create(ctx, refreshToken); err != nil {
		s.logger.Error("failed to create refresh token", pkg.ErrAttr(err))
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
		if errors.Is(err, repositories.ErrUniqueConstraintViolation) {
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
		if errors.Is(err, repositories.ErrRecordNotFound) {
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
