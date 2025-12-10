package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"main/internal/lib/er"
	"net/http"
)

type Service interface {
	Get(ctx context.Context, username string) (*Channel, error)
	Update(ctx context.Context, upd ChannelUpdate) error
	// Delete(ctx context.Context)
}

type Handler struct {
	cs  Service
	log *slog.Logger
}

func NewHandler(log *slog.Logger, s Service) *Handler {
	return &Handler{cs: s, log: log}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting channel"

	channel := r.PathValue("channel")

	if channel == "" {
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, fmt.Errorf("channel is not present in the request"), op, "channel is not present in the request")
		return
	}

	chann, err := h.cs.Get(r.Context(), channel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		er.HandlerError(h.log, w, err, op, "internal error")
		return
	}

	json.NewEncoder(w).Encode(GetResponse{
		IsBanned:        chann.IsBanned,
		IsPartner:       chann.IsPartner,
		Background:      chann.Background,
		FirstLivestream: chann.FirstLivestream,
		LastLivestream:  chann.LastLivestream,
		Description:     chann.Description,
		Links:           chann.Links,
		Tags:            chann.Tags,
	})
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating channel"

	var channelUpdate ChannelUpdate
	if err := json.NewDecoder(r.Body).Decode(&channelUpdate); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, err, op, "invalid data in the request")
		return
	}

	if err := h.cs.Update(r.Context(), channelUpdate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		er.HandlerError(h.log, w, err, op, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
