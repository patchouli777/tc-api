package mw

import (
	"encoding/json"
	"log/slog"
	"main/internal/lib/er"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		slog.Info("request",
			"method", req.Method,
			"uri", req.RequestURI,
			"time", time.Since(start),
		)
	})
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

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				slog.Error("Восстановление после паники: %v\n. Стек трейс:\n%s", rcv, debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(er.RequestError{Error: "Внутренняя ошибка"})
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// TODO: implement ratelimit
// func RateLimit(ul *UserLimiter, next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// In a real application, extract the actual user ID from the request.
// 		// For this example, we'll use a dummy user ID.
// 		userID := "user123" // Replace with actual user ID extraction logic

// 		limiter := ul.GetLimiter(userID)
// 		if !limiter.Allow() {
// 			slog.Error("Превышен лимит запросов")
// 			w.WriteHeader(http.StatusTooManyRequests)
// 			json.NewEncoder(w).Encode(er.RequestError{Error: "Превышен лимит запросов"})
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

// type UserLimiter struct {
// 	limiters map[string]*rate.Limiter
// 	mu       sync.Mutex
// 	rate     rate.Limit
// 	burst    int
// }

// func NewUserLimiter(r rate.Limit, b int) *UserLimiter {
// 	return &UserLimiter{
// 		limiters: make(map[string]*rate.Limiter),
// 		rate:     r,
// 		burst:    b,
// 	}
// }

// func (ul *UserLimiter) GetLimiter(userID string) *rate.Limiter {
// 	ul.mu.Lock()
// 	defer ul.mu.Unlock()

// 	limiter, exists := ul.limiters[userID]
// 	if !exists {
// 		limiter = rate.NewLimiter(ul.rate, ul.burst)
// 		ul.limiters[userID] = limiter
// 	}
// 	return limiter
// }
