package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"time"
	"twitchy-api/internal/external/streamserver"
	"twitchy-api/internal/lib/sl"
	d "twitchy-api/internal/livestream/domain"

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
type Updater struct {
	log        *slog.Logger
	ssa        *streamserver.Adapter
	lsr        Store
	sched      scheduler
	previews   []os.DirEntry
	instanceID string
	rdb        *redis.Client
}

func NewUpdater(log *slog.Logger,
	rdb *redis.Client,
	ssa *streamserver.Adapter,
	lsr Store,
	sched scheduler,
	instanceID string) *Updater {
	entries, err := os.ReadDir("./static/livestreamthumbs")
	if err != nil {
		log.Error("Unable to read ./static/livestreamthumbs. Cancelling updates for livestreams.", sl.Err(err))
		return nil
	}

	log = log.With(slog.String("instance_id", instanceID))

	return &Updater{
		log:        log,
		ssa:        ssa,
		lsr:        lsr,
		sched:      sched,
		previews:   entries,
		instanceID: instanceID,
		rdb:        rdb}
}

// TODO: ttl
func (s *Updater) Run(ctx context.Context, timeout time.Duration) error {
	err := s.startup(ctx)
	if err != nil {
		s.log.Info("startup", sl.Err(err))
		return err
	}

	err = s.pollSRS(ctx)
	if err != nil {
		s.log.Error("poll srs", sl.Err(err))
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(timeout):
			err := s.pollSRS(ctx)
			if err != nil {
				s.log.Error("poll srs", sl.Err(err))
				return err
			}
		}
	}
}

const (
	TaskUpdate = "livestream:update"
)

type updateTaskPayload struct {
	LivestreamID int
	Username     string
}

// polls srs and registers new update tasks
func (s *Updater) pollSRS(ctx context.Context) error {
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
		s.log.Debug("polling srs",
			slog.Int("start", start),
			slog.Int("count", count))

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

func (s *Updater) HandleUpdateTask(ctx context.Context, payload []byte) error {
	var p updateTaskPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
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

func (s *Updater) startup(ctx context.Context) error {
	moreStreams := true
	page := 1
	count := 500
	for moreStreams {
		livestreams, err := s.lsr.List(ctx, d.LivestreamSearch{
			Category: "dota-2",
			Page:     page,
			Count:    count,
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

type scheduler interface {
	Schedule(every time.Duration, taskType string, payload []byte, taskId string) (string, error)
}

func (s *Updater) newTask(ls *d.Livestream) error {
	s.log.Debug("scheduling update task",
		slog.Int("livestream_id", ls.Id),
		slog.String("channel", ls.UserName))

	payload, err := json.Marshal(updateTaskPayload{LivestreamID: ls.Id, Username: ls.UserName})
	if err != nil {
		return err
	}

	// TODO: save {entry: entryId, livestream: ls.Id} to delete task later
	// register new task with taskId = asynq.TaskID to make sure there is no duplicate update tasks for a given livestream
	_, err = s.sched.Schedule(time.Second*15, TaskUpdate, payload, strconv.Itoa(ls.Id))
	if err != nil {
		return err
	}

	return nil
}
