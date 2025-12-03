package livestream

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/er"
	"net/http"
	"strconv"
)

type Service interface {
	Start(ctx context.Context, categoryLink, title, username string) (*Livestream, error)
	Get(ctx context.Context, username string) (*Livestream, error)
	Update(ctx context.Context, user string, ls LivestreamUpdate) (bool, error)
	List(ctx context.Context, s LivestreamSearch) ([]Livestream, error)
	Stop(ctx context.Context, username string) (bool, error)
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
		// TODO: change internal error to not found everywhere
		w.WriteHeader(http.StatusNotFound)
		er.HandlerError(h.log, w, err, op, "resource not found")
		return
	}

	response := GetResponse{
		Id:        int(ls.Id),
		Username:  ls.User.Name,
		Avatar:    ls.User.Avatar,
		StartedAt: int(ls.StartedAt),
		Viewers:   ls.Viewers,
		Category: LivestreamCategory{
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
		const msg = "neither category nor category id is present"
		er.HandlerError(h.log, w, fmt.Errorf("%s", msg), op, msg)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		const msg = "incorrect page"
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, fmt.Errorf("%s", msg), op, msg)
		return
	}

	count := r.URL.Query().Get("count")
	if count == "" {
		count = "10"
	}

	countInt, err := strconv.Atoi(count)
	if err != nil {
		const msg = "incorrect count"
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, fmt.Errorf("%s", msg), op, msg)
		return
	}

	livestreams, err := h.s.List(r.Context(), LivestreamSearch{
		CategoryId: categoryId,
		Category:   category,
		Page:       pageInt,
		Count:      countInt,
	})
	if err != nil {
		er.HandlerError(h.log, w, err, op, "internal error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	listResponse := ListResponse{
		Livestreams: make([]ListResponseItem, len(livestreams)),
	}

	for i, ls := range livestreams {
		listResponse.Livestreams[i] = ListResponseItem{
			Username: ls.User.Name,
			Avatar:   ls.User.Avatar,
			Category: LivestreamCategory{
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

// StartLivestream godoc
// @Summary      Start a new livestream for a user
// @Description  Starts a livestream if the authenticated user matches the username in the path
// @Tags         livestream
// @Accept       json
// @Produce      json
// @Param        username  path      string       true  "Username"
// @Param        data      body      PostRequest  true  "Livestream start data"
// @Success      200       {object}  PostResponse
// @Failure      400       {object}  er.RequestError
// @Failure      500       {object}  er.RequestError
// @Security     ApiKeyAuth
// @Router       /livestreams/{username} [post]
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	const op = "starting livestream"

	username := r.PathValue("username")
	ctx := r.Context()
	claims := ctx.Value(auth.AuthContextKey{})
	cl := claims.(*auth.Claims)

	if cl.Username != username {
		const msg = "identity is not confirmed"
		w.WriteHeader(http.StatusUnauthorized)
		er.HandlerError(h.log, w, fmt.Errorf("%s", msg), op, msg)
		return
	}

	var data PostRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, err, op, "invalid data in the request")
		return
	}

	if data.CategoryLink == "" || data.Title == "" || username == "" {
		const msg = "invalid category link, title or username"
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, fmt.Errorf("%s", msg), op, msg)
		return
	}

	ls, err := h.s.Start(r.Context(), data.CategoryLink, data.Title, username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		er.HandlerError(h.log, w, err, op, "internal error")
		return
	}

	resp := PostResponse{
		Username:  ls.User.Name,
		Category:  ls.Category.Name,
		Title:     ls.Title,
		StartedAt: ls.StartedAt,
		Viewers:   ls.Viewers,
	}

	json.NewEncoder(w).Encode(resp)
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
	user := ctx.Value(auth.AuthContextKey{})
	usr := user.(*auth.Claims)

	if usr.Username != username {
		const msg = "identity is not confirmed"
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, fmt.Errorf("%s", msg), op, msg)
		return
	}

	var request PatchRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, err, op, "invalid data in the request")
		return
	}

	status, err := h.s.Update(r.Context(),
		username, LivestreamUpdate{
			Title:        request.Title,
			CategoryLink: request.CategoryLink,
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		er.HandlerError(h.log, w, err, op, "internal error")
		return
	}

	resp := PatchResponse{
		Status: status,
	}

	json.NewEncoder(w).Encode(resp)
}

// StopLivestream godoc
// @Summary      Stop a livestream for a user
// @Description  Stops the livestream if the authenticated user matches the username in the path
// @Tags         livestream
// @Accept       json
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Success      200       {object}  DeleteResponse
// @Failure      400       {object}  er.RequestError
// @Failure      500       {object}  er.RequestError
// @Security     ApiKeyAuth
// @Router       /livestreams/{username} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "stopping livestream"

	username := r.PathValue("username")
	ctx := r.Context()
	user := ctx.Value(auth.AuthContextKey{})
	usr := user.(*auth.Claims)

	if usr.Username != username {
		const msg = "identity is not confirmed"
		w.WriteHeader(http.StatusBadRequest)
		er.HandlerError(h.log, w, fmt.Errorf("%s", msg), op, msg)
		return
	}

	status, err := h.s.Stop(r.Context(), username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		er.HandlerError(h.log, w, err, op, "internal error")
		return
	}

	resp := DeleteResponse{
		Status: status,
	}

	json.NewEncoder(w).Encode(resp)
}
