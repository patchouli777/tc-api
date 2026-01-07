package follow

import (
	"context"
	"encoding/json"
	"log/slog"
	"main/internal/auth"
	"main/internal/lib/handler"
	f "main/pkg/api/follow"
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
// @Summary is user a follower of other user
// @Description Returns json response
// @Tags follow
// @Param username path string true "follower"
// @Param followed query string true "followed"
// @Produce json
// @Success 200 {object} GetResponse
// @Failure 400 {object} er.RequestError "Bad Request"
// @Failure 500 {object} er.RequestError "Internal Server Error"
// @Router /follow/{username} [get]
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
// @Summary Retrieve follower's follow list
// @Description Lists the followees of a given follower username. If `extended=true` is passed as a query parameter, returns detailed info per followee.
// @Tags follow
// @Accept json
// @Produce json
// @Param follower query string true "Follower username whose follow list is requested"
// @Param extended query string false "If set to 'true', returns extended follow list details"
// @Success 200 {object} ListResponse "Basic follow list response"
// @Success 200 {object} ListExtendedResponse "Extended follow list response"
// @Failure 400 {object} er.RequestError "Bad request: follower username is missing"
// @Failure 500 {object} er.RequestError "Internal server error while retrieving follow list"
// @Router /follows [get]
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
// @Summary Follow a user
// @Description Allows the authenticated user to follow another user by username
// @Tags follow
// @Accept json
// @Produce json
// @Param username path string true "Username of the user to follow"
// @Param following query string true "Username of the user to be followed"
// @Success 200 {string} string "Follow successful"
// @Failure 400 {object} er.RequestError "Bad Request: user mismatch"
// @Failure 500 {object} er.RequestError "Internal Server Error: failed to follow user"
// @Router /follow/{username} [post]
func (h Handler) Post(w http.ResponseWriter, r *http.Request) {
	// TODO: update doc
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
// @Summary Unfollow a user
// @Description Allows the authenticated user to unfollow another user by username
// @Tags follow
// @Accept json
// @Produce json
// @Param username path string true "Username of the user to unfollow"
// @Param unfollowing query string true "Username of the user to be unfollowed"
// @Success 200 {string} string "Unfollow successful"
// @Failure 400 {object} er.RequestError "Bad Request: user mismatch"
// @Failure 500 {object} er.RequestError "Internal Server Error: failed to unfollow user"
// @Router /follow/{username} [delete]
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
