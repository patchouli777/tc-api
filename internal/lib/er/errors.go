package er

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"
	"net/http"
)

type RequestError struct {
	Success bool
	Error   string
}

var ErrNotFound = errors.New("item not found")

func HandlerError(log *slog.Logger, w http.ResponseWriter, err error, errOp, errUserInfo string) {
	log.Error(errOp, sl.Err(err))
	json.NewEncoder(w).Encode(RequestError{
		Success: false,
		Error:   fmt.Sprintf("error %s: %s", errOp, errUserInfo),
	})
}
