package category

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	"main/internal/lib/sl"
	c "main/pkg/api/model/category"
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
	UpdateById(ctx context.Context, id int32, cat CategoryUpdate) error
}

type Deleter interface {
	DeleteById(ctx context.Context, id int32) error
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

func NewHandler(log *slog.Logger, repo Repository) *Handler {
	return &Handler{repo: repo, log: log}
}

// Get retrieves a category by its ID or unique link.
// @Summary      Get category by ID or link
// @Description  Retrieve a category using either its numeric ID or string link (e.g. "path-of-exile")
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        categoryIdentifier  path    string  true  "Category ID or link"  min(1)
// @Success      200  {object}  c.GetResponse
// @Failure      404  {object}  handler.ErrorResponse  "Category not found"
// @Failure      500  {object}  handler.ErrorResponse  "Internal server error"
// @Router       /categories/{categoryIdentifier} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting category"
	ctx := r.Context()

	// {categoryIdentifier} is either int id or string category link (e.g. "path-of-exile")
	categoryIdentifier := r.PathValue("categoryIdentifier")

	var category *Category = nil
	var getErr error = nil

	// TODO: don't cast here
	categoryId, err := strconv.Atoi(categoryIdentifier)
	if err == nil {
		category, getErr = h.repo.GetById(ctx, categoryId)
	} else {
		category, getErr = h.repo.GetByLink(ctx, categoryIdentifier)
	}

	if getErr != nil {
		if getErr == errNotFound {
			handler.Error(h.log, w, op, getErr, http.StatusNotFound, errNotFound.Error())
			return
		}

		handler.Error(h.log, w, op, getErr, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	tags := make([]c.CategoryTag, len(category.Tags))
	for i, t := range category.Tags {
		tags[i] = c.CategoryTag{Id: t.Id, Name: t.Name}
	}

	json.NewEncoder(w).Encode(c.GetResponse{Category: c.Category{
		Id:        category.Id,
		IsSafe:    category.IsSafe,
		Thumbnail: category.Thumbnail,
		Name:      category.Name,
		Link:      category.Link,
		Viewers:   category.Viewers,
		Tags:      tags,
	}})
}

// List retrieves list of categories with page, count and sort by viewers (asc/desc) filters
// @Summary      List categories
// @Description  Retrieve a paginated list of categories with optional sorting
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        page   query    string  false  "Page number (default: 1)"
// @Param        count  query    string  false  "Items per page (default: 10)"
// @Param        sort   query    string  false  "Sort order (asc, desc) (default: desc)"
// @Success      200    {object}  c.ListResponse
// @Failure      400    {object}  handler.ErrorResponse  "Invalid page, count, or sort parameter"
// @Failure      500    {object}  handler.ErrorResponse  "Internal server error"
// @Router       /categories [get]
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
		tags := make([]c.CategoryTag, len(cat.Tags))
		for i, t := range cat.Tags {
			tags[i] = c.CategoryTag{Id: t.Id, Name: t.Name}
		}

		categoriesList[i] = c.ListResponseItem{
			Name:      cat.Name,
			Thumbnail: cat.Thumbnail,
			Link:      cat.Link,
			Viewers:   cat.Viewers,
			Tags:      tags,
			IsSafe:    cat.IsSafe,
		}
	}

	json.NewEncoder(w).Encode(c.ListResponse{Categories: categoriesList})
}

// Post creates a new category
// @Summary      Create new category
// @Description  Create a new category (staff only)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        auth      header     string                  true  "Bearer token"  format(jwt)
// @Param        request   body       c.PostRequest           true  "Category data"
// @Security     BearerAuth
// @Success      204       "Category created successfully"
// @Failure      400       {object}  handler.ErrorResponse  "Invalid auth, insufficient permissions, or malformed request"
// @Failure      500       {object}  handler.ErrorResponse  "Internal server error or category already exists"
// @Router       /categories [post]
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
// @Summary      Update category
// @Description  Partially update a category by ID or link (staff only)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        categoryIdentifier  path    string  true  "Category ID or link"  min(1)
// @Param        auth                header  string  true  "Bearer token"  format(jwt)
// @Param        request             body    c.PatchRequest  true  "Fields to update"
// @Security     BearerAuth
// @Success      204  "Category updated successfully"
// @Failure      400  {object}  handler.ErrorResponse  "Invalid auth, insufficient permissions, or malformed request"
// @Failure      500  {object}  handler.ErrorResponse  "Internal server error"
// @Router       /categories/{categoryIdentifier} [patch]
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

	upd := CategoryUpdate{
		IsSafe:    request.IsSafe,
		Thumbnail: request.Thumbnail,
		Name:      request.Name,
		Link:      request.Link,
		Tags:      request.Tags}

	categoryIdentifier := r.PathValue("categoryIdentifier")
	categoryId, err := strconv.Atoi(categoryIdentifier)
	if err != nil {
		h.log.Error("category identifier is not parsable as id", sl.Err(err))
	} else {
		err := h.repo.UpdateById(ctx, int32(categoryId), upd)
		if err != nil {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = h.repo.UpdateByLink(ctx, categoryIdentifier, upd)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete removes a category by its ID or unique link.
// @Summary      Delete category
// @Description  Delete a category by ID or link (staff only)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        categoryIdentifier  path    string  true  "Category ID or link"  min(1)
// @Param        auth                header  string  true  "Bearer token"  format(jwt)
// @Security     BearerAuth
// @Success      204  "Category deleted successfully"
// @Failure      400  {object}  handler.ErrorResponse  "Invalid or insufficient auth permissions"
// @Failure      500  {object}  handler.ErrorResponse  "Internal server error"
// @Router       /categories/{categoryIdentifier} [delete]
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
		err := h.repo.DeleteById(ctx, int32(categoryId))
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
