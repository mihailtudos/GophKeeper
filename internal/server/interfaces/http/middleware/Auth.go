package middleware

import (
	"github.com/mihailtudos/gophkeeper/pkg/logger"
	"log/slog"
	"net/http"

	tokens "github.com/mihailtudos/gophkeeper/pkg/jwt"
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

			claims, err := tokens.ParseToken(r.Context(), tokenString, secret)
			if err != nil {
				log.Error("failed to parse token", logger.ErrAttr(err))
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			r.Header.Set("user_id", claims.UserID)

			next.ServeHTTP(w, r)
		})
	}
}
