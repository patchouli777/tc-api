package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"twitchy-api/internal/lib/sl"
)

type ErrorResponse struct {
	Success bool              `json:"success"`
	Errors  map[string]string `json:"errors"`
	When    string            `json:"when"` // felt cute might delete later
}

var (
	ErrNotFound   = errors.New("item not found")
	ErrNotAllowed = errors.New("you are not allowed to perform this action")
	ErrClaims     = errors.New("bad claims")
	ErrIdentity   = errors.New("identity is not confirmed")
	ErrBadSort    = errors.New("bad sort parameter")
	ErrBadPage    = errors.New("bad page parameter: only integers are accepted")
	ErrBadCount   = errors.New("bad count parameter: only integers are accepted")
)

const (
	MsgIdentity = "identity is not confirmed"
	MsgInternal = "internal error"
	MsgRequest  = "invalid data in the request"
)

func Errors(log *slog.Logger, w http.ResponseWriter, op string, status int, errs map[string]error) {
	log.Error(op, slog.Any("errors", errs))

	errsStrs := make(map[string]string)
	for k, v := range errs {
		errsStrs[k] = v.Error()
	}

	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Errors:  errsStrs,
		When:    op,
	})
	if err != nil {
		log.Error("response", sl.Op(op))
	}
}

func Error(log *slog.Logger, w http.ResponseWriter, op string, err error, status int, msg string) {
	log.Error(op, sl.Err(err))

	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		// TODO: handler-specific errors instead of msgs
		Errors: map[string]string{"msg": msg},
		When:   op,
	})
	if err != nil {
		log.Error("sending response", sl.Op(op))
	}
}
