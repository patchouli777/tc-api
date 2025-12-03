package livestream

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/redis/go-redis/v9"
)

type Updater interface {
}

type UpdaterImpl struct {
	rdb   *redis.Client
	lsr   Repository
	ur    UpdaterRepository
	wheel *timingwheel.TimingWheel
	// eventCh    <-chan EventLivestream
	InstanceID string
}

type UpdaterRepository interface {
	GetLivestreamViewers(ctx context.Context, channel string) (*int, error)
}

func NewUpdater(rdb *redis.Client,
	lsr Repository,
	ur UpdaterRepository,
	// eventCh <-chan EventLivestream,
	InstanceID string) *UpdaterImpl {
	wheel := timingwheel.NewTimingWheel(5*time.Second, 16)

	return &UpdaterImpl{rdb: rdb,
		lsr:   lsr,
		ur:    ur,
		wheel: wheel,
		// eventCh:    eventCh,
		InstanceID: InstanceID}
}

func (u *UpdaterImpl) Subscribe(ctx context.Context) {
	pubsub := u.rdb.Subscribe(ctx, "updatable-channel")

	_, err := pubsub.Receive(ctx)
	if err != nil {
		slog.Error("receiving", sl.Err(err))
	}

	ch := pubsub.Channel()
	go func() {
		for msg := range ch {
			fmt.Printf("Received message from channel %s: %s\n", msg.Channel, msg.Payload)
		}
	}()
}

// func (u *UpdaterImpl) Subscribe(ctx context.Context, eventCh <-chan EventLivestream) {
// 	u.wheel.Start()

// 	go func() {
// 		for {
// 			select {
// 			case e := <-eventCh:
// 				u.wheel.AfterFunc(2*time.Second, func() {
// 					fmt.Println("hi from timeheel")
// 				})
// 				fmt.Printf("event: type: %s, channel: %s\n", e.Type, e.Channel)
// 			case <-ctx.Done():
// 				fmt.Println("ctx done")
// 				return
// 			}
// 		}
// 	}()
// }
