package channel

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/lib/handler"
	c "main/pkg/api/model/channel"
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

// Get retrieves a channel by its ID (username of owner)
// @Summary      Get channel details
// @Description  Retrieve detailed information about a specific channel
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        channel  path  string  true  "Channel identifier"  min(1)
// @Success      200      {object}  c.GetResponse  "Channel details"
// @Failure      400      {object}  handler.ErrorResponse  "Missing channel identifier"
// @Failure      500      {object}  handler.ErrorResponse  "Internal server error"
// @Router       /channels/{channel} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting channel"

	channel := r.PathValue("channel")

	if channel == "" {
		handler.Error(h.log, w, op, errNotPresent, http.StatusBadRequest, errNotPresent.Error())
		return
	}

	chann, err := h.cs.Get(r.Context(), channel)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	json.NewEncoder(w).Encode(c.GetResponse{
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
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if err := h.cs.Update(r.Context(), channelUpdate); err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
