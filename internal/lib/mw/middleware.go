package mw

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/lib/handler"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

func Logging(ctx context.Context, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, req)
			log.LogAttrs(ctx, slog.LevelInfo, "request",
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.String("time", time.Since(start).String()))
		})
	}
}

func JSONResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(req.URL.Path, "/static/") {
			w.Header().Set("Content-type", "application/json")
		}

		next.ServeHTTP(w, req)
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func PanicRecovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rcv := recover(); rcv != nil {
					log.Error("recover after panic: %v\n. stack trace:\n%s", rcv, debug.Stack())
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(handler.RequestError{Success: false, Error: "internal error"})
					return
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
