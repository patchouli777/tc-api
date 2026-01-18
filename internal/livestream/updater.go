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

type CreaterUpdater interface {
	Creater
	Updater
}

type UpdateScheduler struct {
	log   *slog.Logger
	ssa   *streamserver.Adapter
	lsr   CreaterUpdater
	sched *asynq.Scheduler

	previews   []os.DirEntry
	instanceID string
	rdb        *redis.Client
}

const (
	TypeLivestreamUpdate = "livestream:update"
)

type livestreamTaskPayload struct {
	LivestreamID int
	Username     string
}

// TODO: ttl
func (s *UpdateScheduler) Run(ctx context.Context, timeout time.Duration) {
	// if err := s.Register(ctx); err != nil {
	// 	s.log.Error("register failed", sl.Err(err))
	// }

	// go func() {
	// 	err := s.PollSRS(ctx)
	// 	if err != nil {
	// 		s.log.Error("poll srs", sl.Err(err))
	// 	}
	// }()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(timeout):
			if err := s.Register(ctx); err != nil {
				s.log.Error("register failed", sl.Err(err))
			}

			go func() {
				err := s.PollSRS(ctx)
				if err != nil {
					s.log.Error("poll srs", sl.Err(err))
				}
			}()
		}
	}
}

// At startup, each Go instance registers its instance ID in Redis (set go_poller:<instance_id> with TTL 60s, heartbeat every 30s).
// Discover other instances via SCAN on go_poller:* keys to coordinate SRS assignments.
func (s *UpdateScheduler) Register(ctx context.Context) error {
	return s.rdb.Set(ctx, s.pollerKey(), time.Now().Unix(), 60*time.Second).Err()
}

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
					s.log.Debug("livestream already started",
						slog.String("username", st.Name),
						slog.String("livestream_id", st.ID))
					continue
				}
				s.log.Error("creating livestream",
					sl.Err(err),
					slog.String("username", st.Name),
					slog.String("livestream_id", st.ID))
				continue
			}

			payload, err := json.Marshal(livestreamTaskPayload{LivestreamID: ls.Id, Username: ls.UserName})
			if err != nil {
				s.log.Error("error creating payoload", sl.Err(err))
				return err
			}

			_, err = s.sched.Register("@every 15s", asynq.NewTask(TypeLivestreamUpdate, payload))
			if err != nil {
				s.log.Error("error registering new task", sl.Err(err))
				return err
			}
		}
	}

	return nil
}

func NewUpdateScheduler(log *slog.Logger,
	rdb *redis.Client,
	ssa *streamserver.Adapter,
	lsr CreaterUpdater,
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

func (s *UpdateScheduler) pollerKey() string {
	return fmt.Sprintf("go_poller:%s", s.instanceID)
}

// func (s *UpdateScheduler) NewUpdateTask(id, username string) (*asynq.Task, error) {
// 	payload, err := json.Marshal(livestreamTaskPayload{LivestreamID: id, Username: username})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return asynq.NewTask(TypeLivestreamUpdate, payload), nil
// }

func (s *UpdateScheduler) HandleUpdateTask(ctx context.Context, t *asynq.Task) error {
	var p livestreamTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	// s.log.Info("big handling of update task",
	// 	slog.String("user", p.Username),
	// 	slog.Int("id", p.LivestreamID))

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
