package follow

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	api "main/pkg/api/follow"
	"net/http"
)

type Repository interface {
	IsFollower(ctx context.Context, follower, followed string) (bool, error)
	List(ctx context.Context, follower string) ([]FollowerListItem, error)
	ListExtended(ctx context.Context, follower string) ([]FollowingListExtendedItem, error)
	Follow(ctx context.Context, follower, followed string) error
	Unfollow(ctx context.Context, unfollower, unfollowed string) error
}

type Handler struct {
	r   Repository
	log *slog.Logger
}

func NewHandler(log *slog.Logger, r Repository) *Handler {
	return &Handler{r: r, log: log}
}

// Get godoc
//
//	@Summary		Check follow status
//	@Description	Check if one user follows another
//	@Tags			Follows
//	@Produce		json
//	@Param			username	path		string					true	"Follower username"
//	@Param			followed	query		string					true	"Followed username"
//	@Success		200			{object}	api.GetResponse			"Follow status"
//	@Failure		400			{object}	handler.ErrorResponse	"Missing followed parameter"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/follows/{username} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting follow status"

	follower := r.PathValue("username")
	ctx := r.Context()

	followed := r.URL.Query().Get("followed")
	if followed == "" {
		handler.Error(h.log, w, op, errNoFollowed, http.StatusBadRequest, errNoFollowed.Error())
		return
	}

	isFollower, err := h.r.IsFollower(ctx, follower, followed)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	json.NewEncoder(w).Encode(api.GetResponse{IsFollower: isFollower})
}

// List godoc
//
//	@Summary		List followers or following
//	@Description	Get list of users that follower is following (basic or extended)
//	@Tags			Follows
//	@Produce		json
//	@Param			follower	query		string						true	"Username to get follow list for"
//	@Param			extended	query		string						false	"true for extended info, false/default for basic"
//	@Success		200			{object}	api.ListResponse			"Basic follow list"
//	@Success		200			{object}	api.ListExtendedResponse	"Extended follow list"
//	@Failure		400			{object}	handler.ErrorResponse		"Missing follower parameter"
//	@Failure		500			{object}	handler.ErrorResponse		"Internal server error"
//	@Router			/follows [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "getting follow list"

	ctx := r.Context()
	follower := r.URL.Query().Get("follower")

	if follower == "" {
		handler.Error(h.log, w, op, errNoFollower, http.StatusBadRequest, errNoFollower.Error())
		return
	}

	extended := r.URL.Query().Get("extended")
	if extended == "true" {
		extendedList, err := h.r.ListExtended(ctx, follower)
		if err != nil {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		following := make([]api.ListExtendedResponseItem, len(extendedList))
		for i, item := range extendedList {
			following[i] = api.ListExtendedResponseItem(item)
		}

		json.NewEncoder(w).Encode(api.ListExtendedResponse{FollowList: following})
		return
	}

	followList, err := h.r.List(ctx, follower)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	following := make([]api.ListResponseItem, len(followList))
	for i, item := range followList {
		following[i] = api.ListResponseItem(item)
	}

	json.NewEncoder(w).Encode(api.ListResponse{FollowList: following})
}

// Post godoc
//
//	@Summary		Follow or unfollow user
//	@Description	Follow a user (or unfollow if already following)
//	@Tags			Follows
//	@Accept			json
//	@Security		BearerAuth
//	@Param			username	path		string			true	"Authenticated user's username"
//	@Param			request		body		api.PostRequest	true	"User to follow/unfollow"
//	@Success		204			{object}	nil
//	@Failure		401			{object}	handler.ErrorResponse	"Unauthorized - invalid claims or identity mismatch"
//	@Failure		400			{object}	handler.ErrorResponse	"Invalid request"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/follows/{username} [post]
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	const op = "following user"

	username := r.PathValue("username")
	ctx := r.Context()
	claims := ctx.Value(auth.AuthContextKey{})
	user, ok := claims.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusUnauthorized, handler.MsgIdentity)
		return
	}

	if user.Username != username {
		handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusUnauthorized, handler.MsgIdentity)
		return
	}

	var req api.PostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	err = h.r.Follow(ctx, username, req.Follow)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete godoc
//
//	@Summary		Unfollow user
//	@Description	Unfollow a specific user
//	@Tags			Follows
//	@Accept			json
//	@Security		BearerAuth
//	@Param			username	path		string				true	"Authenticated user's username"
//	@Param			request		body		api.DeleteRequest	true	"User to unfollow"
//	@Success		204			{object}	nil
//	@Failure		401			{object}	handler.ErrorResponse	"Unauthorized - invalid claims or identity mismatch"
//	@Failure		400			{object}	handler.ErrorResponse	"Invalid request"
//	@Failure		500			{object}	handler.ErrorResponse	"Internal server error"
//	@Router			/follows/{username} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "unfollowing user"

	username := r.PathValue("username")
	ctx := r.Context()
	claims := ctx.Value(auth.AuthContextKey{})
	user, ok := claims.(*auth.Claims)

	if !ok {
		handler.Error(h.log, w, op, handler.ErrClaims, http.StatusUnauthorized, handler.MsgIdentity)
		return
	}

	if user.Username != username {
		handler.Error(h.log, w, op, handler.ErrIdentity, http.StatusUnauthorized, handler.MsgIdentity)
		return
	}

	var req api.DeleteRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusBadRequest, handler.MsgRequest)
		return
	}

	err = h.r.Unfollow(ctx, username, req.Unfollow)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
