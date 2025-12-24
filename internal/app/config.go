package app

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Config struct {
	Postgres          PostgresConfig
	HTTP              HttpServerConfig
	GRPC              GrpcClientConfig
	Update            UpdateConfig
	Redis             *redis.Options
	Log               *slog.Logger
	Env               string
	InstanceID        uuid.UUID
	AuthServiceMock   bool
	StreamServiceMock bool
}

func GetConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbUserPassword := os.Getenv("POSTGRES_PASSWORD")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("POSTGRES_NAME")
	dbHost := os.Getenv("POSTGRES_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	apiPort := os.Getenv("API_PORT")
	env := os.Getenv("ENV")

	authServiceMockEnv := os.Getenv("AUTH_SERVICE_MOCK")
	authServiceMock, err := strconv.ParseBool(authServiceMockEnv)
	if err != nil {
		authServiceMock = false
	}

	streamServiceMockEnv := os.Getenv("STREAM_SERVICE_MOCK")
	streamServiceMock, err := strconv.ParseBool(streamServiceMockEnv)
	if err != nil {
		streamServiceMock = false
	}

	connURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbUserPassword, dbHost, dbPort, dbName)

	logger := setupLogger(env)
	slog.SetDefault(logger)

	return Config{
		Postgres: PostgresConfig{ConnURL: connURL},
		Redis: &redis.Options{
			// https://github.com/redis/go-redis/issues/3536
			MaintNotificationsConfig: &maintnotifications.Config{
				Mode: maintnotifications.ModeDisabled,
			},
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password: redisPassword,
			DB:       0,
		},
		Log: logger,
		Update: UpdateConfig{
			CategoriesTimer:  time.Second * 30,
			LivestreamsTimer: time.Second * 50,
		},
		HTTP: HttpServerConfig{
			Port:         apiPort,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 25 * time.Second,
			IdleTimeout:  45 * time.Second,
		},
		GRPC: GrpcClientConfig{
			Host:    "0.0.0.0",
			Port:    "44044",
			Timeout: time.Second * 5,
			Retries: 5,
		},
		Env:               env,
		InstanceID:        uuid.New(),
		AuthServiceMock:   authServiceMock,
		StreamServiceMock: streamServiceMock,
	}
}

type HttpServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Port         string
	Host         string
}

type UpdateConfig struct {
	CategoriesTimer  time.Duration
	LivestreamsTimer time.Duration
}

type PostgresConfig struct {
	ConnURL string
}

type GrpcClientConfig struct {
	Host    string
	Port    string
	Timeout time.Duration
	Retries int
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
}
