package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/lib/er"
	"net/http"
	"time"
)

type Service interface {
	Login(ctx context.Context, username, password string) (*TokenPair, error)
	Register(ctx context.Context, email, username, password string) (*TokenPair, error)
}

type Handler struct {
	log *slog.Logger
	as  Service
}

func NewHandler(log *slog.Logger, s Service) *Handler {
	return &Handler{log: log, as: s}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return access token (refresh token set as cookie)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    LoginRequest  true  "Login credentials"
// @Success      200  {object}  LoginResponse  "Access token returned"
// @Failure      400  {string}  string  "Invalid credentials or request data"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /login [post]
func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "logging in"

	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, err, op, "invalid data in the request")
		return
	}

	tokenPair, err := h.as.Login(r.Context(), request.Username, request.Password)
	if err != nil {
		if errors.Is(err, ErrWrongCredentials) {
			w.WriteHeader(http.StatusBadRequest)
			er.HandlerError(h.log, w, err, op, "wrong credentials")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		er.HandlerError(h.log, w, err, op, "internal error")
		return
	}

	tokenCookie := createTokenCookie(string(tokenPair.Refresh), 720)
	http.SetCookie(w, &tokenCookie)

	json.NewEncoder(w).Encode(LoginResponse{Access: string(tokenPair.Access)})
}

// Register godoc
// @Summary      User registration
// @Description  Register new user account and return access token (refresh token set as cookie)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body    RegisterRequest  true  "Registration data"
// @Success      200  {object}  RegisterResponse  "Access token returned"
// @Failure      400  {string}  string  "Invalid request data"
// @Failure      409  {string}  string  "User already exists"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /register [post]
func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "register"

	var request RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, err, op, "invalid data in the request")
		return
	}

	tokenPair, err := h.as.Register(r.Context(), request.Email, request.Username, request.Password)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			er.HandlerError(h.log, w, err, op, "user already exists")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		er.HandlerError(h.log, w, err, op, "internal error")
		return
	}

	tokenCookie := createTokenCookie(string(tokenPair.Refresh), 720)
	http.SetCookie(w, &tokenCookie)

	json.NewEncoder(w).Encode(RegisterResponse{Access: string(tokenPair.Access)})
}

func createTokenCookie(token string, hours int) http.Cookie {
	return http.Cookie{
		Name:     "refreshToken",
		Value:    token,
		Expires:  time.Now().Add(time.Duration(hours) * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}
