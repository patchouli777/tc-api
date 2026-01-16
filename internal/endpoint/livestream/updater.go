package livestream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"main/internal/external/streamserver"
	"main/internal/lib/sl"
	"math/rand"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

type CreaterUpdater interface {
	Creater
	Updater
}

type UpdateScheduler struct {
	log   *slog.Logger
	ssa   *streamserver.Adapter
	lsr   CreaterUpdater
	sched *asynq.Scheduler
}

func NewUpdateScheduler(log *slog.Logger, ssa *streamserver.Adapter, lsr CreaterUpdater, sched *asynq.Scheduler) *UpdateScheduler {
	return &UpdateScheduler{
		log:   log,
		ssa:   ssa,
		lsr:   lsr,
		sched: sched}
}

func (q *UpdateScheduler) Update(ctx context.Context, timeout time.Duration) {
	for {
		resp, err := q.ssa.List(ctx)
		if err != nil {
			q.log.Error("Unable to get a response from streaming server. Cancelling updates for livestreams.", sl.Err(err))
			return
		}

		entries, err := os.ReadDir("./static/livestreamthumbs")
		if err != nil {
			q.log.Error("Unable to read ./static/livestreamthumbs. Cancelling updates for livestreams.", sl.Err(err))
			return
		}

		// TODO: currently all livestreams updated every 25s for simplicity
		// implement single-stream updating with timewheel
		// also currently stream server is not serving livestream thumbs
		// but it should (apparently) !!
		for _, st := range resp.Streams {
			ls, err := q.lsr.Create(ctx, LivestreamCreate{
				Category: "apex",
				Title:    fmt.Sprintf("livestream of %s", st.Name),
				Username: st.Name,
			})
			if err != nil {
				q.log.Error("error creating livestreams", sl.Err(err))
			}

			thumbnailId := rand.Intn(len(entries))
			thumbnail := entries[thumbnailId].Name()

			payload, err := json.Marshal(livestreamTaskPayload{LivestreamID: ls.Id, Username: ls.UserName})
			if err != nil {
				q.log.Error("error creating payoload", sl.Err(err))
				return
			}

			entryId, err := q.sched.Register("@every 15s", asynq.NewTask(TypeLivestreamUpdate, payload))
			if err != nil {
				q.log.Error("error registering new task", sl.Err(err))
				return
			}

			q.log.Info("entry id", slog.String("entry id", entryId))

			// lsr.UpdateViewers(ctx, st.Name, int32(st.Clients))
			q.lsr.UpdateThumbnail(ctx, st.Name, "livestreamthumbs/"+thumbnail)
		}

		q.log.Info(fmt.Sprintf("Viewers count updated. Next in %s.", timeout.String()))
		time.Sleep(timeout)
	}
}

const (
	TypeLivestreamUpdate = "livestream:update"
)

type livestreamTaskPayload struct {
	LivestreamID int32
	Username     string
}

func (q *UpdateScheduler) NewUpdateTask(id int32, username string) (*asynq.Task, error) {
	payload, err := json.Marshal(livestreamTaskPayload{LivestreamID: id, Username: username})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeLivestreamUpdate, payload), nil
}

func (q *UpdateScheduler) HandleUpdateTask(ctx context.Context, t *asynq.Task) error {
	var p livestreamTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	q.log.Info("big handling of update task", slog.String("user", p.Username), slog.Int("id", int(p.LivestreamID)))

	resp, err := q.ssa.Get(ctx, p.Username)
	if err != nil {
		return err
	}

	err = q.lsr.UpdateViewers(ctx, p.Username, int32(resp.Stream.Clients))
	if err != nil {
		return fmt.Errorf("user: %s. %w", p.Username, err)
	} else {
		log.Printf("updated %s with %d. full response: %+v\n", p.Username, int32(resp.Stream.Clients), resp)
	}

	return nil
}
