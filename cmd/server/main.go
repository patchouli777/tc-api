package main

import (
	"context"
	"log/slog"
	"main/internal/app"
	"main/internal/lib/sl"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//	@title			Swagger Example API
//	@version		1.0
//	@description	desc
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8090
//	@BasePath	/api

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg := app.GetConfig()

	log := app.NewLogger(cfg.Logger)
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
