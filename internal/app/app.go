package app

import (
	"context"
	"fmt"
	"log/slog"
	appAuth "main/internal/auth"
	"main/internal/endpoint/auth"
	"main/internal/endpoint/category"
	"main/internal/endpoint/channel"
	"main/internal/endpoint/follow"
	"main/internal/endpoint/livestream"
	"main/internal/endpoint/user"
	authExternal "main/internal/external/auth"
	"main/internal/external/streamserver"
	"main/internal/lib/mw"
	"main/internal/lib/sl"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func New(ctx context.Context, log *slog.Logger, cfg Config) *http.Server {
	rdb := redis.NewClient(&redis.Options{
		// https://github.com/redis/go-redis/issues/3536
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	postgresConnURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Name)

	pool, err := pgxpool.New(ctx, postgresConnURL)
	if err != nil {
		log.Error("unable to create connection pool", sl.Err(err))
		return nil
	}

	authClient, err := NewAuthClient(log, cfg.Env, cfg.AuthServiceMock, cfg.GRPC)
	if err != nil {
		log.Error("unable to initialize auth client", sl.Err(err))
		return nil
	}

	app := InitApp(ctx,
		log,
		cfg.InstanceID.String(),
		cfg.Env,
		authClient,
		rdb,
		pool)

	authMw := NewAuthMiddleware(log, cfg.AuthMiddlewareMock)
	handler := CreateHandler(ctx, log, authMw, app)

	taskqserv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: fmt.Sprintf("%s:%s", "0.0.0.0", "6379")},
		asynq.Config{Concurrency: 10},
	)

	sched := asynq.NewScheduler(asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%s", "0.0.0.0", "6379")}, nil)

	asyncqMux := asynq.NewServeMux()
	updSched := livestream.NewUpdateScheduler(log, app.StreamServerAdapter, app.Livestream, sched)
	asyncqMux.HandleFunc(livestream.TypeLivestreamUpdate, updSched.HandleUpdateTask)

	go func() {
		if err := taskqserv.Run(asyncqMux); err != nil {
			log.Error("asynq server down")
		}
	}()

	go func() {
		if err := sched.Run(); err != nil {
			log.Error("scheduler down", sl.Err(err))
			return
		}
	}()

	app.CategoryUpdater.Update(ctx, cfg.Update.CategoriesTimeout)
	go updSched.Update(ctx, cfg.Update.LivestreamsTimeout)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout}

	return server
}

func CreateHandler(ctx context.Context, log *slog.Logger, authMw authMw, app *App) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, log,
		authMw,
		app.Category,
		app.Livestream,
		app.Channel,
		app.Auth,
		app.Follow,
		app.User)

	panicRecovery := mw.PanicRecovery(log)
	logging := mw.Logging(log)
	return mw.RequestID(panicRecovery(mw.JSONResponse(mw.CORS(logging(mux)))))
}

type App struct {
	Auth                *auth.ServiceImpl
	Livestream          *livestream.RepositoryImpl
	Category            *category.RepositoryImpl
	StreamServerAdapter *streamserver.Adapter
	CategoryUpdater     *category.CategoryUpdater
	Follow              *follow.RepositoryImpl
	User                *user.RepositoryImpl
	Channel             *channel.RepositoryImpl
}

func InitApp(ctx context.Context,
	log *slog.Logger,
	instanceID string,
	env string,
	client auth.Client,
	rdb *redis.Client,
	pool *pgxpool.Pool) *App {
	livestreamRepo := livestream.NewRepo(rdb, pool)

	channelDBAdapter := channel.NewAdapter(pool)
	channelRepo := channel.NewRepository(log, channelDBAdapter)

	categoryRepo := category.NewRepo(rdb, pool)
	categoryUpdater := category.NewUpdater(log, livestreamRepo, categoryRepo)

	authDBAdapter := auth.NewAdapter(pool)
	authService := auth.NewService(client, authDBAdapter)

	followRepo := follow.NewRepository(pool)

	userRepo := user.NewRepository(pool)

	streamServerAdapter := streamserver.NewAdapter(log)

	return &App{
		Auth:                authService,
		Livestream:          livestreamRepo,
		Channel:             channelRepo,
		Category:            categoryRepo,
		CategoryUpdater:     categoryUpdater,
		Follow:              followRepo,
		User:                userRepo,
		StreamServerAdapter: streamServerAdapter}
}

type authMw = func(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc

func NewAuthMiddleware(log *slog.Logger, isMock bool) authMw {
	if isMock {
		log.Info("Initializating mock auth middleware because AUTH_MIDDLEWARE_MOCK == true")
		return appAuth.AuthMiddlewareMock
	} else {
		return appAuth.AuthMiddleware
	}
}

func NewAuthClient(log *slog.Logger, env string, isMock bool, cfg GrpcClientConfig) (auth.Client, error) {
	if env == envProd {
		return authExternal.NewClient(log,
			cfg.Host,
			cfg.Port,
			cfg.Timeout,
			cfg.Retries)
	}

	if !isMock {
		log.Info("Initializating real grpc auth client because AUTH_SERVICE_MOCK == false")
		return authExternal.NewClient(log,
			cfg.Host,
			cfg.Port,
			cfg.Timeout,
			cfg.Retries)
	}

	log.Info("ENV is not prod and AUTH_SERVICE_MOCK is not false. Initializing mock grpc auth client.")
	return &authExternal.ClientMock{}, nil
}

func NewLogger(cfg LoggerConfig) *slog.Logger {
	var lev slog.Leveler
	switch cfg.Level {
	case "debug":
		lev = slog.LevelDebug
	case "info":
		lev = slog.LevelInfo
	case "error":
		lev = slog.LevelError
	default:
		lev = slog.LevelInfo
	}

	var h slog.Handler
	switch cfg.Handler {
	case "text":
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lev})
	case "json":
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lev})
	default:
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lev})
	}

	return slog.New(h)
}
