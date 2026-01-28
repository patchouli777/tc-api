package tag

import (
	"context"
	"log/slog"
	"net/http"
	d "twitchy-api/internal/category/domain"
)

type Getter interface {
	Get(ctx context.Context, id int) (*d.Category, error)
}

type Lister interface {
	List(ctx context.Context, f d.CategoryFilter) ([]d.Category, error)
}

type Creater interface {
	Create(ctx context.Context, cat d.CategoryCreate) error
}

type Updater interface {
	Update(ctx context.Context, id int32, cat d.CategoryUpdate) error
}

type Deleter interface {
	Delete(ctx context.Context, id int32) error
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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {}
