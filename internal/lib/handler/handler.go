package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"
	"net/http"
)

type ErrorResponse struct {
	Success bool
	Error   string
}

var (
	ErrNotFound   = errors.New("item not found")
	ErrNotAllowed = errors.New("you are not allowed to perform this action")
	ErrClaims     = errors.New("bad claims")
	ErrIdentity   = errors.New("identity is not confirmed")
	ErrBadSort    = errors.New("incorrect sort parameter")
)

const (
	MsgIdentity = "identity is not confirmed"
	MsgInternal = "internal error"
	MsgRequest  = "invalid data in the request"
	MsgBadPage  = "incorrect page parameter"
	MsgBadSort  = "incorrect sort parameter"
	MsgBadCount = "incorrect count parameter"
)

func Error(log *slog.Logger, w http.ResponseWriter, op string, err error, status int, msg string) {
	log.Error(op, sl.Err(err))

	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error:   fmt.Sprintf("error %s: %s", op, msg),
	})
	if err != nil {
		log.Error("response", sl.Op(op))
	}
}
