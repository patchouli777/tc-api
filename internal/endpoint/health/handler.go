package health

import (
	"log/slog"
	"net/http"
)

func Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		slog.Error("health bad")
	}
}
