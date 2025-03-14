package middleware

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mihailtudos/gophkeeper/server/internal/application/services/auth"
	"github.com/mihailtudos/gophkeeper/server/internal/pkg"
)

func Auth(secret string, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				log.Debug("Authorization header required")
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[7:]

			claims := auth.Claims{}

			token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				log.Error("Invalid token", pkg.ErrAttr(err))
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			r.Header.Set("user_id", claims.UserID)

			next.ServeHTTP(w, r)
		})
	}
}
