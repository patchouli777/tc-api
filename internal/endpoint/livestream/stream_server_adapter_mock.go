package livestream

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/redis/go-redis/v9"
)

type livestreamData struct {
	channel string
	viewers int
}

type StreamServerAdapterMock struct {
	livestreams []livestreamData
	wheel       *timingwheel.TimingWheel
	lsr         Repository
	InstanceID  string
	log         *slog.Logger
}

func NewStreamServerAdapterMock(log *slog.Logger, rdb *redis.Client, lsr Repository,
	InstanceID string) *StreamServerAdapterMock {
	wheel := timingwheel.NewTimingWheel(5*time.Second, 16)

	livestreams, err := lsr.ListAll(context.Background())
	if err != nil {
		log.Error("unable to list all livestreams", sl.Err(err))
	}

	livestreamsLen := len(livestreams)
	lss := make([]livestreamData, livestreamsLen)
	for i := range livestreamsLen {
		lss[i] = livestreamData{channel: livestreams[i].User.Name, viewers: int(livestreams[i].Viewers)}
	}

	return &StreamServerAdapterMock{wheel: wheel,
		log:         log,
		lsr:         lsr,
		InstanceID:  InstanceID,
		livestreams: lss}
}

func (u *StreamServerAdapterMock) Update(ctx context.Context) {
	go func() {
		for {
			u.log.Info("livestreams mock update started")

			entries, err := os.ReadDir("./static/livestreamthumbs")
			if err != nil {
				fmt.Printf("error reading ./static/livestreamthumbs: %v\n", err)
				return
			}

			for i, st := range u.livestreams {
				thumbnailId := rand.Intn(len(entries))
				thumbnail := entries[thumbnailId].Name()
				// TODO: pipe

				viewers := rand.Intn(100)
				u.livestreams[i].viewers = viewers
				u.lsr.UpdateViewers(ctx, st.channel, int32(st.viewers))
				u.lsr.UpdateThumbnail(ctx, st.channel, "livestreamthumbs/"+thumbnail)
			}

			u.log.Info("livestreams mock update ended. next in 25s")
			time.Sleep(time.Second * 25)
		}
	}()
}

// TODO: sim error
func (u *StreamServerAdapterMock) Start(ctx context.Context, username string) error {
	u.livestreams = append(u.livestreams, livestreamData{channel: username, viewers: 0})
	return nil
}

func (u *StreamServerAdapterMock) End(ctx context.Context, username string) error {
	for i, ls := range u.livestreams {
		if username != ls.channel {
			continue
		}

		u.livestreams = slices.Delete(u.livestreams, i, i+1)
		break
	}

	return nil
}

func (u *StreamServerAdapterMock) List(ctx context.Context) (*StreamServerResponse, error) {
	return nil, errors.New("not implemented")
}
