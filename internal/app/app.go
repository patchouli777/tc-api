package app

import (
	"context"
	"fmt"
	"log/slog"
	appAuth "main/internal/app/auth"
	"main/internal/auth"
	"main/internal/category"
	categoryStorage "main/internal/category/storage"
	"main/internal/channel"
	authExternal "main/internal/external/auth"
	"main/internal/external/streamserver"
	"main/internal/follow"
	"main/internal/lib/mw"
	"main/internal/lib/sl"
	"main/internal/livestream"
	livestreamStorage "main/internal/livestream/storage"
	"main/internal/user"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func New(ctx context.Context, log *slog.Logger, cfg Config) *http.Server {
	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		// https://github.com/redis/go-redis/issues/3536
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
		Addr:     redisAddr,
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

	app := NewApp(ctx,
		log,
		rdb,
		pool,
		authClient,
		cfg.Env,
		cfg.InstanceID.String(),
		cfg.StreamServer,
		cfg.Asynq)

	InitApp(ctx, log, rdb, cfg.InstanceID.String(), app, cfg.Update, cfg.StreamServer)

	authMw := NewAuthMiddleware(log, cfg.AuthMiddlewareMock)
	handler := CreateHandler(ctx, log, authMw, app)
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout}

	return server
}

// spins up workers, subscribes to stream server event and stuff
func InitApp(ctx context.Context, log *slog.Logger, r *redis.Client, instanceID string, app *App, updCfg UpdateConfig, ssCfg StreamServerConfig) {
	asyncqMux := asynq.NewServeMux()
	updSched := livestream.NewUpdateScheduler(log, r, app.StreamServerAdapter, app.LivestreamRepo, app.TaskScheduler, instanceID)
	asyncqMux.HandleFunc(livestream.TypeLivestreamUpdate, updSched.HandleUpdateTask)

	go func() {
		if err := app.TaskQServer.Run(asyncqMux); err != nil {
			log.Error("asynq server down", sl.Err(err))
		}
	}()

	go func() {
		if err := app.TaskScheduler.Run(); err != nil {
			log.Error("scheduler down", sl.Err(err))
		}
	}()

	app.CategoryUpdater.Update(ctx, updCfg.CategoriesTimeout)
	go updSched.Run(ctx, updCfg.LivestreamsTimeout)

	// TODO: events
	// cl := baseclient.NewClient()
	// ssURL := fmt.Sprintf("http://%s:%s%s/", ssCfg.Host, ssCfg.Port, ssCfg.Endpoint)
	// req, err := cl.Post(ssURL+"/subscribe",
	// 	streamserver.SubscribeRequest{CallbackURL: "/webhooks/livestreams"})
	// if err != nil {
	// 	log.Error("big error 1", sl.Err(err))
	// }

	// resp, err := cl.Client.Do(req)
	// if err != nil {
	// 	log.Error("big error 2", sl.Err(err))
	// }

	// var subr streamserver.SubscribeResponse
	// err = json.NewDecoder(resp.Body).Decode(&subr)
	// if err != nil {
	// 	log.Error("big error 3", sl.Err(err))
	// }
	// fmt.Println(subr)
}

func CreateHandler(ctx context.Context, log *slog.Logger, authMw authMw, app *App) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, log,
		authMw,
		app.CategoryRepo,
		app.LivestreamRepo,
		app.ChannelRepo,
		app.AuthService,
		app.FollowRepo,
		app.UserRepo)

	panicRecovery := mw.PanicRecovery(log)
	logging := mw.Logging(log)
	return mw.RequestID(panicRecovery(mw.JSONResponse(mw.CORS(logging(mux)))))
}

type App struct {
	AuthService         *auth.ServiceImpl
	LivestreamRepo      *livestreamStorage.RepositoryImpl
	CategoryRepo        *categoryStorage.RepositoryImpl
	StreamServerAdapter *streamserver.Adapter
	CategoryUpdater     *category.CategoryUpdater
	FollowRepo          *follow.RepositoryImpl
	UserRepo            *user.RepositoryImpl
	ChannelRepo         *channel.RepositoryImpl
	TaskQServer         *asynq.Server
	TaskScheduler       *asynq.Scheduler
}

// initialize pgx, redis and auth client outside because of tests
func NewApp(ctx context.Context,
	log *slog.Logger,
	rdb *redis.Client,
	pool *pgxpool.Pool,
	// TODO: why client here?
	client auth.Client,
	env string,
	instanceID string,
	ssCfg StreamServerConfig,
	asynqCfg AsynqConfig) *App {
	livestreamRepo := livestreamStorage.NewRepo(rdb, pool)

	channelDBAdapter := channel.NewAdapter(pool)
	channelRepo := channel.NewRepository(log, channelDBAdapter)

	categoryRepo := categoryStorage.NewRepo(rdb, pool)
	categoryUpdater := category.NewUpdater(log, livestreamRepo, categoryRepo)

	authDBAdapter := auth.NewAdapter(pool)
	authService := auth.NewService(client, authDBAdapter)

	followRepo := follow.NewRepository(pool)

	userRepo := user.NewRepository(pool)

	ssURL := fmt.Sprintf("http://%s:%s%s/", ssCfg.Host, ssCfg.Port, ssCfg.Endpoint)
	streamServerAdapter := streamserver.NewAdapter(log, ssURL)

	asyncRedis := fmt.Sprintf("%s:%s", asynqCfg.RedisHost, asynqCfg.RedisPort)
	taskqserv := asynq.NewServer(asynq.RedisClientOpt{Addr: asyncRedis},
		asynq.Config{Concurrency: asynqCfg.MaxConcurrentTasks})
	sched := asynq.NewScheduler(asynq.RedisClientOpt{Addr: asyncRedis}, nil)

	return &App{
		AuthService:         authService,
		LivestreamRepo:      livestreamRepo,
		ChannelRepo:         channelRepo,
		CategoryRepo:        categoryRepo,
		CategoryUpdater:     categoryUpdater,
		FollowRepo:          followRepo,
		UserRepo:            userRepo,
		StreamServerAdapter: streamServerAdapter,
		TaskQServer:         taskqserv,
		TaskScheduler:       sched}
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

	log.Info("ENV is not prod and AUTH_SERVICE_MOCK == true. Initializing mock grpc auth client.")
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
