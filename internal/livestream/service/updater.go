package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/external/streamserver"
	"main/internal/lib/sl"
	d "main/internal/livestream/domain"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Store interface {
	Create(ctx context.Context, cr d.LivestreamCreate) (*d.Livestream, error)
	Update(ctx context.Context, id int, upd d.LivestreamUpdate) (*d.Livestream, error)
	List(ctx context.Context, s d.LivestreamSearch) ([]d.Livestream, error)
	UpdateViewers(ctx context.Context, id int, viewers int) error
	UpdateThumbnail(ctx context.Context, id int, thumbnail string) error
}

// TODO: add polling count, polling timeout to config
type UpdateScheduler struct {
	log   *slog.Logger
	ssa   *streamserver.Adapter
	lsr   Store
	sched *asynq.Scheduler

	previews   []os.DirEntry
	instanceID string
	rdb        *redis.Client
}

func NewUpdateScheduler(log *slog.Logger,
	rdb *redis.Client,
	ssa *streamserver.Adapter,
	lsr Store,
	sched *asynq.Scheduler,
	instanceID string) *UpdateScheduler {
	entries, err := os.ReadDir("./static/livestreamthumbs")
	if err != nil {
		log.Error("Unable to read ./static/livestreamthumbs. Cancelling updates for livestreams.", sl.Err(err))
		return nil
	}

	log = log.With(slog.String("instance_id", instanceID))

	return &UpdateScheduler{
		log:        log,
		ssa:        ssa,
		lsr:        lsr,
		sched:      sched,
		previews:   entries,
		instanceID: instanceID,
		rdb:        rdb}
}

// TODO: ttl
func (s *UpdateScheduler) Run(ctx context.Context, timeout time.Duration) {
	err := s.startup(ctx)
	if err != nil {
		s.log.Info("startup", sl.Err(err))
	}

	go func() {
		err := s.PollSRS(ctx)
		if err != nil {
			s.log.Error("poll srs", sl.Err(err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(timeout):
			go func() {
				err := s.PollSRS(ctx)
				if err != nil {
					s.log.Error("poll srs", sl.Err(err))
					return
				}
				s.log.Info(fmt.Sprintf("Livestreams updated. Next in %v", timeout))
			}()
		}
	}
}

const (
	TypeLivestreamUpdate = "livestream:update"
)

type livestreamTaskPayload struct {
	LivestreamID int
	Username     string
}

// polls srs and registers new update tasks
func (s *UpdateScheduler) PollSRS(ctx context.Context) error {
	lockValue := s.instanceID + ":" + strconv.FormatInt(time.Now().Unix(), 10)

	ok, err := s.rdb.SetNX(ctx, "srs_lock", lockValue, 30*time.Second).Result()
	if err != nil || !ok {
		return nil
	}

	defer func() {
		s.rdb.GetDel(ctx, "srs_lock")
	}()

	moreStreams := true
	start := 0
	count := 500
	for moreStreams {
		resp, err := s.ssa.List(ctx, start, count)
		if err != nil {
			return err
		}

		if len(resp.Streams) < count {
			moreStreams = false
		}

		for _, st := range resp.Streams {
			ls, err := s.lsr.Create(ctx, d.LivestreamCreate{Username: st.Name})
			if err != nil {
				if errors.Is(err, d.ErrAlreadyStarted) {
					// s.log.Debug("livestream already started",
					// 	slog.String("username", st.Name),
					// 	slog.String("livestream_id", st.ID))
					continue
				}
				s.log.Error("creating livestream",
					sl.Err(err),
					slog.String("username", st.Name),
					slog.String("livestream_id", st.ID))
				continue
			}

			err = s.newTask(ls)
			if err != nil {
				s.log.Error("registering new task", sl.Err(err))
			}
			start += count
		}
	}

	return nil
}

func (s *UpdateScheduler) HandleUpdateTask(ctx context.Context, t *asynq.Task) error {
	var p livestreamTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	s.log.Debug("Handling update task",
		slog.String("user", p.Username),
		slog.Int("id", p.LivestreamID))

	resp, err := s.ssa.Get(ctx, p.Username)
	if err != nil {
		return err
	}

	err = s.lsr.UpdateViewers(ctx, p.LivestreamID, resp.Stream.Clients)
	if err != nil {
		s.log.Error("updating viewers",
			sl.Err(err),
			slog.Int("livestream_id", p.LivestreamID))
		return err
	} else {
		s.log.Debug("viewers updated",
			slog.Int("livestream_id", p.LivestreamID),
			slog.String("username", p.Username),
			slog.Int("viewers", resp.Stream.Clients))
	}

	thumbnailId := rand.Intn(len(s.previews))
	thumbnail := s.previews[thumbnailId].Name()
	err = s.lsr.UpdateThumbnail(ctx, p.LivestreamID, "livestreamthumbs/"+thumbnail)
	if err != nil {
		return err
	}

	return nil
}

func (s *UpdateScheduler) startup(ctx context.Context) error {
	moreStreams := true
	page := 1
	count := 500
	for moreStreams {
		livestreams, err := s.lsr.List(ctx, d.LivestreamSearch{
			Page:  page,
			Count: count,
		})
		if err != nil {
			return err
		}

		for _, ls := range livestreams {
			s.newTask(&ls)
		}

		if len(livestreams) < count {
			moreStreams = false
		}

		page += 1
	}

	return nil
}

func (s *UpdateScheduler) newTask(ls *d.Livestream) error {
	payload, err := json.Marshal(livestreamTaskPayload{LivestreamID: ls.Id, Username: ls.UserName})
	if err != nil {
		return err
	}

	// TODO: save {entry: entryId, livestream: ls.Id} to delete task later
	// register new task with taskId = asynq.TaskID to make sure there is no duplicate update tasks for a given livestream
	_, err = s.sched.Register("@every 15s", asynq.NewTask(TypeLivestreamUpdate, payload), asynq.TaskID(strconv.Itoa(ls.Id)))
	if err != nil {
		return err
	}

	return nil
}
