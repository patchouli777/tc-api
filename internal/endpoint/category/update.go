package category

import (
	"context"
	"log/slog"
	"main/internal/endpoint/livestream"
	"main/internal/lib/sl"
	"math"
	"time"
)

type livestreamLister interface {
	List(ctx context.Context, category string, page, count int) ([]livestream.Livestream, error)
}

type listerUpdater interface {
	Lister
	ViewerUpdater
}

type CategoryUpdater struct {
	lsLister livestreamLister
	lu       listerUpdater
	log      *slog.Logger
}

func NewUpdater(log *slog.Logger, lsLister livestreamLister, lu listerUpdater) *CategoryUpdater {
	return &CategoryUpdater{lsLister: lsLister, lu: lu, log: log}
}

func (cu *CategoryUpdater) Update(ctx context.Context, timeout time.Duration) {
	const op = "category.Updater.Update"

	go func() {
		for {
			timeoutCtx, cancelTimeout := context.WithCancel(ctx)

			categories, err := cu.lu.List(timeoutCtx, CategoryFilter{
				Page:  1,
				Count: 10000,
				Sort:  "desc",
			})
			if err != nil {
				cu.log.Error("list categories", sl.Err(err), sl.Op(op))
			}

			for _, cat := range categories {
				go func() {
					lsArr, err := cu.lsLister.List(timeoutCtx, cat.Link, 1, 99999)
					if err != nil {
						cu.log.Error("list livestreams", sl.Err(err), sl.Op(op), slog.String("category", cat.Name))
						return
					}

					var viewers int32 = 0

					// NOTE: для обновления категории беруется первые по зрителям 100 трансляций
					// (или меньше если их всего меньше 100)
					for i := range int(math.Min(100, float64(len(lsArr)))) {
						viewers += lsArr[i].Viewers
					}

					err = cu.lu.UpdateViewersById(timeoutCtx, int(cat.Id), int(viewers))
					if err != nil {
						cu.log.Error("update viewers", sl.Err(err), sl.Op(op), slog.String("category", cat.Name))
					}
				}()
			}

			cu.log.Info("category data updated", slog.String("next", timeout.String()))
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
