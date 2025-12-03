package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main/internal/app"
	"main/internal/lib/sl"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	conf := app.GetConfig()
	log := conf.Log

	server := app.New(ctx, conf)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", sl.Err(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")
	shutdownCtx, cancelTimeout := context.WithTimeout(ctx, 3*time.Second)
	defer cancelTimeout()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("can't shutdown gracefully", sl.Err(err))
		stop()
	}

	log.Info("server shut down gracefully")
}
