package auth

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/lib/handler"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func FromContext(ctx context.Context) (*Claims, bool) {
	claims := ctx.Value(AuthContextKey{})
	casted, ok := claims.(*Claims)
	return casted, ok
}

func AuthMiddleware(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	const op = "auth middleware"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			handler.Error(log, w, op, fmt.Errorf("missing Authorization header"),
				http.StatusUnauthorized, "missing Authorization header")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JWTKey, nil
		})

		if err != nil || !token.Valid {
			handler.Error(log, w, op, err, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), AuthContextKey{}, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func AuthMiddlewareMock(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := &Claims{
			Id:       0,
			Username: "admin",
			Role:     "staff",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "test-user-id",
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		ctx := context.WithValue(r.Context(), AuthContextKey{}, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
