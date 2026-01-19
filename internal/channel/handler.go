package channel

import (
	"context"
	"encoding/json"
	"log/slog"
	d "main/internal/channel/domain"
	"main/internal/lib/handler"
	api "main/pkg/api/channel"
	"net/http"
)

type Repository interface {
	Get(ctx context.Context, username string) (*d.Channel, error)
	Update(ctx context.Context, upd d.ChannelUpdate) error
}

type Handler struct {
	cr  Repository
	log *slog.Logger
}

func NewHandler(log *slog.Logger, s Repository) *Handler {
	return &Handler{cr: s, log: log}
}

// Get retrieves a channel by its ID (username of owner)
//
//	@Summary		Get channel details
//	@Description	Retrieve detailed information about a specific channel
//	@Tags			Channels
//	@Accept			json
//	@Produce		json
//	@Param			channel	path		string					true	"Channel identifier"	min(1)
//	@Success		200		{object}	api.GetResponse			"Channel details"
//	@Failure		400		{object}	handler.ErrorResponse	"Missing channel identifier"
//	@Failure		500		{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/channels/{channel} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting channel"

	channel := r.PathValue("channel")

	if channel == "" {
		handler.Error(h.log, w, op, d.ErrNotPresent, http.StatusBadRequest, d.ErrNotPresent.Error())
		return
	}

	res, err := h.cr.Get(r.Context(), channel)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	json.NewEncoder(w).Encode(api.GetResponse{
		IsBanned:        res.IsBanned,
		IsPartner:       res.IsPartner,
		Background:      res.Background,
		FirstLivestream: res.FirstLivestream,
		LastLivestream:  res.LastLivestream,
		Description:     res.Description,
		// Links:           res.Links,
		// Tags:            res.Tags,
	})
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating channel"

	var channelUpdate d.ChannelUpdate
	if err := json.NewDecoder(r.Body).Decode(&channelUpdate); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if err := h.cr.Update(r.Context(), channelUpdate); err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
