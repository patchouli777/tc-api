package follow

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	f "main/pkg/api/model/follow"
	"net/http"
)

type Service interface {
	IsFollower(ctx context.Context, follower, followed string) (bool, error)
	List(ctx context.Context, follower string) ([]FollowerListItem, error)
	ListExtended(ctx context.Context, follower string) ([]FollowingListExtendedItem, error)
	Follow(ctx context.Context, follower, followed string) error
	Unfollow(ctx context.Context, unfollower, unfollowed string) error
}

type Handler struct {
	s   Service
	log *slog.Logger
}

func NewHandler(log *slog.Logger, s Service) *Handler {
	return &Handler{s: s, log: log}
}

// Get godoc
// @Summary      Check follow status
// @Description  Check if one user follows another
// @Tags         Follows
// @Accept       json
// @Produce      json
// @Param        username  path     string  true  "Follower username"  min(1)
// @Param        followed  query    string  true  "Followed username"  min(1)
// @Success      200       {object}  f.GetResponse
// @Failure      400       {object}  handler.ErrorResponse  "Missing followed parameter"
// @Failure      500       {object}  handler.ErrorResponse  "Internal server error"
// @Router       /follows/{username} [get]
func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "getting follow status"

	follower := r.PathValue("username")
	ctx := r.Context()

	followed := r.URL.Query().Get("followed")
	if followed == "" {
		handler.Error(h.log, w, op, errNoFollowed, http.StatusBadRequest, errNoFollowed.Error())
		return
	}

	isFollower, err := h.s.IsFollower(ctx, follower, followed)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	json.NewEncoder(w).Encode(f.GetResponse{IsFollower: isFollower})
}

// List godoc
// @Summary      List user's follows
// @Description  Get list of users that a follower is following (basic or extended)
// @Tags         Follows
// @Accept       json
// @Produce      json
// @Param        follower  query    string  true  "Follower username"  min(1)
// @Param        extended  query    string  false "Include extended data (true/false)"  Enums(true, false)
// @Success      200  {object}  f.ListResponse          "Basic follow list"
// @Success      200  {object}  f.ListExtendedResponse  "Extended follow list"
// @Failure      400  {object}  handler.ErrorResponse  "Missing follower parameter"
// @Failure      500  {object}  handler.ErrorResponse  "Internal server error"
// @Router       /follows [get]
func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	const op = "getting follow list"

	ctx := r.Context()
	follower := r.URL.Query().Get("follower")

	if follower == "" {
		handler.Error(h.log, w, op, errNoFollower, http.StatusBadRequest, errNoFollower.Error())
		return
	}

	extended := r.URL.Query().Get("extended")
	if extended == "true" {
		extendedList, err := h.s.ListExtended(ctx, follower)
		if err != nil {
			handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
			return
		}

		following := make([]f.ListExtendedResponseItem, len(extendedList))
		for i, item := range extendedList {
			following[i] = f.ListExtendedResponseItem(item)
		}

		json.NewEncoder(w).Encode(f.ListExtendedResponse{FollowList: following})
		return
	}

	followList, err := h.s.List(ctx, follower)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	following := make([]f.ListResponseItem, len(followList))
	for i, item := range followList {
		following[i] = f.ListResponseItem(item)
	}

	json.NewEncoder(w).Encode(f.ListResponse{FollowList: following})
}

// Post godoc
// @Summary      Follow a user
// @Description  Current user follows another user
// @Tags         Follows
// @Accept       json
// @Produce      json
// @Param        username   path     string  true  "Username to follow for (must match auth user)"  min(1)
// @Param        auth       header   string  true  "Bearer token"  format(jwt)
// @Param        following  query    string  true  "Username to follow"  min(1)
// @Security     BearerAuth
// @Success      204  "User followed successfully"
// @Failure      400  {object}  handler.ErrorResponse  "Invalid auth or username mismatch"
// @Failure      500  {object}  handler.ErrorResponse  "Internal server error"
// @Router       /follows/{username} [post]
func (h Handler) Post(w http.ResponseWriter, r *http.Request) {
	const op = "following user"

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

	following := r.URL.Query().Get("following")

	err := h.s.Follow(ctx, username, following)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete godoc
// @Summary      Unfollow a user
// @Description  Current user stops following another user
// @Tags         Follows
// @Accept       json
// @Produce      json
// @Param        username    path     string  true  "Username to unfollow for (must match auth user)"  min(1)
// @Param        auth        header   string  true  "Bearer token"  format(jwt)
// @Param        unfollowing query    string  true  "Username to unfollow"  min(1)
// @Security     BearerAuth
// @Success      204  "User unfollowed successfully"
// @Failure      400  {object}  handler.ErrorResponse  "Invalid auth or username mismatch"
// @Failure      500  {object}  handler.ErrorResponse  "Internal server error"
// @Router       /follows/{username} [delete]
func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "unfollowing user"

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

	unfollowing := r.URL.Query().Get("unfollowing")
	err := h.s.Unfollow(ctx, username, unfollowing)
	if err != nil {
		handler.Error(h.log, w, op, err, http.StatusInternalServerError, handler.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
