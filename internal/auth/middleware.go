package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/lib/er"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	const op = "auth.middleware"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(er.RequestError{Error: "missing Authorization header"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JWTKey, nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			er.HandlerError(log, w, err, op, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), AuthContextKey{}, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
