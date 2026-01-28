package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
	d "twitchy-api/internal/auth/domain"
	"twitchy-api/internal/lib/handler"
	"twitchy-api/internal/lib/sl"
	api "twitchy-api/pkg/api/auth"
)

type Service interface {
	SignIn(ctx context.Context, username, password string) (*d.TokenPair, error)
	SignUp(ctx context.Context, email, username, password string) (*d.TokenPair, error)
}

type Handler struct {
	log *slog.Logger
	as  Service
}

func NewHandler(log *slog.Logger, s Service) *Handler {
	return &Handler{log: log, as: s}
}

// SignIn godoc
//
//	@Summary		Sign in user
//	@Description	Authenticate user and return access token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		api.LoginRequest		true	"Login credentials"
//	@Success		200		{object}	api.LoginResponse		"Access token"
//	@Failure		400		{object}	handler.ErrorResponse	"Invalid request or credentials"
//	@Failure		500		{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/signin [post]
func (h Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	const op = "signing in"

	var request api.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	tokenPair, err := h.as.SignIn(r.Context(), request.Username, request.Password)
	if err != nil {
		if errors.Is(err, d.ErrWrongCredentials) {
			handler.Error(h.log, w, op, err, http.StatusBadRequest, d.ErrWrongCredentials.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	tokenCookie := createTokenCookie(string(tokenPair.Refresh), 720)
	http.SetCookie(w, &tokenCookie)

	err = json.NewEncoder(w).Encode(api.LoginResponse{Access: string(tokenPair.Access)})
	if err != nil {

	}
}

// SignUp godoc
//
//	@Summary		Sign up new user
//	@Description	Register new user and return access token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		api.RegisterRequest		true	"Registration data"
//	@Success		200		{object}	api.RegisterResponse	"Access token"
//	@Failure		400		{object}	handler.ErrorResponse	"Invalid request"
//	@Failure		409		{object}	handler.ErrorResponse	"User already exists"
//	@Failure		500		{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/signup [post]
func (h Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	const op = "signing up"

	var request api.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	tokenPair, err := h.as.SignUp(r.Context(), request.Email, request.Username, request.Password)
	if err != nil {
		h.log.Error(op, sl.Err(err))

		if errors.Is(err, d.ErrAlreadyExists) {
			handler.Error(h.log, w, op, err, http.StatusConflict, d.ErrAlreadyExists.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	tokenCookie := createTokenCookie(string(tokenPair.Refresh), 720)
	http.SetCookie(w, &tokenCookie)

	json.NewEncoder(w).Encode(api.RegisterResponse{Access: string(tokenPair.Access)})
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
