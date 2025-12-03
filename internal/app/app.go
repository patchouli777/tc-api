package app

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/endpoint/auth"
	"main/internal/endpoint/category"
	"main/internal/endpoint/follow"
	"main/internal/endpoint/livestream"
	"main/internal/endpoint/presence"
	"main/internal/endpoint/user"
	"main/internal/lib/mw"
	"main/internal/lib/sl"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, cfg Config) *http.Server {
	rdb := redis.NewClient(cfg.Redis)

	pool, err := pgxpool.New(ctx, cfg.Postgres.ConnURL)
	if err != nil {
		cfg.Log.Error("unable to create connection pool", sl.Err(err))
		return nil
	}

	grpcClient, err := NewGRPClient(cfg.Log, cfg.Env, cfg.GRPC)
	if err != nil {
		cfg.Log.Error("unable to initialize grpc client", sl.Err(err))
		return nil
	}

	srvcs := InitServices(ctx,
		cfg.Log,
		cfg.InstanceID.String(),
		cfg.Update.LivestreamsTimer,
		grpcClient,
		rdb,
		pool)

	handler := CreateHandler(ctx, cfg, srvcs)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout}

	return server
}

func CreateHandler(ctx context.Context,
	cfg Config,
	srvcs Services) http.Handler {
	go srvcs.Presence.Run(ctx, cfg.InstanceID, cfg.Update.LivestreamsTimer)

	mux := http.NewServeMux()
	addRoutes(mux, cfg.Log,
		srvcs.Livestream,
		srvcs.Category,
		srvcs.Auth,
		srvcs.Follow,
		srvcs.User,
		srvcs.Presence)

	// go update.UpdateCategories(updateCtx, log, upd.CategoriesTimer, cs, lsr)

	return mw.PanicRecovery(mw.JSONResponse(mw.CORS(mw.Logging(mux))))
}

type Services struct {
	Auth       *auth.ServiceImpl
	Livestream *livestream.ServiceImpl
	Category   *category.RepositoryImpl
	Follow     *follow.ServiceImpl
	LsUpdater  *livestream.UpdaterImpl
	User       *user.ServiceImpl
	Presence   *presence.Server
}

func InitServices(
	ctx context.Context,
	log *slog.Logger,
	instanceID string,
	lsUpdateTimeout time.Duration,
	grpcClient auth.GRPCCLient,
	rdb *redis.Client,
	pool *pgxpool.Pool) Services {
	livestreamRepo := livestream.NewRepo(rdb, pool)
	livestreamUpdaterRepo := livestream.NewUpdaterRepo(rdb)
	// livestreamEventCh := make(chan livestream.EventLivestream)
	livestreamUpdater := livestream.NewUpdater(rdb,
		livestreamRepo,
		livestreamUpdaterRepo,
		// livestreamEventCh,
		instanceID)
	livestreamUpdater.Subscribe(ctx)
	livestreamsService := livestream.NewService(log, livestreamRepo)

	categoryRepo := category.NewRepo(rdb, pool)

	authDBAdapter := auth.NewAdapter(pool)
	authService := auth.NewService(log, grpcClient, authDBAdapter)

	followService := follow.NewService(log, pool)

	userRepo := user.NewRepository(pool)
	userService := user.NewService(log, userRepo)

	viewerStore := presence.NewViewerStore(rdb)
	wsServer := presence.NewServer(log, viewerStore)

	return Services{
		Auth:       authService,
		Livestream: livestreamsService,
		LsUpdater:  livestreamUpdater,
		Category:   categoryRepo,
		Follow:     followService,
		User:       userService,
		Presence:   wsServer}
}

func NewGRPClient(log *slog.Logger, env string, cfg GrpcClientConfig) (auth.GRPCCLient, error) {
	switch env {
	case envLocal:
		log.Info("skipped initialization of grpc client because env is local or dev")
		return &auth.GRPCClientMock{}, nil
	case envProd:
		return auth.NewGRPClient(log,
			cfg.Host,
			cfg.Port,
			cfg.Timeout,
			cfg.Retries)
	default:
		log.Error(`failed to initialize grpc client due
	to unknown value of "env" variable. Accepted values are: "prod", "dev", "local". Defaulting to mock. (big danger)`)
		return &auth.GRPCClientMock{}, nil
	}
}
