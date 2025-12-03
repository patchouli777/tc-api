package category

import (
	"context"
	"main/internal/endpoint/livestream"
)

type CategoryUpdater interface {
	Updater
	Lister
}

type LivestreamUpdater interface {
	LivestreamLister
	UpdateViewers(ctx context.Context, user string, viewers int32) error
	UpdateThumbnail(ctx context.Context, user, thumbnail string) error
}

type LivestreamLister interface {
	List(ctx context.Context, catLink string, page, count int) ([]livestream.Livestream, error)
}

// TODO: update
// func UpdateCategories(
// 	ctx context.Context,
// 	log *slog.Logger,
// 	timeout time.Duration,
// 	catUpdater CategoryUpdater,
// 	ll LivestreamLister) {
// 	for {
// 		timeoutCtx, cancelTimeout := context.WithCancel(ctx)

// 		categories, err := catUpdater.List(timeoutCtx, &CategoryFilter{
// 			Page:  1,
// 			Count: 10000,
// 			Sort:  "desc",
// 		})
// 		if err != nil {
// 			log.Error("error getting categories while updating categories", sl.Err(err))
// 		}

// 		for _, cat := range categories {
// 			go func() {
// 				lsArr, err := ll.List(timeoutCtx, cat.Link, 1, 99999)
// 				if err != nil {
// 					log.Error("error getting livestreams while updating categories", sl.Err(err), slog.Attr{
// 						Key:   "category name",
// 						Value: slog.StringValue(cat.Name),
// 					})
// 					return
// 				}

// 				var viewers int32 = 0

// 				// NOTE: для обновления категории беруется первые по зрителям 100 трансляций
// 				// (или меньше если их всего меньше 100)
// 				for i := range int(math.Min(100, float64(len(lsArr)))) {
// 					viewers += lsArr[i].Viewers
// 				}

// 				err = catUpdater.UpdateViewers(timeoutCtx, cat.Link, viewers)
// 				if err != nil {
// 					log.Error("error updating category", sl.Err(err), slog.Attr{
// 						Key:   "category name",
// 						Value: slog.StringValue(cat.Name),
// 					})
// 				}
// 			}()
// 		}

// 		log.Info(fmt.Sprintf("categories data updated. next in %s", timeout.String()))
// 		select {
// 		case <-ctx.Done():
// 			log.Info("category updating ended")
// 			cancelTimeout()
// 			return
// 		case <-time.After(timeout):
// 			cancelTimeout()
// 			log.Info("category updating started")
// 		}
// 	}
// }
