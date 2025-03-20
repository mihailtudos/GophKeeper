package jwt

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewToken(ctx context.Context, secretKey, userID string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
	})

	return token.SignedString([]byte(secretKey))
}

func ParseToken(ctx context.Context, tokenString, secretKey string) (*Claims, error) {
	claims := Claims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}
