package app

import (
	"log/slog"
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

func addRoutes(mux *http.ServeMux,
	log *slog.Logger,
	authMw authMw,
	cr *category.RepositoryImpl,
	lsr *livestream.RepositoryImpl,
	chr *channel.RepositoryImpl,
	as *auth.ServiceImpl,
	fr *follow.RepositoryImpl,
	ur *user.RepositoryImpl) {
	apiMux := http.NewServeMux()

	livestreamsHandler := livestream.NewHandler(log, lsr)
	apiMux.HandleFunc("GET /livestreams", livestreamsHandler.List)
	apiMux.HandleFunc("GET /livestreams/{username}", livestreamsHandler.Get)
	apiMux.HandleFunc("PATCH /livestreams/{username}", authMw(log, livestreamsHandler.Patch))

	// {identifier} is either int id or category link (e.g. "path-of-exile")
	categoriesHandler := category.NewHandler(log, cr)
	apiMux.HandleFunc("GET /categories", categoriesHandler.List)
	apiMux.HandleFunc("GET /categories/{identifier}", categoriesHandler.Get)
	apiMux.HandleFunc("POST /categories", authMw(log, categoriesHandler.Post))
	apiMux.HandleFunc("PATCH /categories/{identifier}", authMw(log, categoriesHandler.Patch))
	// NOTE: currently deleting a category which id is in tc_livestream table is not possible
	// due to non-nullable fk constraint (which is fine but might want to change later)
	apiMux.HandleFunc("DELETE /categories/{identifier}", authMw(log, categoriesHandler.Delete))

	authHandler := auth.NewHandler(log, as)
	apiMux.HandleFunc("POST /auth/signin", authHandler.SignIn)
	apiMux.HandleFunc("POST /auth/signup", authHandler.SignUp)

	followHandler := follow.NewHandler(log, fr)
	apiMux.HandleFunc("GET /follow", followHandler.List)
	apiMux.HandleFunc("GET /follow/{username}", followHandler.Get)
	apiMux.HandleFunc("POST /follow/{username}", authMw(log, followHandler.Post))
	apiMux.HandleFunc("DELETE /follow/{username}", authMw(log, followHandler.Delete))

	userHandler := user.NewHandler(log, ur)
	apiMux.HandleFunc("GET /users", authMw(log, userHandler.List))
	apiMux.HandleFunc("GET /users/{id}", authMw(log, userHandler.Get))
	apiMux.HandleFunc("POST /users", userHandler.Post)
	apiMux.HandleFunc("PATCH /users/{id}", authMw(log, userHandler.Patch))
	apiMux.HandleFunc("DELETE /users/{id}", authMw(log, userHandler.Delete))

	channelHandler := channel.NewHandler(log, chr)
	apiMux.HandleFunc("GET /channels/{channel}", channelHandler.Get)
	apiMux.HandleFunc("PATCH /channels/{channel}", authMw(log, channelHandler.Patch))

	apiMux.HandleFunc("GET /health", health.Get)

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	fileserver := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileserver))
}
