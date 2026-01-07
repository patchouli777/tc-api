package app

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/endpoint/auth"
	"main/internal/endpoint/category"
	"main/internal/endpoint/channel"
	"main/internal/endpoint/follow"
	"main/internal/endpoint/livestream"
	"main/internal/endpoint/user"
	"main/internal/lib/mw"
	"main/internal/lib/sl"
	"net/http"
	"os"

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

	grpcClient, err := NewGRPClient(log, cfg.Env, cfg.AuthServiceMock, cfg.GRPC)
	if err != nil {
		log.Error("unable to initialize grpc client", sl.Err(err))
		return nil
	}

	srvcs := InitApp(ctx,
		log,
		cfg.InstanceID.String(),
		cfg.Env,
		grpcClient,
		rdb,
		pool)
	srvcs.StreamServerAdapter.Update(ctx, cfg.Update.LivestreamsTimeout)
	srvcs.CategoryUpdater.Update(ctx, cfg.Update.CategoriesTimeout)

	handler := CreateHandler(ctx, log, cfg, srvcs)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout}

	return server
}

func CreateHandler(ctx context.Context, log *slog.Logger, cfg Config, srvcs App) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, log,
		srvcs.Category,
		srvcs.Livestream,
		srvcs.Channel,
		srvcs.Auth,
		srvcs.Follow,
		srvcs.User)

	panicRecovery := mw.PanicRecovery(log)
	logging := mw.Logging(ctx, log)
	return panicRecovery(mw.JSONResponse(mw.CORS(logging(mux))))
}

type App struct {
	Auth                *auth.ServiceImpl
	Livestream          *livestream.ServiceImpl
	StreamServerAdapter *livestream.StreamServerAdapter
	Category            *category.RepositoryImpl
	CategoryUpdater     *category.CategoryUpdater
	Follow              *follow.ServiceImpl
	User                *user.ServiceImpl
	Channel             *channel.ServiceImpl
}

func InitApp(ctx context.Context,
	log *slog.Logger,
	instanceID string,
	env string,
	grpcClient auth.GRPCCLient,
	rdb *redis.Client,
	pool *pgxpool.Pool) App {
	livestreamRepo := livestream.NewRepo(rdb, pool)
	streamServerAdapter := livestream.NewStreamServerAdapter(log, livestreamRepo, instanceID)
	livestreamsService := livestream.NewService(livestreamRepo)

	channelDBAdapter := channel.NewAdapter(pool)
	channelService := channel.NewService(log, channelDBAdapter)

	categoryRepo := category.NewRepo(rdb, pool)
	categoryUpdater := category.NewUpdater(log, livestreamRepo, categoryRepo)

	authDBAdapter := auth.NewAdapter(pool)
	authService := auth.NewService(grpcClient, authDBAdapter)

	followService := follow.NewService(pool)

	userRepo := user.NewRepository(pool)
	userService := user.NewService(userRepo)

	return App{
		Auth:                authService,
		Livestream:          livestreamsService,
		Channel:             channelService,
		Category:            categoryRepo,
		CategoryUpdater:     categoryUpdater,
		Follow:              followService,
		User:                userService,
		StreamServerAdapter: streamServerAdapter}
}

func NewGRPClient(log *slog.Logger, env string, isMock bool, cfg GrpcClientConfig) (auth.GRPCCLient, error) {
	if env == envProd {
		return auth.NewGRPClient(log,
			cfg.Host,
			cfg.Port,
			cfg.Timeout,
			cfg.Retries)
	}

	if !isMock {
		log.Info("Initializating real grpc client because AUTH_SERVICE_MOCK == false")
		return auth.NewGRPClient(log,
			cfg.Host,
			cfg.Port,
			cfg.Timeout,
			cfg.Retries)
	}

	log.Info("ENV is not prod and AUTH_SERVICE_MOCK is not false. Initializing mock grpc client.")
	return &auth.GRPCClientMock{}, nil
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
