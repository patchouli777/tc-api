package user

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/app/auth"
	"main/internal/lib/handler"
	api "main/pkg/api/user"
	"net/http"
	"strconv"
)

type Repository interface {
	Get(ctx context.Context, id int32) (*User, error)
	Create(ctx context.Context, u UserCreate) error
	Update(ctx context.Context, id int32, upd UserUpdate) error
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, l UserList) ([]User, error)
}

type Handler struct {
	s   Repository
	log *slog.Logger
}

func NewHandler(log *slog.Logger, s Repository) *Handler {
	return &Handler{s: s, log: log}
}

// Get godoc
//
//	@Summary		Get user by ID
//	@Description	Retrieve user profile information
//	@Tags			Users
//	@Produce		json
//	@Param			id	path		int						true	"User ID"
//	@Success		200	{object}	api.GetResponse			"User profile"
//	@Failure		400	{object}	handler.ErrorResponse	"Invalid user ID"
//	@Failure		500	{object}	handler.ErrorResponse	"User not found or internal error"
//	@Router			/users/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting user"

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	user, err := h.s.Get(r.Context(), int32(idInt))
	if err != nil {
		if errors.Is(err, errNotFound) {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, err.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	json.NewEncoder(w).Encode(api.GetResponse{
		Id:              int(user.Id),
		Name:            user.Name,
		IsBanned:        user.IsBanned,
		IsLive:          false,
		IsPartner:       user.IsPartner,
		FirstLivestream: user.FirstLivestream,
		LastLivestream:  user.LastLivestream,
		Pfp:             user.Pfp,
	})
}

// Post godoc
//
//	@Summary		Create user account
//	@Description	Register new user account
//	@Tags			Users
//	@Accept			json
//	@Param			request	body		api.PostRequest	true	"User creation data"
//	@Success		204		{object}	nil
//	@Failure		400		{object}	handler.ErrorResponse	"Invalid request, missing name/password, weak password"
//	@Failure		409		{object}	handler.ErrorResponse	"User already exists"
//	@Failure		500		{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/users [post]
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	const op = "creating user"

	var req api.PostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if req.Name == "" {
		handler.Error(h.log, w, op, errUsernameRequired, http.StatusBadRequest, errUsernameRequired.Error())
		return
	}

	if req.Password == "" {
		handler.Error(h.log, w, op, errPasswordRequired, http.StatusBadRequest, errPasswordRequired.Error())
		return
	}

	if err := h.s.Create(r.Context(), UserCreate{
		Name:     req.Name,
		Password: req.Password,
		Pfp:      *req.Pfp,
	}); err != nil {
		if errors.Is(err, errAlreadyExists) {
			handler.Error(h.log, w, op, err, http.StatusConflict, errAlreadyExists.Error())
			return
		}

		if errors.Is(err, errWeakPassword) {
			handler.Error(h.log, w, op, err, http.StatusBadRequest, errWeakPassword.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Patch godoc
//
//	@Summary		Update user
//	@Description	Update user profile (self or staff only)
//	@Tags			Users
//	@Accept			json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"User ID"
//	@Param			request	body		api.PatchRequest	true	"Update data (name, password, avatar, banned, partner)"
//	@Success		204		{object}	nil
//	@Failure		400		{object}	handler.ErrorResponse	"Invalid ID, claims, identity mismatch, request, or weak password"
//	@Failure		409		{object}	handler.ErrorResponse	"Name already exists"
//	@Failure		500		{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/users/{id} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating user"

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	ctx := r.Context()
	user, ok := auth.FromContext(ctx)
	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if user.Role != auth.RoleStaff {
		if user.Id != int32(idInt) {
			handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusBadRequest, handler.MsgIdentity)
			return
		}
	}

	var req api.PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if err := h.s.Update(ctx, int32(idInt), UserUpdate{
		Name:      req.Name,
		Password:  req.Password,
		Pfp:       req.Pfp,
		IsBanned:  req.IsBanned,
		IsPartner: req.IsPartner,
	}); err != nil {
		if errors.Is(err, errAlreadyExists) {
			handler.Error(h.log, w, op, err, http.StatusConflict, errAlreadyExists.Error())
			return
		}

		if errors.Is(err, errWeakPassword) {
			handler.Error(h.log, w, op, err, http.StatusBadRequest, errWeakPassword.Error())
			return
		}

		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete godoc
//
//	@Summary		Delete user account
//	@Description	Delete own user account (self only)
//	@Tags			Users
//	@Accept			json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"User ID"
//	@Param			request	body		api.DeleteRequest	true	"Delete confirmation"
//	@Success		204		{object}	nil
//	@Failure		400		{object}	handler.ErrorResponse	"Invalid ID, claims, identity mismatch, or request"
//	@Failure		500		{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/users/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "deleting user"

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	ctx := r.Context()
	user, ok := auth.FromContext(ctx)
	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	// TODO: int check instead of username + staff check
	if user.Role != auth.RoleStaff {
		if user.Id != int32(idInt) {
			handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusBadRequest, handler.MsgIdentity)
			return
		}
	}

	var req api.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if err := h.s.Delete(ctx, int32(req.Id)); err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "list users"

	ctx := r.Context()
	user, ok := auth.FromContext(ctx)
	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if user.Role != auth.RoleStaff {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	// var req api.ListRequest
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
	// 	return
	// }

	// users, err := h.s.List(ctx, UserList(req))
	// if err != nil {
	// 	handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
	// 	return
	// }

	// usersResponse := make([]api.ListResponseItem, len(users))

	// for i, user := range users {
	// 	usersResponse[i] = api.ListResponseItem{
	// 		Id:              user.Id,
	// 		Name:            user.Name,
	// 		IsBanned:        user.IsBanned,
	// 		IsPartner:       user.IsPartner,
	// 		FirstLivestream: user.FirstLivestream,
	// 		LastLivestream:  user.LastLivestream,
	// 		Pfp:          user.Pfp,
	// 		Description:     user.Description,
	// 	}
	// }

	// listResponse := api.ListResponse{
	// 	Users: usersResponse,
	// }

	// json.NewEncoder(w).Encode(listResponse)

	json.NewEncoder(w).Encode(struct{ Error string }{Error: "not implemented"})
}
