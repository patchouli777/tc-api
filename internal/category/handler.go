package category

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/app/auth"
	d "main/internal/category/domain"
	"main/internal/lib/handler"
	api "main/pkg/api/category"
	"net/http"
	"strconv"
)

type Getter interface {
	Get(ctx context.Context, id int) (*d.Category, error)
	GetByLink(ctx context.Context, link string) (*d.Category, error)
}

type Lister interface {
	List(ctx context.Context, f d.CategoryFilter) ([]d.Category, error)
}

type Creater interface {
	Create(ctx context.Context, cat d.CategoryCreate) error
}

type Updater interface {
	Update(ctx context.Context, id int32, cat d.CategoryUpdate) error
	UpdateByLink(ctx context.Context, link string, cat d.CategoryUpdate) error
}

type Deleter interface {
	Delete(ctx context.Context, id int32) error
	DeleteByLink(ctx context.Context, link string) error
}

type Repository interface {
	Getter
	Lister
	Creater
	Deleter
	Updater
}

type Handler struct {
	repo Repository
	log  *slog.Logger
}

func NewHandler(log *slog.Logger, repo Repository) *Handler {
	return &Handler{repo: repo, log: log}
}

// Get retrieves a category by its ID or unique link.
//
//	@Summary		Get category by ID or link
//	@Description	Retrieve a category using either its numeric ID or string link (e.g. "path-of-exile")
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			identifier	path		string	true	"Category ID or link"	min(1)
//	@Success		200			{object}	c.GetResponse
//	@Failure		404			{object}	handler.ErrorResponse	"Category not found"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/categories/{identifier} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting category"
	ctx := r.Context()

	// {identifier} is either int id or string category link (e.g. "path-of-exile")
	identifier := r.PathValue("identifier")

	var category *d.Category = nil
	var getErr error = nil

	categoryId, err := strconv.Atoi(identifier)
	if err == nil {
		category, getErr = h.repo.Get(ctx, categoryId)
	} else {
		category, getErr = h.repo.GetByLink(ctx, identifier)
	}

	if getErr != nil {
		if getErr == d.ErrNotFound {
			handler.Error(h.log, w, op, getErr, http.StatusNotFound, d.ErrNotFound.Error())
			return
		}

		handler.Error(h.log, w, op, getErr, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	response := category.ToGetResponse()
	json.NewEncoder(w).Encode(response)
}

// List retrieves list of categories with page, count and sort by viewers (asc/desc) filters
//
//	@Summary		List categories
//	@Description	Retrieve a paginated list of categories with optional sorting
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"Page number (default: 1)"
//	@Param			count	query		string	false	"Items per page (default: 10)"
//	@Param			sort	query		string	false	"Sort order (asc, desc) (default: desc)"
//	@Success		200		{object}	c.ListResponse
//	@Failure		400		{object}	handler.ErrorResponse	"Invalid page, count, or sort parameter"
//	@Failure		500		{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "getting categories"

	page := r.URL.Query().Get("page")
	count := r.URL.Query().Get("count")
	sort := r.URL.Query().Get("sort")

	if page == "" {
		page = "1"
	}
	errs := make(map[string]error)

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		errs["page"] = handler.ErrBadPage
	}

	if count == "" {
		count = "10"
	}

	countInt, err := strconv.Atoi(count)
	if err != nil {
		errs["count"] = handler.ErrBadCount
	}

	if sort == "" {
		sort = "desc"
	}

	if sort != "asc" && sort != "desc" {
		errs["sort"] = handler.ErrBadSort
	}

	if len(errs) != 0 {
		handler.Errors(h.log, w, op, http.StatusBadRequest, errs)
		return
	}

	categories, err := h.repo.List(r.Context(), d.CategoryFilter{
		Page:  uint32(pageInt),
		Count: uint64(countInt),
		Sort:  sort,
	})
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	categoriesList := make([]api.ListResponseItem, len(categories))
	for i, cat := range categories {
		categoriesList[i] = cat.ToListResponseItem()
	}

	json.NewEncoder(w).Encode(api.ListResponse{Categories: categoriesList})
}

