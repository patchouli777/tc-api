package category

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/endpoint/livestream"
	"main/internal/lib/sl"
	"math"
	"time"
)

type LivestreamLister interface {
	List(ctx context.Context, category string, page, count int) ([]livestream.Livestream, error)
}

type CategoryUpdater struct {
	lsr LivestreamLister
	cr  Repository
	log *slog.Logger
}

func NewUpdater(log *slog.Logger, lsr LivestreamLister, cr Repository) *CategoryUpdater {
	return &CategoryUpdater{lsr: lsr, cr: cr, log: log}
}

func (cu *CategoryUpdater) Update(ctx context.Context, timeout time.Duration) {
	go func() {
		for {
			timeoutCtx, cancelTimeout := context.WithCancel(ctx)

			categories, err := cu.cr.List(timeoutCtx, CategoryFilter{
				Page:  1,
				Count: 10000,
				Sort:  "desc",
			})
			if err != nil {
				cu.log.Error("error getting categories while updating categories", sl.Err(err))
			}

			for _, cat := range categories {
				go func() {
					lsArr, err := cu.lsr.List(timeoutCtx, cat.Link, 1, 99999)
					if err != nil {
						cu.log.Error("error getting livestreams while updating categories", sl.Err(err), slog.Attr{
							Key:   "category name",
							Value: slog.StringValue(cat.Name),
						})
						return
					}

					var viewers int32 = 0

					// NOTE: для обновления категории беруется первые по зрителям 100 трансляций
					// (или меньше если их всего меньше 100)
					for i := range int(math.Min(100, float64(len(lsArr)))) {
						viewers += lsArr[i].Viewers
					}

					err = cu.cr.UpdateViewersById(timeoutCtx, int(cat.Id), int(viewers))
					if err != nil {
						cu.log.Error("error updating category", sl.Err(err), slog.Attr{
							Key:   "category name",
							Value: slog.StringValue(cat.Name),
						})
					}
				}()
			}

			cu.log.Info(fmt.Sprintf("categories data updated. next in %s", timeout.String()))
			select {
			case <-ctx.Done():
				cu.log.Info("category updating ended")
				cancelTimeout()
				return
			case <-time.After(timeout):
				cancelTimeout()
				cu.log.Info("category updating started")
			}
		}
	}()
}
