package category

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"
	d "twitchy-api/internal/category/domain"
	"twitchy-api/internal/lib/sl"
	lsd "twitchy-api/internal/livestream/domain"
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

func (c *CategoryUpdater) Run(ctx context.Context, timeout time.Duration) error {
	const op = "category.Updater.Run"

	for {
		categories, err := c.lu.List(ctx, d.CategoryFilter{
			Page:  1,
			Count: 10000,
			Sort:  "desc",
		})
		if err != nil {
			c.log.Error("list categories", sl.Err(err), sl.Op(op))
		}

		for _, cat := range categories {
			go func() {
				lsArr, err := c.lsLister.List(ctx, lsd.LivestreamSearch{
					Category: cat.Link, Page: 1, Count: 9999})
				if err != nil {
					c.log.Error("list livestreams",
						slog.String("category", cat.Name),
						sl.Err(err),
						sl.Op(op))
					return
				}

				var viewers int32 = 0

				for i := range int(math.Min(100, float64(len(lsArr)))) {
					viewers += int32(lsArr[i].Viewers)
				}

				err = c.lu.UpdateViewers(ctx, int(cat.Id), int(viewers))
				if err != nil {
					c.log.Error("update viewers",
						slog.String("category", cat.Name),
						sl.Err(err),
						sl.Op(op))
				}
			}()
		}

		c.log.Info(fmt.Sprintf("Categories updated. Next in %v", timeout))
		select {
		case <-ctx.Done():
			c.log.Info("category updating ended")
			return nil
		case <-time.After(timeout):
			return nil
		}
	}
}
