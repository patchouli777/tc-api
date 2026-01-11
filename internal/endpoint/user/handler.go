package user

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	u "main/pkg/api/model/user"
	"net/http"
)

type Service interface {
	Get(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, u UserCreate) error
	Update(ctx context.Context, u UserUpdate) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, l UserList) ([]User, error)
}

type Handler struct {
	s   Service
	log *slog.Logger
}

func NewHandler(log *slog.Logger, s Service) *Handler {
	return &Handler{s: s, log: log}
}

// Get godoc
// @Summary      Get user by username
// @Description  Retrieve user information by username
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        username  path    string  true  "Username"
// @Success      200  {object}  GetResponse
// @Failure      404  {object}  GetResponseNotFound
// @Failure      500  {string}  string  "Internal server error"
// @Router       /users/{username} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting user"

	username := r.PathValue("username")

	user, err := h.s.Get(r.Context(), username)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	if user == nil {
		json.NewEncoder(w).Encode(u.GetResponseNotFound{
			Result: "specified user not found",
		})
		return
	}

	json.NewEncoder(w).Encode(u.GetResponse{
		Id:              user.Id,
		Name:            user.Name,
		IsBanned:        user.IsBanned,
		IsPartner:       user.IsPartner,
		FirstLivestream: user.FirstLivestream,
		LastLivestream:  user.LastLivestream,
		Avatar:          user.Avatar,
	})
}

// Post godoc
// @Summary      Create new user
// @Description  Create a new user account with name, password, and optional avatar
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      PostRequest  true  "User creation data"
// @Success      204  {object}  nil
// @Failure      400  {string}  string  "Invalid request data or missing fields"
// @Failure      409  {string}  string  "Username already exists"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /users [post]
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	const op = "creating user"

	var req u.PostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if req.Name == nil {
		handler.Error(h.log, w, op, errUsernameRequired, http.StatusBadRequest, errUsernameRequired.Error())
		return
	}

	if req.Password == nil {
		handler.Error(h.log, w, op, errPasswordRequired, http.StatusBadRequest, errPasswordRequired.Error())
		return
	}

	av := ""
	if req.Avatar != nil {
		av = *req.Avatar
	}

	if err := h.s.Create(r.Context(), UserCreate(UserCreate{
		Name:     *req.Name,
		Password: *req.Password,
		Avatar:   av,
	})); err != nil {
		if errors.Is(err, errUserExists) {
			handler.Error(h.log, w, op, err, http.StatusConflict, errUserExists.Error())
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
// @Summary      Update user profile
// @Description  Update authenticated user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        username  path    string  true  "Username"
// @Param        request   body    PatchRequest  true  "User update data"
// @Security     BearerAuth
// @Success      204  {object}  nil
// @Failure      400  {string}  string  "Invalid claims, identity mismatch, or invalid request data"
// @Failure      409  {string}  string  "Username already exists"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /users/{username} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	const op = "updating user"

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

	var req u.PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if err := h.s.Update(ctx, UserUpdate{
		Name:      req.Name,
		Password:  req.Password,
		Avatar:    req.Avatar,
		IsBanned:  req.IsBanned,
		IsPartner: req.IsPartner,
	}); err != nil {
		if errors.Is(err, errUserExists) {
			handler.Error(h.log, w, op, err, http.StatusConflict, errUserExists.Error())
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
// @Summary      Delete user account
// @Description  Delete authenticated user's account by providing user ID in request body
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        username  path    string  true  "Username"
// @Param        request   body    DeleteRequest  true  "User ID to delete"
// @Security     BearerAuth
// @Success      204  {object}  nil
// @Failure      400  {string}  string  "Invalid claims, identity mismatch, or invalid request data"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /users/{username} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "deleting user"

	username := r.PathValue("username")

	ctx := r.Context()
	claims := ctx.Value(auth.AuthContextKey{})
	// TODO: добавить везде ok
	user, ok := claims.(*auth.Claims)
	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if user.Username != username {
		handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	var req u.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	if err := h.s.Delete(ctx, req.UserId); err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List godoc
// @Summary      List all users
// @Description  List users with optional filters (staff only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body    ListRequest  true  "List filters"
// @Security     BearerAuth
// @Success      200  {object}  ListResponse
// @Failure      400  {string}  string  "Invalid claims or insufficient permissions"
// @Failure      500  {string}  string  "Internal server error"
// @Router       /users [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "list users"

	ctx := r.Context()
	claims := ctx.Value(auth.AuthContextKey{})
	user, ok := claims.(*auth.Claims)
	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	if user.Role != "staff" {
		handler.Error(h.log, w, op, handler.ErrNotAllowed, http.StatusBadRequest, handler.MsgIdentity)
		return
	}

	var req u.ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	users, err := h.s.List(ctx, UserList(req))
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	usersResponse := make([]u.ListResponseItem, len(users))

	for i, user := range users {
		usersResponse[i] = u.ListResponseItem{
			Id:              user.Id,
			Name:            user.Name,
			IsBanned:        user.IsBanned,
			IsPartner:       user.IsPartner,
			FirstLivestream: user.FirstLivestream,
			LastLivestream:  user.LastLivestream,
			Avatar:          user.Avatar,
			Description:     user.Description,
		}
	}

	listResponse := u.ListResponse{
		Users: usersResponse,
	}

	json.NewEncoder(w).Encode(listResponse)
}
