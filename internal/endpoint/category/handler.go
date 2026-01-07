package category

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	"main/internal/lib/sl"
	c "main/pkg/api/category"
	"net/http"
	"strconv"
)

type Getter interface {
	GetById(ctx context.Context, id int) (*Category, error)
	GetByLink(ctx context.Context, link string) (*Category, error)
}

type Lister interface {
	List(ctx context.Context, f CategoryFilter) ([]Category, error)
}

type Creater interface {
	Create(ctx context.Context, cat CategoryCreate) error
}

type Updater interface {
	UpdateByLink(ctx context.Context, link string, cat CategoryUpdate) error
	UpdateById(ctx context.Context, id int, cat CategoryUpdate) error
}

type Deleter interface {
	DeleteById(ctx context.Context, id int) error
	DeleteByLink(ctx context.Context, link string) error
}

type ViewerUpdater interface {
	UpdateViewersByLink(ctx context.Context, link string, viewers int) error
	UpdateViewersById(ctx context.Context, id int, viewers int) error
}

type Repository interface {
	Getter
	Lister
	Creater
	Deleter
	Updater
	ViewerUpdater
}

type Handler struct {
	repo Repository
	log  *slog.Logger
}

// TODO: remove
// handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
// handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
// handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
// handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusBadRequest, handler.MsgIdentity)

func NewHandler(log *slog.Logger, repo Repository) *Handler {
	return &Handler{repo: repo, log: log}
}

// Get retrieves a category by its ID or unique link.
//
// @Summary Get category by ID or link
// @Description Retrieves detailed information about a category by either its numeric ID or a unique link (slug).
// @Tags category
// @Param categoryIdentifier path string true "Category identifier (either numeric ID or unique link)"
// @Accept json
// @Produce json
// @Success 200 {object} GetResponse "Category found successfully"
// @Failure 500 {object} handler.RequestError "Internal server error"
// @Router /categories/{categoryIdentifier} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting category"
	ctx := r.Context()

	// {categoryIdentifier} is either int id or string category link (e.g. "path-of-exile")
	categoryIdentifier := r.PathValue("categoryIdentifier")

	categoryId, err := strconv.Atoi(categoryIdentifier)
	var category *Category = nil
	if err == nil {
		res, err := h.repo.GetById(ctx, categoryId)
		if err != nil {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		category = res
	} else {
		res, err := h.repo.GetByLink(ctx, categoryIdentifier)
		if err != nil {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		category = res
	}

	json.NewEncoder(w).Encode(c.GetResponse{Category: c.Category{
		Id:        category.Id,
		IsSafe:    category.IsSafe,
		Thumbnail: category.Thumbnail,
		Name:      category.Name,
		Link:      category.Link,
		Viewers:   category.Viewers,
		Tags:      category.Tags,
	}})
}

// List retrieves list of categories with page, count and sort by viewers (asc/desc) filters
//
// @Summary List categories with pagination and sorting
// @Description Retrieves a list of categories with optional pagination and sorting parameters
// @Tags category
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param count query int false "Number of items per page" default(10)
// @Param sort query string false "Sort order (asc or desc)" Enums(asc, desc) default(desc)
// @Success 200 {object} ListResponse "List of categories"
// @Failure 400 {object} handler.RequestError "Invalid query parameter"
// @Failure 500 {object} handler.RequestError "Internal server error while fetching categories"
// @Router /categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "getting categories"

	page := r.URL.Query().Get("page")
	count := r.URL.Query().Get("count")
	sort := r.URL.Query().Get("sort")

	if page == "" {
		page = "1"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgBadPage)
		return
	}

	if count == "" {
		count = "10"
	}

	countInt, err := strconv.Atoi(count)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgBadCount)
		return
	}

	if sort == "" {
		sort = "desc"
	}

	if sort != "asc" && sort != "desc" {
		handler.Error(h.log, w, op, handler.ErrBadSort, http.StatusBadRequest, handler.MsgBadSort)
		return
	}

	categories, err := h.repo.List(r.Context(), CategoryFilter{
		Page:  uint32(pageInt),
		Count: uint64(countInt),
		Sort:  sort,
	})
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	categoriesList := make([]c.ListResponseItem, len(categories))
	for i, cat := range categories {
		categoriesList[i] = c.ListResponseItem{
			Name:      cat.Name,
			Thumbnail: cat.Thumbnail,
			Link:      cat.Link,
			Viewers:   cat.Viewers,
			Tags:      cat.Tags,
			IsSafe:    cat.IsSafe,
		}
	}

	json.NewEncoder(w).Encode(c.ListResponse{Categories: categoriesList})
}

