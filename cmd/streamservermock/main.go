package main

import (
	"fmt"
	"log/slog"
	"main/internal/lib/mw"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// TODO: dockerize mock (xd)
func main() {
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

	srv := &http.Server{
		Addr:    "127.0.0.1:1985",
		Handler: mw.PanicRecovery(mw.JSONResponse(mw.CORS(mw.Logging(mainMux)))),
	}

	fmt.Println("stream server mock running")
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
