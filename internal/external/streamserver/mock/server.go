package streamservermock

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/lib/mw"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	repo := newRepository()

	go func() {
		for {
			for i := range repo.livestreams {
				viewers := rand.Intn(100)
				repo.livestreams[i].viewers = viewers
			}

			fmt.Println("Viewers count updated. Next in 30s.")
			time.Sleep(time.Second * 30)
		}
	}()

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /streams/{id}", Get(repo))
	apiMux.HandleFunc("GET /streams", List(repo))
	apiMux.HandleFunc("POST /streams", Post(repo))
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
