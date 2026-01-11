package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	l "main/pkg/api/model/livestream"
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
// @Summary      Get livestream details
// @Description  Retrieve current livestream data for a streamer
// @Tags         Livestreams
// @Accept       json
// @Produce      json
// @Param        username  path  string  true  "Streamer username"  min(1)
// @Success      200       {object}  l.GetResponse  "Livestream details"
// @Failure      404       {object}  handler.ErrorResponse  "Livestream not found"
// @Failure      500       {object}  handler.ErrorResponse  "Internal server error"
// @Router       /livestreams/{username} [get]
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

	response := l.GetResponse{
		Id:        int(ls.Id),
		Username:  ls.User.Name,
		Avatar:    ls.User.Avatar,
		StartedAt: int(ls.StartedAt),
		Viewers:   ls.Viewers,
		Category: l.LivestreamCategory{
			Link: ls.Category.Link,
			Name: ls.Category.Name,
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
// @Summary      List livestreams
// @Description  Get paginated list of livestreams filtered by category
// @Tags         Livestreams
// @Accept       json
// @Produce      json
// @Param        categoryId  query  string  false  "Category ID (numeric)"
// @Param        category    query  string  false  "Category link (e.g. path-of-exile)"  min(1)
// @Param        page        query  string  false  "Page number (default: 1)"
// @Param        count       query  string  false  "Items per page (default: 10)"
// @Success      200         {object}  l.ListResponse
// @Failure      400         {object}  handler.ErrorResponse  "Missing category filter or invalid pagination"
// @Failure      500         {object}  handler.ErrorResponse  "Internal server error"
// @Router       /livestreams [get]
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

	listResponse := l.ListResponse{
		Livestreams: make([]l.ListResponseItem, len(livestreams)),
	}

	for i, ls := range livestreams {
		listResponse.Livestreams[i] = l.ListResponseItem{
			Username: ls.User.Name,
			Avatar:   ls.User.Avatar,
			Category: l.LivestreamCategory{
				Name: ls.Category.Name,
				Link: ls.Category.Link,
			},
			StartedAt: int(ls.StartedAt),
			Thumbnail: ls.Thumbnail,
			Viewers:   ls.Viewers,
			Title:     ls.Title,
		}
	}

	json.NewEncoder(w).Encode(listResponse)
}

// UpdateLivestream godoc
// @Summary      Update livestream data for a user
// @Description  Updates title and category of the livestream if the authenticated user matches the username in the path
// @Tags         livestream
// @Accept       json
// @Produce      json
// @Param        username  path      string       true  "Username"
// @Param        data      body      PatchRequest  true  "Updated livestream data"
// @Success      200       {object}  PatchResponse
// @Failure      400       {object}  er.RequestError
// @Failure      500       {object}  er.RequestError
// @Security     ApiKeyAuth
// @Router       /livestreams/{username} [patch]

// Patch updates the livestream data for the given username.
// @Summary Update livestream data
// @Description Update livestream information like title and category for the authenticated user.
// @Tags livestream
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param request body PatchRequest true "Patch livestream request body"
// @Success 200 {object} PatchResponse "Successful update response"
// @Failure 400 {object} er.RequestError "Bad request or unauthorized update attempt"
// @Failure 500 {object} er.RequestError "Internal server error updating livestream"
// @Router /livestream/{username} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating livestream"

	username := r.PathValue("username")
	ctx := r.Context()
	claims := ctx.Value(auth.AuthContextKey{})
	user, ok := claims.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if user.Username != username {
		handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	var request l.PatchRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	// TODO: eventual consistency between postgres and redis
	// (postgres got updated while redis is down -> big bad)
	_, err = h.r.Update(r.Context(), username, LivestreamUpdate{
		Title:      request.Title,
		CategoryId: request.CategoryId,
	})
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	resp := l.PatchResponse{
		Status: true,
	}

	json.NewEncoder(w).Encode(resp)
}
