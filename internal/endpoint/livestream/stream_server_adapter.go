package livestream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"main/internal/lib/sl"
	"main/pkg/api/streamserver"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/RussellLuo/timingwheel"
)

type CreaterUpdater interface {
	Updater
	Creater
}

type StreamServerAdapter struct {
	log        *slog.Logger
	endpoint   string
	wheel      *timingwheel.TimingWheel
	InstanceID string
	lsr        CreaterUpdater
}

func NewStreamServerAdapter(log *slog.Logger, lsr CreaterUpdater, InstanceID string) *StreamServerAdapter {
	wheel := timingwheel.NewTimingWheel(5*time.Second, 16)

	return &StreamServerAdapter{
		lsr:        lsr,
		log:        log,
		wheel:      wheel,
		InstanceID: InstanceID,
		endpoint:   "http://localhost:1985/api/v1/streams"}
}

func (u *StreamServerAdapter) List(ctx context.Context) (*streamserver.ListResponse, error) {
	const op = "livestream.StreamServerAdapter.List"

	cl := &http.Client{}
	response, err := cl.Get(u.endpoint)
	if err != nil {
		u.log.Error("unable to get livestreams from server", sl.Err(err), slog.String("op", op))
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		u.log.Error("unable to read response from streaming server", sl.Err(err), slog.String("op", op))
		return nil, err
	}

	var resp streamserver.ListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		u.log.Error("unable to unmarshal response from streaming server", sl.Err(err), slog.String("op", op))
		return nil, err
	}

	return &resp, nil
}

func (u *StreamServerAdapter) Get(ctx context.Context, channel string) (*streamserver.GetResponse, error) {
	const op = "livestream.StreamServerAdapter.Get"

	cl := &http.Client{}
	response, err := cl.Get(u.endpoint)
	if err != nil {
		u.log.Error("unable to get livestream from server", sl.Err(err), slog.String("channel", channel), slog.String("op", op))
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		u.log.Error("unable to read response from streaming server", sl.Err(err), slog.String("op", op))
		return nil, err
	}

	var resp streamserver.GetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		u.log.Error("unable to unmarshal response from streaming server", sl.Err(err), slog.String("op", op))
		return nil, err
	}

	return &resp, nil
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
