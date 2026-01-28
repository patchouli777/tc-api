package mw

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	"twitchy-api/internal/lib/handler"

	"github.com/google/uuid"
)

func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, req)

			ctx := req.Context()

			id, ok := ReqIdFromContext(ctx)
			if !ok {
				log.Error("request id big bad")
			}

			log.LogAttrs(ctx, slog.LevelInfo, "request",
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.String("done_in", time.Since(start).String()),
				slog.String("request_id", id))
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
					fmt.Println("----------------------------")
					debug.PrintStack()
					fmt.Println("----------------------------")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(handler.ErrorResponse{Success: false, Errors: map[string]string{"unknown": "internal error"}})
					return
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func ReqIdFromContext(ctx context.Context) (string, bool) {
	id := ctx.Value(RequestIDContextKey{})
	idStr, ok := id.(string)
	return idStr, ok
}

type RequestIDContextKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("tc-Request-ID")
		if id == "" {
			id = uuid.New().String()
			r.Header.Set("tc-Request-ID", id)
		}

		ctx := context.WithValue(r.Context(), RequestIDContextKey{}, id)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
