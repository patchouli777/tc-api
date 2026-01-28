package app

import (
	"log/slog"
	"net/http"
	"twitchy-api/internal/auth"
	authStorage "twitchy-api/internal/auth/storage"
	"twitchy-api/internal/category"
	categoryStorage "twitchy-api/internal/category/storage"
	"twitchy-api/internal/channel"
	channelStorage "twitchy-api/internal/channel/storage"
	"twitchy-api/internal/follow"
	followStorage "twitchy-api/internal/follow/storage"
	"twitchy-api/internal/health"
	"twitchy-api/internal/livestream"
	livestreamStorage "twitchy-api/internal/livestream/storage"
	"twitchy-api/internal/user"
	userStorage "twitchy-api/internal/user/storage"

	_ "twitchy-api/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func addRoutes(mux *http.ServeMux,
	log *slog.Logger,
	authMw mware,
	cr *categoryStorage.RepositoryImpl,
	lsr *livestreamStorage.RepositoryImpl,
	chr *channelStorage.RepositoryImpl,
	as *authStorage.ServiceImpl,
	fr *followStorage.RepositoryImpl,
	ur *userStorage.RepositoryImpl) {
	apiMux := http.NewServeMux()

	livestreamsHandler := livestream.NewHandler(log, lsr)
	apiMux.HandleFunc("GET /livestreams", livestreamsHandler.List)
	apiMux.HandleFunc("GET /livestreams/{id}", livestreamsHandler.Get)
	apiMux.HandleFunc("GET /users/{username}/livestream", livestreamsHandler.GetByUsername)

	// {identifier} is either int id or category link (e.g. "path-of-exile")
	categoriesHandler := category.NewHandler(log, cr)
	apiMux.HandleFunc("GET /categories", categoriesHandler.List)
	apiMux.HandleFunc("GET /categories/{identifier}", categoriesHandler.Get)
	apiMux.HandleFunc("POST /categories", authMw(categoriesHandler.Post))
	apiMux.HandleFunc("PATCH /categories/{identifier}", authMw(categoriesHandler.Patch))
	// NOTE: currently deleting a category which id is in tc_livestream table is not possible
	// due to non-nullable fk constraint (which is fine but might want to change later)
	apiMux.HandleFunc("DELETE /categories/{identifier}", authMw(categoriesHandler.Delete))

	authHandler := auth.NewHandler(log, as)
	apiMux.HandleFunc("POST /auth/signin", authHandler.SignIn)
	apiMux.HandleFunc("POST /auth/signup", authHandler.SignUp)

	followHandler := follow.NewHandler(log, fr)
	apiMux.HandleFunc("GET /follow", followHandler.List)
	apiMux.HandleFunc("GET /follow/{username}", followHandler.Get)
	apiMux.HandleFunc("POST /follow/{username}", authMw(followHandler.Post))
	apiMux.HandleFunc("DELETE /follow/{username}", authMw(followHandler.Delete))

	userHandler := user.NewHandler(log, ur)
	apiMux.HandleFunc("GET /users", authMw(userHandler.List))
	apiMux.HandleFunc("GET /users/{id}", authMw(userHandler.Get))
	apiMux.HandleFunc("POST /users", userHandler.Post)
	apiMux.HandleFunc("PATCH /users/{id}", authMw(userHandler.Patch))
	apiMux.HandleFunc("DELETE /users/{id}", authMw(userHandler.Delete))

	channelHandler := channel.NewHandler(log, chr)
	apiMux.HandleFunc("GET /channels/{channel}", channelHandler.Get)
	apiMux.HandleFunc("PATCH /channels/{channel}", authMw(channelHandler.Patch))

	apiMux.HandleFunc("GET /health", health.Get)

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// mux.HandleFunc("POST /webhooks/livestreams", livestreamsWebhooksHandler)

	fileserver := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileserver))
}

// TODO: events
// func livestreamsWebhooksHandler(w http.ResponseWriter, r *http.Request) {
// 	var p streamserver.StreamStartedPayload
// 	err := json.NewDecoder(r.Body).Decode(&p)
// 	if err != nil {
// 		fmt.Printf("bad: %+v\n", err)
// 	}

// 	fmt.Println("good:", p)
// }