// Post creates a new category
//
// @Summary Add a new category
// @Description Creates a new category with the given details
// @Tags category
// @Accept json
// @Produce json
// @Param request body PostRequest true "Category creation request body"
// @Success 204 "No Content - category created successfully"
// @Failure 400 {object} handler.RequestError "Bad Request: invalid data"
// @Failure 500 {object} handler.RequestError "Internal Server Error: failed to add category"
// @Router /categories [post]
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	const op = "creating category"

	ctx := r.Context()
	cl := ctx.Value(auth.AuthContextKey{})
	claims, ok := cl.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if claims.Role != "staff" {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.ErrNotAllowed.Error())
		return
	}

	var req c.PostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	err = h.repo.Create(ctx, CategoryCreate{
		Thumbnail: req.Thumbnail,
		Name:      req.Name,
		Link:      req.Link,
		Viewers:   0,
		Tags:      req.Tags,
	})
	if err != nil {
		if errors.Is(err, errAlreadyExists) {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, errAlreadyExists.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Patch updates a category by its ID or unique link.
//
// @Summary Update category by ID or link
// @Description Partially updates a category entity using provided fields, identified by either numeric ID or unique link (slug).
// @Tags category
// @Accept json
// @Produce json
// @Param categoryIdentifier path string true "Category identifier (either numeric ID or unique link)"
// @Param patchRequest body PatchRequest true "Fields to update in the category"
// @Success 204 "No Content - category updated successfully"
// @Failure 400 {object} handler.RequestError "Invalid request data"
// @Failure 500 {object} handler.RequestError "Internal server error"
// @Router /categories/{categoryIdentifier} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating category"

	ctx := r.Context()
	cl := ctx.Value(auth.AuthContextKey{})
	claims, ok := cl.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if claims.Role != "staff" {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	var request c.PatchRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	categoryIdentifier := r.PathValue("categoryIdentifier")
	categoryId, err := strconv.Atoi(categoryIdentifier)
	if err != nil {
		h.log.Error("category identifier is not parsable as id", sl.Err(err))
	} else {
		err := h.repo.UpdateById(ctx, categoryId, CategoryUpdate{
			Thumbnail: request.Thumbnail,
			Name:      request.Name,
			Link:      request.Link,
			Tags:      request.Tags,
			IsSafe:    request.IsSafe,
		})
		if err != nil {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = h.repo.UpdateByLink(ctx, categoryIdentifier, CategoryUpdate{
		Thumbnail: request.Thumbnail,
		Name:      request.Name,
		Link:      request.Link,
		Tags:      request.Tags,
		IsSafe:    request.IsSafe,
	})
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete removes a category by its ID or unique link.
//
// @Summary Delete category by ID or link
// @Description Deletes a category entity identified by either its numeric ID or unique link (slug).
// @Tags category
// @Param categoryIdentifier path string true "Category identifier (either numeric ID or unique link)"
// @Accept json
// @Produce json
// @Success 204 "No Content - category delted successfully"
// @Failure 500 {object} handler.RequestError "Internal server error"
// @Router /categories/{categoryIdentifier} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "deleting category"

	ctx := r.Context()
	cl := ctx.Value(auth.AuthContextKey{})
	claims, ok := cl.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if claims.Role != "staff" {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.ErrNotAllowed.Error())
		return
	}

	categoryIdentifier := r.PathValue("categoryIdentifier")
	categoryId, err := strconv.Atoi(categoryIdentifier)
	if err != nil {
		h.log.Error("category identifier is not parsable as id", sl.Err(err))
	} else {
		err := h.repo.DeleteById(ctx, categoryId)
		if err != nil {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}
	}

	err = h.repo.DeleteByLink(ctx, categoryIdentifier)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