// Post creates a new category
//
//	@Summary		Create new category
//	@Description	Create a new category (staff only)
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			auth	header	string			true	"Bearer token"	format(jwt)
//	@Param			request	body	c.PostRequest	true	"Category data"
//	@Security		BearerAuth
//	@Success		204	"Category created successfully"
//	@Failure		400	{object}	handler.ErrorResponse	"Invalid auth, insufficient permissions, or malformed request"
//	@Failure		500	{object}	handler.ErrorResponse	"Internal server error or category already exists"
//	@Router			/categories [post]
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	const op = "creating category"

	ctx := r.Context()
	cl := ctx.Value(auth.AuthContextKey{})
	user, ok := cl.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if user.Role != auth.RoleStaff {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.ErrNotAllowed.Error())
		return
	}

	var req api.PostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if req.Link == "" || req.Name == "" {
		handler.Error(h.log, w, op, d.ErrEmptyNameLink, http.StatusBadRequest, d.ErrEmptyNameLink.Error())
		return
	}

	err = h.repo.Create(ctx, d.CategoryCreate{
		Thumbnail: req.Thumbnail,
		Name:      req.Name,
		Link:      req.Link,
		Tags:      req.Tags,
	})
	if err != nil {
		if errors.Is(err, d.ErrAlreadyExists) {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, d.ErrAlreadyExists.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Patch updates a category by its ID or unique link.
//
//	@Summary		Update category
//	@Description	Partially update a category by ID or link (staff only)
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			identifier	path	string			true	"Category ID or link"	min(1)
//	@Param			auth		header	string			true	"Bearer token"			format(jwt)
//	@Param			request		body	c.PatchRequest	true	"Fields to update"
//	@Security		BearerAuth
//	@Success		204	"Category updated successfully"
//	@Failure		400	{object}	handler.ErrorResponse	"Invalid auth, insufficient permissions, or malformed request"
//	@Failure		500	{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/categories/{identifier} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating category"
	ctx := r.Context()

	cl := ctx.Value(auth.AuthContextKey{})
	user, ok := cl.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if user.Role != auth.RoleStaff {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	var req api.PatchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if req.Link.Value == "" || req.Name.Value == "" {
		handler.Error(h.log, w, op, d.ErrEmptyNameLink, http.StatusBadRequest, d.ErrEmptyNameLink.Error())
		return
	}

	upd := d.CategoryUpdate{
		IsSafe:    req.IsSafe,
		Thumbnail: req.Thumbnail,
		Name:      req.Name,
		Link:      req.Link,
		Tags:      req.Tags}

	identifier := r.PathValue("identifier")
	categoryId, err := strconv.Atoi(identifier)
	if err != nil {
		h.log.Info("updating category by link", slog.String("link", identifier))

		err = h.repo.UpdateByLink(ctx, identifier, upd)
		if err != nil {
			if errors.Is(err, d.ErrNotFound) {
				handler.Error(h.log, w, op, err, http.StatusInternalServerError, err.Error())
				return
			}

			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	} else {
		h.log.Info("updating category by id", slog.Int("id", categoryId))

		err := h.repo.Update(ctx, int32(categoryId), upd)
		if err != nil {
			if errors.Is(err, d.ErrNotFound) {
				handler.Error(h.log, w, op, err, http.StatusInternalServerError, err.Error())
				return
			}

			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Delete removes a category by its ID or unique link.
//
//	@Summary		Delete category
//	@Description	Delete a category by ID or link (staff only)
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			identifier	path	string	true	"Category ID or link"	min(1)
//	@Param			auth		header	string	true	"Bearer token"			format(jwt)
//	@Security		BearerAuth
//	@Success		204	"Category deleted successfully"
//	@Failure		400	{object}	handler.ErrorResponse	"Invalid or insufficient auth permissions"
//	@Failure		500	{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/categories/{identifier} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "deleting category"

	ctx := r.Context()
	cl := ctx.Value(auth.AuthContextKey{})
	claims, ok := cl.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if claims.Role != auth.RoleStaff {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.ErrNotAllowed.Error())
		return
	}

	identifier := r.PathValue("identifier")
	categoryId, err := strconv.Atoi(identifier)
	if err != nil {
		h.log.Error("deleting category by link", slog.String("link", identifier))

		err = h.repo.DeleteByLink(ctx, identifier)
		if err != nil {
			if errors.Is(err, d.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}
	} else {
		h.log.Error("deleting category by id", slog.Int("id", categoryId))

		err := h.repo.Delete(ctx, int32(categoryId))
		if err != nil {
			if errors.Is(err, d.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}
	}
}
