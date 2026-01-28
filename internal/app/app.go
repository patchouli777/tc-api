package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	appAuth "twitchy-api/internal/app/auth"
	authStorage "twitchy-api/internal/auth/storage"
	categoryService "twitchy-api/internal/category/service"
	categoryStorage "twitchy-api/internal/category/storage"
	channelStorage "twitchy-api/internal/channel/storage"
	authExternal "twitchy-api/internal/external/auth"
	"twitchy-api/internal/external/streamserver"
	"twitchy-api/internal/external/taskqueue"
	followStorage "twitchy-api/internal/follow/storage"
	"twitchy-api/internal/lib/mw"
	"twitchy-api/internal/lib/sl"
	livestreamService "twitchy-api/internal/livestream/service"
	livestreamStorage "twitchy-api/internal/livestream/storage"
	userStorage "twitchy-api/internal/user/storage"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"golang.org/x/sync/errgroup"
)

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := GetConfig()
	log := NewLogger(cfg.Logger)
	slog.SetDefault(log)

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
		return err
	}

	app, err := NewApp(log, rdb, pool, cfg)
	if err != nil {
		log.Error("unable to create app", sl.Err(err))
		return err
	}

	// spins up task queue and updaters
	// TODO: shutdown everyhthing if any fails?
	eg, ctx := errgroup.WithContext(ctx)
	app.Init(ctx, cfg.Update, eg)

	authMw := NewAuthMiddleware(log, cfg.AuthMiddlewareMock)
	handler := app.CreateHandler(authMw)
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server down", sl.Err(err))
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

	return err
}

type App struct {
	log                 *slog.Logger
	AuthService         *authStorage.ServiceImpl
	StreamServerAdapter *streamserver.Adapter
	LivestreamRepo      *livestreamStorage.RepositoryImpl
	LivestreamUpdater   *livestreamService.Updater
	CategoryRepo        *categoryStorage.RepositoryImpl
	CategoryUpdater     *categoryService.CategoryUpdater
	FollowRepo          *followStorage.RepositoryImpl
	UserRepo            *userStorage.RepositoryImpl
	ChannelRepo         *channelStorage.RepositoryImpl
	TaskQServer         *asynq.Server
	TaskScheduler       *taskqueue.Scheduler
}

// initialize pgx and redis outside because of tests
func NewApp(log *slog.Logger,
	rdb *redis.Client,
	pool *pgxpool.Pool,
	cfg Config) (*App, error) {
	livestreamRepo := livestreamStorage.NewRepo(rdb, pool)

	channelRepo := channelStorage.NewRepository(pool)

	categoryRepo := categoryStorage.NewRepo(rdb, pool)
	categoryUpdater := categoryService.NewUpdater(log, livestreamRepo, categoryRepo)

	authClient, err := NewAuthClient(log, cfg.Env, cfg.AuthServiceMock, cfg.GRPC)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize auth client: %v", err)
	}
	authService := authStorage.NewService(authClient, pool)

	followRepo := followStorage.NewRepository(pool)

	userRepo := userStorage.NewRepository(pool)

	ssURL := fmt.Sprintf("http://%s:%s%s/",
		cfg.StreamServer.Host, cfg.StreamServer.Port, cfg.StreamServer.Endpoint)
	streamServerAdapter := streamserver.NewAdapter(ssURL)

	asyncRedis := fmt.Sprintf("%s:%s", cfg.Asynq.RedisHost, cfg.Asynq.RedisPort)
	taskqserv := asynq.NewServer(asynq.RedisClientOpt{Addr: asyncRedis},
		asynq.Config{Concurrency: cfg.Asynq.MaxConcurrentTasks})
	sc := asynq.NewScheduler(asynq.RedisClientOpt{Addr: asyncRedis}, nil)
	sched := taskqueue.NewScheduler(sc)

	livestreamUpdater := livestreamService.NewUpdater(log,
		rdb,
		streamServerAdapter,
		livestreamRepo,
		sched,
		cfg.InstanceID.String())

	return &App{
		log:                 log,
		AuthService:         authService,
		LivestreamRepo:      livestreamRepo,
		LivestreamUpdater:   livestreamUpdater,
		ChannelRepo:         channelRepo,
		CategoryRepo:        categoryRepo,
		CategoryUpdater:     categoryUpdater,
		FollowRepo:          followRepo,
		UserRepo:            userRepo,
		StreamServerAdapter: streamServerAdapter,
		TaskQServer:         taskqserv,
		TaskScheduler:       sched}, nil
}

func (a *App) Init(ctx context.Context, cfg UpdateConfig, eg *errgroup.Group) {
	asyncqMux := asynq.NewServeMux()
	asyncqMux.HandleFunc(livestreamService.TaskUpdate,
		taskqueue.TaskHandler(a.LivestreamUpdater.HandleUpdateTask))

	eg.Go(func() error {
		err := a.TaskQServer.Run(asyncqMux)
		if err != nil {
			a.log.Error("asynq server down", sl.Err(err))
		}

		return err
	})

	eg.Go(func() error {
		err := a.TaskScheduler.Run()
		if err != nil {
			a.log.Error("scheduler down", sl.Err(err))
		}

		return err
	})

	eg.Go(func() error {
		return a.CategoryUpdater.Run(ctx, cfg.CategoriesTimeout)
	})

	eg.Go(func() error {
		return a.LivestreamUpdater.Run(ctx, cfg.LivestreamsTimeout)
	})
}

func (a *App) CreateHandler(authMw mware) http.Handler {
	mux := http.NewServeMux()

	addRoutes(mux, a.log,
		authMw,
		a.CategoryRepo,
		a.LivestreamRepo,
		a.ChannelRepo,
		a.AuthService,
		a.FollowRepo,
		a.UserRepo)

	panicRecovery := mw.PanicRecovery(a.log)
	logging := mw.Logging(a.log)
	return mw.RequestID(panicRecovery(mw.JSONResponse(mw.CORS(logging(mux)))))
}

type mware = func(next http.HandlerFunc) http.HandlerFunc

func NewAuthMiddleware(log *slog.Logger, isMock bool) mware {
	if isMock {
		log.Info("Initializating mock auth middleware because AUTH_MIDDLEWARE_MOCK == true")
		return appAuth.AuthMiddlewareMock(log)
	} else {
		return appAuth.AuthMiddleware(log)
	}
}

func NewAuthClient(log *slog.Logger, env string, isMock bool, cfg GrpcClientConfig) (authStorage.Client, error) {
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
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: lev,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					t := a.Value.Time()
					a.Value = slog.StringValue(t.Format(time.DateTime))
				}
				return a
			}})
	case "json":
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lev})
	default:
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lev})
	}

	return slog.New(h)
}
