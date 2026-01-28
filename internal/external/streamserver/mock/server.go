package mock

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"
	"twitchy-api/internal/app"
	"twitchy-api/internal/lib/mw"
	"twitchy-api/internal/lib/sl"
	baseclient "twitchy-api/pkg/api/client"

	"github.com/ilyakaznacheev/cleanenv"
)

func Run(ctx context.Context, log *slog.Logger) error {
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	state := serverState{streams: newRepository(), cl: baseclient.NewClient()}
	handler := handler{state: &state}

	go func() {
		for {
			for i := range state.streams.livestreams {
				viewers := rand.Intn(100)
				state.streams.livestreams[i].viewers = viewers
			}

			timeout := time.Second * 30
			fmt.Printf("Viewers count updated. Next in %+v.\n", timeout)
			time.Sleep(timeout)
		}
	}()

	root := app.GetProjectRoot()
	var cfg app.StreamServerConfig
	err := cleanenv.ReadConfig(root+"/.env", &cfg)
	if err != nil {
		log.Error("config big bad", sl.Err(err))
		return err
	}

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /streams/{id}", handler.Get)
	apiMux.HandleFunc("GET /streams", handler.List)
	apiMux.HandleFunc("POST /streams", handler.Post)
	apiMux.HandleFunc("DELETE /streams/{id}", Delete)
	apiMux.HandleFunc("POST /subscribe", handler.Subscribe)

	mainMux := http.NewServeMux()
	mainMux.Handle(cfg.Endpoint+"/", http.StripPrefix(cfg.Endpoint, apiMux))

	panicRecovery := mw.PanicRecovery(log)
	logging := mw.Logging(log)

	ssURL := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    ssURL,
		Handler: panicRecovery(mw.JSONResponse(mw.CORS(logging(mainMux)))),
	}

	fmt.Println("stream server mock running")
	return srv.ListenAndServe()
}
