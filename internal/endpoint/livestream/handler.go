package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	api "main/pkg/api/livestream"
	"net/http"
	"strconv"
)

type Getter interface {
	Get(ctx context.Context, username string) (*Livestream, error)
}

type Creater interface {
	Create(ctx context.Context, cr LivestreamCreate) (*Livestream, error)
}

type Updater interface {
	Update(ctx context.Context, channel string, upd LivestreamUpdate) (*Livestream, error)
	UpdateViewers(ctx context.Context, channel string, viewers int32) error
	UpdateThumbnail(ctx context.Context, channel, thumbnail string) error
}

type Lister interface {
	List(ctx context.Context, category string, page, count int) ([]Livestream, error)
	ListAll(ctx context.Context) ([]Livestream, error)
}

type Deleter interface {
	Delete(ctx context.Context, username string) (bool, error)
}

type Repository interface {
	Getter
	Creater
	Updater
	Lister
	Deleter
}

type GetterListerUpdater interface {
	Getter
	Lister
	Updater
}

type Handler struct {
	r   GetterListerUpdater
	log *slog.Logger
}

func NewHandler(log *slog.Logger, r GetterListerUpdater) *Handler {
	return &Handler{r: r, log: log}
}

// Get godoc
//
//	@Summary		Get livestream data
//	@Description	Retrieve livestream information for a streamer
//	@Tags			Livestreams
//	@Produce		json
//	@Param			username	path		string					true	"Streamer username"
//	@Success		200			{object}	api.GetResponse			"Livestream data"
//	@Failure		404			{object}	handler.ErrorResponse	"Livestream not found"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/livestreams/{username} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting livestream data"

	streamer := r.PathValue("username")
	ls, err := h.r.Get(r.Context(), streamer)
	if err != nil {
		if errors.Is(err, errNotFound) {
			handler.Error(h.log, w, op, err, http.StatusNotFound, errNotFound.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusNotFound, handler.MsgInternal)
		return
	}

	response := api.GetResponse{
		Id:        int(ls.Id),
		Username:  ls.UserName,
		Avatar:    ls.UserAvatar,
		StartedAt: int(ls.StartedAt),
		Viewers:   ls.Viewers,
		Category: api.LivestreamCategory{
			Link: ls.CategoryLink,
			Name: ls.CategoryName,
		},
		Title:         ls.Title,
		IsLive:        true,
		IsMultistream: false,
		Thumbnail:     ls.Thumbnail,
		IsFollowing:   false,
		IsSubscriber:  false,
	}

	json.NewEncoder(w).Encode(response)
}

// List godoc
//
//	@Summary		List livestreams by category
//	@Description	Get paginated list of livestreams filtered by category
//	@Tags			Livestreams
//	@Produce		json
//	@Param			category	query		string					false	"Category name"
//	@Param			categoryId	query		string					false	"Category identifier"	minLength(1)
//	@Param			page		query		int						false	"Page number"			default(1)	minimum(1)
//	@Param			count		query		int						false	"Items per page"		default(10)	minimum(1)	maximum(100)
//	@Success		200			{object}	api.ListResponse		"Paginated livestreams"
//	@Failure		400			{object}	handler.ErrorResponse	"Missing category, invalid page/count"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/livestreams [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "getting livestreams"

	category := r.URL.Query().Get("category")
	categoryIdentifier := r.URL.Query().Get("categoryId")
	if categoryIdentifier == "" && category == "" {
		handler.Error(h.log, w, op, errNoCategory, http.StatusBadRequest, errNoCategory.Error())
		return
	}

	if categoryIdentifier == "" {
		categoryIdentifier = category
	}

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgBadPage)
		return
	}

	if pageInt < 1 {
		pageInt = 1
	}

	count := r.URL.Query().Get("count")
	if count == "" {
		count = "10"
	}

	countInt, err := strconv.Atoi(count)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgBadCount)
		return
	}

	if countInt < 1 {
		countInt = 10
	}

	livestreams, err := h.r.List(r.Context(), categoryIdentifier, pageInt, countInt)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	listResponse := api.ListResponse{
		Livestreams: make([]api.ListResponseItem, len(livestreams)),
	}

	for i, ls := range livestreams {
		listResponse.Livestreams[i] = api.ListResponseItem{
			Username: ls.UserName,
			Avatar:   ls.UserAvatar,
			Category: api.LivestreamCategory{
				Name: ls.CategoryName,
				Link: ls.CategoryLink,
			},
			StartedAt: int(ls.StartedAt),
			Thumbnail: ls.Thumbnail,
			Viewers:   ls.Viewers,
			Title:     ls.Title,
		}
	}

	json.NewEncoder(w).Encode(listResponse)
}

// Patch godoc
//
//	@Summary		Update livestream
//	@Description	Update livestream title and category (owner only)
//	@Tags			Livestreams
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			username	path		string					true	"Livestream owner username"
//	@Param			request		body		api.PatchRequest		true	"Update data (title, categoryId)"
//	@Success		200			{object}	api.PatchResponse		"Update status"
//	@Failure		401			{object}	handler.ErrorResponse	"Unauthorized - invalid claims or identity mismatch"
//	@Failure		400			{object}	handler.ErrorResponse	"Invalid request"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/livestreams/{username} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating livestream"

	username := r.PathValue("username")
	ctx := r.Context()
	user, ok := auth.FromContext(ctx)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusUnauthorized, handler.MsgIdentity)
		return
	}

	if user.Username != username {
		handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusUnauthorized, handler.MsgIdentity)
		return
	}

	var request api.PatchRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	// TODO: eventual consistency between postgres and redis
	// (postgres got updated while redis is down -> big bad)
	_, err = h.r.Update(ctx, username, LivestreamUpdate{
		Title:      request.Title,
		CategoryId: request.CategoryId,
	})
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	resp := api.PatchResponse{
		Status: true,
	}

	json.NewEncoder(w).Encode(resp)
}
