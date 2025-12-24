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

type updater interface {
	Updater
	Creater
}

type StreamServerAdapter struct {
	log        *slog.Logger
	endpoint   string
	wheel      *timingwheel.TimingWheel
	InstanceID string
	lsr        updater
}

func NewStreamServerAdapter(log *slog.Logger, rdb *redis.Client, lsr updater,
	InstanceID string) *StreamServerAdapter {
	wheel := timingwheel.NewTimingWheel(5*time.Second, 16)

	return &StreamServerAdapter{
		lsr:        lsr,
		log:        log,
		wheel:      wheel,
		InstanceID: InstanceID,
		endpoint:   "http://localhost:1985/api/v1/streams"}
}

func (u *StreamServerAdapter) List(ctx context.Context) (*StreamServerResponse, error) {
	cl := &http.Client{}
	response, err := cl.Get(u.endpoint)
	if err != nil {
		u.log.Error("unable to get livestreams from server", sl.Err(err))
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		u.log.Error("unable to read response from streaming server", sl.Err(err))
		return nil, err
	}

	var resp StreamServerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		u.log.Error("unable to unmarshal response from streaming server", sl.Err(err))
		return nil, err
	}

	return &resp, nil
}

func (u *StreamServerAdapter) Get(ctx context.Context) (*StreamServerResponse, error) {
	return nil, errors.New("not implemented")
}

// TODO: adapter must respect context
func (u *StreamServerAdapter) Update(ctx context.Context, timeout time.Duration) {
	go func() {
		for {
			resp, err := u.List(ctx)
			if err != nil {
				u.log.Error("Unable to get a response from streaming server. Cancelling updates for livestreams.", sl.Err(err))
				return
			}

			entries, err := os.ReadDir("./static/livestreamthumbs")
			if err != nil {
				u.log.Error("Unable to read ./static/livestreamthumbs. Cancelling updates for livestreams.", sl.Err(err))
				return
			}

			// TODO: currently all livestreams updated every 25s for simplicity
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

				// TODO: pipe
				u.lsr.UpdateViewers(ctx, st.Name, int32(st.Clients))
				u.lsr.UpdateThumbnail(ctx, st.Name, "livestreamthumbs/"+thumbnail)
			}

			u.log.Info(fmt.Sprintf("Viewers count updated. Next in %s.", timeout.String()))
			time.Sleep(timeout)
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
