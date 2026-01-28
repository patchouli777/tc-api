package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"twitchy-api/internal/lib/handler"
	d "twitchy-api/internal/livestream/domain"
	api "twitchy-api/pkg/api/livestream"
)

type Getter interface {
	Get(ctx context.Context, id int) (*d.Livestream, error)
	GetByUsername(ctx context.Context, username string) (*d.Livestream, error)
}

type Lister interface {
	List(ctx context.Context, s d.LivestreamSearch) ([]d.Livestream, error)
}

type GetterLister interface {
	Getter
	Lister
}

type Handler struct {
	r   GetterLister
	log *slog.Logger
}

func NewHandler(log *slog.Logger, r GetterLister) *Handler {
	return &Handler{r: r, log: log}
}

// Get godoc
//
//	@Summary		Get livestream by ID
//	@Description	Retrieve a specific livestream by its ID
//	@Tags			Livestreams
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Livestream ID"
//	@Success		200	{object}	LivestreamGetResponse
//	@Failure		400	{object}	ErrorResponse	"Invalid ID"
//	@Failure		404	{object}	ErrorResponse	"Livestream not found"
//	@Router			/livestreams/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting livestream"

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	ls, err := h.r.Get(r.Context(), idInt)
	if err != nil {
		if errors.Is(err, d.ErrNotFound) {
			handler.Error(h.log, w, op, err, http.StatusNotFound, d.ErrNotFound.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusNotFound, handler.MsgInternal)
		return
	}

	response := ls.ToGetResponse()
	json.NewEncoder(w).Encode(response)
}

// Get godoc
//
//	@Summary		Get livestream by username
//	@Description	Retrieve livestream for a specific user by username
//	@Tags			Livestreams
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	LivestreamGetResponse
//	@Failure		404			{object}	ErrorResponse	"User livestream not found"
//	@Router			/livestreams/username/{username} [get]
func (h *Handler) GetByUsername(w http.ResponseWriter, r *http.Request) {
	const op = "getting user livestream"

	user := r.PathValue("username")
	ls, err := h.r.GetByUsername(r.Context(), user)
	if err != nil {
		if errors.Is(err, d.ErrNotFound) {
			handler.Error(h.log, w, op, err, http.StatusNotFound, d.ErrNotFound.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusNotFound, handler.MsgInternal)
		return
	}

	response := ls.ToGetResponse()
	json.NewEncoder(w).Encode(response)
}

// List godoc
//
//	@Summary		List livestreams
//	@Description	Get paginated list of livestreams with optional filtering
//	@Tags			Livestreams
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	false	"Page number (default: 1)"
//	@Param			count		query		string	false	"Items per page (default: 10, min: 1)"
//	@Param			category	query		string	false	"Category name filter"
//	@Param			categoryId	query		string	false	"Category ID filter"
//	@Success		200			{object}	ListResponse
//	@Failure		400			{object}	ErrorResponse	"Invalid page or count parameters"
//	@Failure		500			{object}	ErrorResponse	"Internal server error"
//	@Router			/livestreams [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "getting livestreams"

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	errs := make(map[string]error)

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		errs["page"] = handler.ErrBadPage
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
		errs["count"] = handler.ErrBadCount
	}

	if len(errs) != 0 {
		handler.Errors(h.log, w, op, http.StatusBadRequest, errs)
		return
	}

	if countInt < 1 {
		countInt = 10
	}

	category := r.URL.Query().Get("category")
	categoryId := r.URL.Query().Get("categoryId")

	livestreams, err := h.r.List(r.Context(), d.LivestreamSearch{
		Page:       pageInt,
		Count:      countInt,
		CategoryId: categoryId,
		Category:   category,
	})
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	listResponse := api.ListResponse{
		Livestreams: make([]api.ListResponseItem, len(livestreams)),
	}

	for i, ls := range livestreams {
		listResponse.Livestreams[i] = ls.ToListResponseItem()
	}

	json.NewEncoder(w).Encode(listResponse)
}
