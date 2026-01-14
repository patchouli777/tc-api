package app

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ilyakaznacheev/cleanenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Config struct {
	Postgres           PostgresConfig
	HTTP               HttpServerConfig
	GRPC               GrpcClientConfig
	Redis              RedisConfig
	Logger             LoggerConfig
	Update             UpdateConfig
	Env                string `env:"ENV" env-default:"prod"`
	InstanceID         uuid.UUID
	AuthServiceMock    bool `env:"AUTH_SERVICE_MOCK" env-default:"false"`
	AuthMiddlewareMock bool `env:"AUTH_MIDDLEWARE_MOCK" env-default:"false"`
}

func GetConfig() Config {
	var cfg Config

	// root := util.GetProjectRoot()
	// err := cleanenv.ReadConfig(root+"\\.env", &cfg)
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		log.Fatalf("config big bad: %v", err)
	}
	cfg.InstanceID = uuid.New()

	return cfg
}

type HttpServerConfig struct {
	Host         string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port         string        `env:"HTTP_PORT" env-default:"8090"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"30s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"30s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"30s"`
}

type UpdateConfig struct {
	LivestreamsTimeout time.Duration `env:"UPDATE_LIVESTREAMS_TIMEOUT_SECONDS" env-default:"15s"`
	CategoriesTimeout  time.Duration `env:"UPDATE_CATEGORIES_TIMEOUT_SECONDS" env-default:"10s"`
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"cherry"`
	Name     string `env:"POSTGRES_DB" env-default:"baklava"`
	Password string `env:"POSTGRES_PASSWORD"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" env-default:"127.0.0.1"`
	Port     string `env:"REDIS_PORT" env-default:"6379"`
	Password string `env:"REDIS_PASSWORD"`
}

type GrpcClientConfig struct {
	Host    string        `env:"GRPC_CLIENT_HOST" env-default:"0.0.0.0"`
	Port    string        `env:"GRPC_CLIENT_PORT" env-default:"44044"`
	Timeout time.Duration `env:"GRPC_CLIENT_TIMEOUT_SECONDS" env-default:"5s"`
	Retries int           `env:"GRPC_CLIENT_RETRIES" env-default:"3"`
}

type LoggerConfig struct {
	Level   string `env:"LOG_LEVEL" env-default:"debug"`
	Handler string `env:"LOG_HANDLER" env-default:"text"`
}
