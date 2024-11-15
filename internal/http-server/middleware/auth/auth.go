package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/exp/slog"
	"net/http"
	"shortener/internal/lib/logger/sl"
	"strings"
)

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrFailedIsAdminCheck = errors.New("failed to check if user is admin")
)

type PermissionProvider interface {
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

// New creates new auth middleware.
func New(
	log *slog.Logger,
	appSecret string,
	permProvider PermissionProvider,
) func(next http.Handler) http.Handler {
	const op = "middleware.auth.New"

	log = log.With(slog.String("op", op))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				// It's ok, if user is not authorized
				next.ServeHTTP(w, r)
				return
			}

			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(appSecret), nil
			})
			if err != nil {
				log.Warn("failed to parse token", sl.Err(err))

				// But if token is invalid, we shouldn't handle request
				ctx := context.WithValue(r.Context(), "error", ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			log.Info("user authorized", slog.Any("claims", claims))

			uid, ok := claims["uid"].(float64)
			if !ok {
				log.Warn("uid is missing or invalid in claims")
				ctx := context.WithValue(r.Context(), "error", ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			isAdmin, err := permProvider.IsAdmin(r.Context(), int64(uid))
			if err != nil {
				log.Error("failed to check if user is admin", sl.Err(err))

				ctx := context.WithValue(r.Context(), "error", ErrFailedIsAdminCheck)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			ctx := context.WithValue(r.Context(), "uid", int64(uid))
			ctx = context.WithValue(r.Context(), "isAdmin", isAdmin)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}

	return splitToken[1]
}
