package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/RussellLuo/timingwheel"
	"github.com/redis/go-redis/v9"
)

type StreamServerAdapterImpl struct {
	wheel      *timingwheel.TimingWheel
	InstanceID string
	lsr        Repository
	log        *slog.Logger
}

func NewStreamServerAdapter(log *slog.Logger, rdb *redis.Client, lsr Repository,
	InstanceID string) *StreamServerAdapterImpl {
	wheel := timingwheel.NewTimingWheel(5*time.Second, 16)

	return &StreamServerAdapterImpl{
		wheel:      wheel,
		InstanceID: InstanceID}
}

func (u *StreamServerAdapterImpl) Start(ctx context.Context, username string) error {
	return errors.New("not implemented")
}

func (u *StreamServerAdapterImpl) End(ctx context.Context, username string) error {
	return errors.New("not implemented")
}

func (u *StreamServerAdapterImpl) List(ctx context.Context) (*StreamServerResponse, error) {
	return nil, errors.New("not implemented")
}

func (u *StreamServerAdapterImpl) Update(ctx context.Context) {
	go func() {
		for {
			endpoint := "http://localhost:1985/api/v1/streams/"

			cl := &http.Client{}
			response, err := cl.Get(endpoint)
			if err != nil {
				fmt.Printf("error getting livestreams from server: %v\n", err)
				return
			}
			defer response.Body.Close()

			var resp StreamServerResponse
			err = json.NewDecoder(response.Body).Decode(&resp)
			if err != nil {
				fmt.Printf("error getting livestreams from server: %v\n", err)
				return
			}

			entries, err := os.ReadDir("./static/livestreamthumbs")
			if err != nil {
				fmt.Printf("error reading ./static/livestreamthumbs: %v\n", err)
				return
			}

			for _, st := range resp.Streams {
				thumbnailId := rand.Intn(len(entries))
				thumbnail := entries[thumbnailId].Name()
				fmt.Printf("thumbnail: %s\n", thumbnail)
				// TODO: pipe
				u.lsr.UpdateViewers(ctx, st.Name, int32(st.Clients))
				u.lsr.UpdateThumbnail(ctx, st.Name, thumbnail)
			}

			time.Sleep(time.Second * 25)
		}
	}()
}

// func (u *StreamServerAdapterImpl) Subscribe(ctx context.Context) {
// 	pubsub := u.rdb.Subscribe(ctx, "updatable-channel")

// 	_, err := pubsub.Receive(ctx)
// 	if err != nil {
// 		slog.Error("receiving", sl.Err(err))
// 	}

// 	ch := pubsub.Channel()
// 	go func() {
// 		for msg := range ch {
// 			fmt.Printf("Received message from channel %s: %s\n", msg.Channel, msg.Payload)
// 		}
// 	}()
// }

// func (u *StreamServerAdapterImpl) Subscribe(ctx context.Context, eventCh <-chan EventLivestream) {
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

type StreamServerResponse struct {
	Code    int    `json:"code"`
	Server  string `json:"server"`
	Service string `json:"service"`
	Pid     string `json:"pid"`
	Streams []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Vhost     string `json:"vhost"`
		App       string `json:"app"`
		TcURL     string `json:"tcUrl"`
		URL       string `json:"url"`
		LiveMs    int64  `json:"live_ms"`
		Clients   int    `json:"clients"`
		Frames    int    `json:"frames"`
		SendBytes int    `json:"send_bytes"`
		RecvBytes int    `json:"recv_bytes"`
		Kbps      struct {
			Recv30S int `json:"recv_30s"`
			Send30S int `json:"send_30s"`
		} `json:"kbps"`
		Publish struct {
			Active bool   `json:"active"`
			Cid    string `json:"cid"`
		} `json:"publish"`
		Video struct {
			Codec   string `json:"codec"`
			Profile string `json:"profile"`
			Level   string `json:"level"`
			Width   int    `json:"width"`
			Height  int    `json:"height"`
		} `json:"video"`
		Audio struct {
			Codec      string `json:"codec"`
			SampleRate int    `json:"sample_rate"`
			Channel    int    `json:"channel"`
			Profile    string `json:"profile"`
		} `json:"audio"`
	} `json:"streams"`
}
