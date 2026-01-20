package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/lib/handler"
	d "main/internal/livestream/domain"
	api "main/pkg/api/livestream"
	"net/http"
	"strconv"
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
//	@Summary		Get livestream data
//	@Description	Retrieve livestream information for a streamer
//	@Tags			Livestreams
//	@Produce		json
//	@Param			username	path		string					true	"Streamer username"
//	@Success		200			{object}	api.GetResponse			"Livestream data"
//	@Failure		404			{object}	handler.ErrorResponse	"Livestream not found"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/livestreams/{username} [get]
//
// TODO: update doc
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

// TODO: update doc
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
//
// TODO: update doc
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
