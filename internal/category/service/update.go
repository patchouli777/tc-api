package category

import (
	"context"
	"fmt"
	"log/slog"
	d "main/internal/category/domain"
	"main/internal/lib/sl"
	lsd "main/internal/livestream/domain"
	"math"
	"time"
)

type livestreamLister interface {
	List(ctx context.Context, s lsd.LivestreamSearch) ([]lsd.Livestream, error)
}

type listerUpdater interface {
	List(ctx context.Context, f d.CategoryFilter) ([]d.Category, error)
	UpdateViewers(ctx context.Context, id int, viewers int) error
}

type CategoryUpdater struct {
	lsLister livestreamLister
	lu       listerUpdater
	log      *slog.Logger
}

func NewUpdater(log *slog.Logger, lsLister livestreamLister, lu listerUpdater) *CategoryUpdater {
	return &CategoryUpdater{lsLister: lsLister, lu: lu, log: log}
}

// TODO: use asynq
func (c *CategoryUpdater) Update(ctx context.Context, timeout time.Duration) {
	const op = "category.Updater.Update"

	go func() {
		for {
			timeoutCtx, cancelTimeout := context.WithCancel(ctx)

			categories, err := c.lu.List(timeoutCtx, d.CategoryFilter{
				Page:  1,
				Count: 10000,
				Sort:  "desc",
			})
			if err != nil {
				c.log.Error("list categories", sl.Err(err), sl.Op(op))
			}

			for _, cat := range categories {
				go func() {
					lsArr, err := c.lsLister.List(timeoutCtx, lsd.LivestreamSearch{
						Category: cat.Link, Page: 1, Count: 9999})
					if err != nil {
						c.log.Error("list livestreams", sl.Err(err), sl.Op(op), slog.String("category", cat.Name))
						return
					}

					var viewers int32 = 0

					for i := range int(math.Min(100, float64(len(lsArr)))) {
						viewers += int32(lsArr[i].Viewers)
					}

					err = c.lu.UpdateViewers(timeoutCtx, int(cat.Id), int(viewers))
					if err != nil {
						c.log.Error("update viewers", sl.Err(err), sl.Op(op), slog.String("category", cat.Name))
					}
				}()
			}

			c.log.Info(fmt.Sprintf("Categories updated. Next in %v", timeout))
			select {
			case <-ctx.Done():
				c.log.Info("category updating ended")
				cancelTimeout()
				return
			case <-time.After(timeout):
				cancelTimeout()
			}
		}
	}()
}
