package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	l "main/pkg/api/livestream"
	"net/http"
	"strconv"
)

type Service interface {
	Get(ctx context.Context, username string) (*Livestream, error)
	Update(ctx context.Context, user string, ls LivestreamUpdate) (bool, error)
	List(ctx context.Context, s LivestreamSearch) ([]Livestream, error)
}

type Handler struct {
	s   Service
	log *slog.Logger
}

func NewHandler(log *slog.Logger, s Service) *Handler {
	return &Handler{s: s, log: log}
}

// GetLivestream godoc
// @Summary      Get livestream data by username
// @Description  Retrieves livestream information for the specified streamer username
// @Tags         livestream
// @Accept       json
// @Produce      json
// @Param        username  path      string  true  "Streamer Username"
// @Success      200       {object}  GetResponse
// @Failure      500       {object}  er.RequestError
// @Router       /livestreams/{username} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting livestream data"

	streamer := r.PathValue("username")
	ls, err := h.s.Get(r.Context(), streamer)
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
// @Summary List livestreams by category or category ID
// @Description Retrieves a paginated list of livestreams filtered by category or category ID.
// @Tags livestream
// @Accept json
// @Produce json
// @Param category query string false "Category name to filter by"
// @Param categoryId query string false "Category ID to filter by"
// @Param page query int false "Page number for pagination (default 1)" default(1)
// @Param count query int false "Number of results per page (default 10)" default(10)
// @Success 200 {object} ListResponse "Livestream list response"
// @Failure 400 {object} er.RequestError "Bad request: missing or invalid parameters"
// @Failure 500 {object} er.RequestError "Internal server error"
// @Router /livestreams [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "getting livestreams"

	category := r.URL.Query().Get("category")
	categoryId := r.URL.Query().Get("categoryId")
	if categoryId == "" && category == "" {
		handler.Error(h.log, w, op, errNoCategory, http.StatusBadRequest, errNoCategory.Error())
		return
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

	count := r.URL.Query().Get("count")
	if count == "" {
		count = "10"
	}

	countInt, err := strconv.Atoi(count)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgBadCount)
		return
	}

	livestreams, err := h.s.List(r.Context(), LivestreamSearch{
		CategoryId: categoryId,
		Category:   category,
		Page:       pageInt,
		Count:      countInt,
	})
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
			StartedAt:     int(ls.StartedAt),
			IsLive:        true,
			IsMultistream: false,
			Thumbnail:     ls.Thumbnail,
			Viewers:       ls.Viewers,
			Title:         ls.Title,
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

	// TODO: sentinel error
	status, err := h.s.Update(r.Context(),
		username, LivestreamUpdate{
			Title:        request.Title,
			CategoryLink: request.CategoryLink,
		})
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	resp := l.PatchResponse{
		Status: status,
	}

	json.NewEncoder(w).Encode(resp)
}
