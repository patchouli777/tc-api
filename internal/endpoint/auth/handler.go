package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/lib/handler"
	"main/internal/lib/sl"
	"net/http"
	"time"
)

type Service interface {
	SignIn(ctx context.Context, username, password string) (*TokenPair, error)
	SignUp(ctx context.Context, email, username, password string) (*TokenPair, error)
}

type Handler struct {
	log *slog.Logger
	as  Service
}

func NewHandler(log *slog.Logger, s Service) *Handler {
	return &Handler{log: log, as: s}
}

// TODO: update doc
func (h Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	const op = "signing in"

	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	tokenPair, err := h.as.SignIn(r.Context(), request.Username, request.Password)
	if err != nil {
		h.log.Error(op, sl.Err(err))

		if errors.Is(err, errWrongCredentials) {
			handler.Error(h.log, w, op, err, http.StatusBadRequest, errWrongCredentials.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	tokenCookie := createTokenCookie(string(tokenPair.Refresh), 720)
	http.SetCookie(w, &tokenCookie)

	json.NewEncoder(w).Encode(LoginResponse{Access: string(tokenPair.Access)})
}

// TODO: update doc
func (h Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	const op = "signing up"

	var request RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	tokenPair, err := h.as.SignUp(r.Context(), request.Email, request.Username, request.Password)
	if err != nil {
		h.log.Error(op, sl.Err(err))

		if errors.Is(err, errUserAlreadyExists) {
			handler.Error(h.log, w, op, err, http.StatusConflict, errUserAlreadyExists.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
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
