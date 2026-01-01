package main

import (
	"context"
	"log/slog"
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
	cfg := app.GetConfig()

	log := app.NewLogger(cfg.Logger)
	log.With(slog.String("env", cfg.Env))
	slog.SetDefault(log)

	server := app.New(ctx, log, cfg)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", sl.Err(err))
			stop()
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
