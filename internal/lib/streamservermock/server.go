package streamservermock

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"main/internal/lib/mw"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	streamServer := NewStreamServerMock(log)

	go func() {
		for {
			for i := range streamServer.livestreams {
				viewers := rand.Intn(100)
				streamServer.livestreams[i].viewers = viewers
			}

			fmt.Println("Viewers count updated. Next in 30s.")
			time.Sleep(time.Second * 30)
		}
	}()

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /streams/{id}", Get(streamServer))
	apiMux.HandleFunc("GET /streams", List(streamServer))
	apiMux.HandleFunc("POST /streams", Post(streamServer))
	apiMux.HandleFunc("DELETE /streams/{id}", Delete)

	mainMux := http.NewServeMux()
	mainMux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiMux))

	panicRecovery := mw.PanicRecovery(log)
	logging := mw.Logging(log)

	srv := &http.Server{
		Addr:    "127.0.0.1:1985",
		Handler: panicRecovery(mw.JSONResponse(mw.CORS(logging(mainMux)))),
	}

	fmt.Println("stream server mock running")
	return srv.ListenAndServe()
}

type StreamServerMock struct {
	livestreams []livestream
	log         *slog.Logger
}

func NewStreamServerMock(log *slog.Logger) *StreamServerMock {
	return &StreamServerMock{log: log,
		livestreams: []livestream{}}
}

func (u *StreamServerMock) List(ctx context.Context) ([]livestream, error) {
	return u.livestreams, nil
}

func (u *StreamServerMock) Get(ctx context.Context, id string) (*livestream, error) {
	for _, ls := range u.livestreams {
		if ls.channel == id {
			copy := ls
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (u *StreamServerMock) Start(ctx context.Context, username string) error {
	u.livestreams = append(u.livestreams, livestream{channel: username, viewers: 0})
	return nil
}

func (u *StreamServerMock) End(ctx context.Context, username string) error {
	for i, ls := range u.livestreams {
		if username != ls.channel {
			continue
		}

		u.livestreams = slices.Delete(u.livestreams, i, i+1)
		break
	}

	return nil
}
