package app

import (
	"log/slog"
	authUtils "main/internal/auth"
	"main/internal/endpoint/auth"
	"main/internal/endpoint/category"
	"main/internal/endpoint/channel"
	"main/internal/endpoint/follow"
	"main/internal/endpoint/health"
	"main/internal/endpoint/livestream"
	"main/internal/endpoint/user"
	"net/http"

	_ "main/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           twitchclone api
// @version         1.0
// @description     yes
// @termsOfService  no

// @host      localhost:8090
// @BasePath  /api

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func addRoutes(mux *http.ServeMux,
	log *slog.Logger,
	cs category.Repository,
	lss livestream.Service,
	chs channel.Service,
	as auth.Service,
	fs follow.Service,
	us user.Service) {
	// TODO: mock mw
	var authMiddleware = authUtils.AuthMiddleware
	apiMux := http.NewServeMux()

	livestreamsHandler := livestream.NewHandler(log, lss)
	apiMux.HandleFunc("GET /livestreams", livestreamsHandler.List)
	apiMux.HandleFunc("GET /livestreams/{username}", livestreamsHandler.Get)
	apiMux.HandleFunc("PATCH /livestreams/{username}", authMiddleware(log, livestreamsHandler.Patch))
	apiMux.HandleFunc("POST /livestreams/{username}", authMiddleware(log, livestreamsHandler.Post))
	apiMux.HandleFunc("DELETE /livestreams/{username}", authMiddleware(log, livestreamsHandler.Delete))

	// {categoryIdentifier} is either int id or category link (e.g. "path-of-exile")
	categoriesHandler := category.NewHandler(log, cs)
	apiMux.HandleFunc("GET /categories", categoriesHandler.List)
	apiMux.HandleFunc("GET /categories/{categoryIdentifier}", categoriesHandler.Get)
	apiMux.HandleFunc("POST /categories", authMiddleware(log, categoriesHandler.Post))
	apiMux.HandleFunc("PATCH /categories/{categoryIdentifier}", authMiddleware(log, categoriesHandler.Patch))
	apiMux.HandleFunc("DELETE /categories/{categoryIdentifier}", authMiddleware(log, categoriesHandler.Delete))

	authHandler := auth.NewHandler(log, as)
	apiMux.HandleFunc("POST /auth/login", authHandler.Login)
	apiMux.HandleFunc("POST /auth/register", authHandler.Register)
	// apiMux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	// apiMux.HandleFunc("POST /auth/logout", authHandler.Logout)

	followHandler := follow.NewHandler(log, fs)
	apiMux.HandleFunc("GET /follow", followHandler.List)
	apiMux.HandleFunc("GET /follow/{username}", followHandler.Get)
	apiMux.HandleFunc("POST /follow/{username}", authMiddleware(log, followHandler.Post))
	apiMux.HandleFunc("DELETE /follow/{username}", authMiddleware(log, followHandler.Delete))

	userHandler := user.NewHandler(log, us)
	apiMux.HandleFunc("GET /users", authMiddleware(log, userHandler.List))
	apiMux.HandleFunc("GET /users/{username}", userHandler.Get)
	apiMux.HandleFunc("POST /users", userHandler.Post)
	apiMux.HandleFunc("PATCH /users/{username}", authMiddleware(log, userHandler.Patch))
	apiMux.HandleFunc("DELETE /users/{username}", authMiddleware(log, userHandler.Delete))

	channelHandler := channel.NewHandler(log, chs)
	apiMux.HandleFunc("GET /channels/{channel}", channelHandler.Get)
	// apiMux.HandleFunc("POST /channels/{channel}", authMiddleware(log, channelHandler.Post))
	apiMux.HandleFunc("PATCH /channels/{channel}", authMiddleware(log, channelHandler.Patch))
	// apiMux.HandleFunc("DELETE /channels/{channel}", authMiddleware(log, channelHandler.Delete))

	apiMux.HandleFunc("GET /health", health.Get)

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	fileserver := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileserver))
}
