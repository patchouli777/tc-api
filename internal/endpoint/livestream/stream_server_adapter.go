package livestream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"main/internal/lib/sl"
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
	endpoint   string
}

func NewStreamServerAdapter(log *slog.Logger, rdb *redis.Client, lsr Repository,
	InstanceID string) *StreamServerAdapterImpl {
	wheel := timingwheel.NewTimingWheel(5*time.Second, 16)

	return &StreamServerAdapterImpl{
		lsr:        lsr,
		log:        log,
		wheel:      wheel,
		InstanceID: InstanceID,
		endpoint:   "http://localhost:1985/api/v1/streams"}
}

func (u *StreamServerAdapterImpl) List(ctx context.Context) (*StreamServerResponse, error) {
	return nil, errors.New("not implemented")
}

func (u *StreamServerAdapterImpl) Get(ctx context.Context) (*StreamServerResponse, error) {
	return nil, errors.New("not implemented")
}

func (u *StreamServerAdapterImpl) Update(ctx context.Context) {
	go func() {
		for {
			cl := &http.Client{}
			response, err := cl.Get(u.endpoint)
			if err != nil {
				fmt.Printf("error getting livestreams from server: %v\n", err)
				return
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("Reading body failed: %v\n", err)
				return
			}

			var resp StreamServerResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				fmt.Printf("%+v", response.Body)
				fmt.Printf("error getting livestreams from server: %v\n", err)
				return
			}

			entries, err := os.ReadDir("./static/livestreamthumbs")
			if err != nil {
				fmt.Printf("error reading ./static/livestreamthumbs: %v\n", err)
				return
			}

			// TODO: currently all livestream updated every 25s for simplicity
			// implement single-stream updating with timewheel
			// also currently stream server is not serving livestream thumbs
			// but it should (apparently) !!
			for _, st := range resp.Streams {
				_, err := u.lsr.Create(ctx, LivestreamCreate{CategoryLink: "apex",
					Title:    fmt.Sprintf("livestream of %s", st.Name),
					Username: st.Name,
				})
				if err != nil {
					u.log.Error("error", sl.Err(err), slog.String("user", st.Name))
				}

				thumbnailId := rand.Intn(len(entries))
				thumbnail := entries[thumbnailId].Name()

				fmt.Printf("thumbnail: %s\n", thumbnail)
				// TODO: pipe
				u.lsr.UpdateViewers(ctx, st.Name, int32(st.Clients))
				u.lsr.UpdateThumbnail(ctx, st.Name, "livestreamthumbs/"+thumbnail)
			}

			u.log.Info("Viewers count updated. Next in 25s.")
			time.Sleep(time.Second * 25)
		}
	}()
}

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
