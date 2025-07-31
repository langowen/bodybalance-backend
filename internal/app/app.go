package app

import (
	"context"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage/postgres"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage/redis"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/logpretty"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/service/api"
	"github.com/theartofdevel/logging"
	"log"
	"log/slog"
	"os"
	"runtime"
)

type App struct {
	Cfg     *config.Config
	Logger  *logging.Logger
	Storage *postgres.Storage
	Redis   *redis.Storage
	Service *api.ServiceApi
}

func NewApp(cfg *config.Config) *App {
	return &App{
		Cfg: cfg,
	}
}

func (a *App) GetLogger() {
	logger := newLogger(a.Cfg)

	a.Logger = logger
}

func (a *App) GetStorage(ctx context.Context) {
	pgPool, err := postgres.InitStorage(ctx, a.Cfg)
	if err != nil {
		log.Fatalln("Failed to initialize PostgresSQL storage", sl.Err(err))
	}

	pgStorage := postgres.NewStorage(pgPool, a.Cfg)

	a.Storage = pgStorage
}

func (a *App) GetRedis(ctx context.Context) {
	if a.Cfg.Redis.Enable == true {
		rdb, err := redis.InitRedis(ctx, a.Cfg)
		if err != nil {
			log.Fatalln("Failed to initialize redis storage", sl.Err(err))
		}
		redisStorage := redis.NewStorage(rdb, a.Cfg)

		a.Redis = redisStorage
	}
}

func (a *App) GetService() {
	service := api.NewServiceApi(a.Cfg, a.Storage.Api, a.Redis)

	a.Service = service
}

func newLogger(cfg *config.Config) *logging.Logger {
	var logger *logging.Logger

	switch cfg.Env {
	case "local":
		logger = setupPrettySlog()
	case "prod":
		logger = logging.NewLogger(
			logging.WithLevel(cfg.LogLevel),
			logging.WithIsJSON(true),
			logging.WithSetDefault(true),
			logging.WithLogFilePath(cfg.PatchLog),
		)
	case "dev", "test":
		logger = logging.NewLogger(
			logging.WithLevel(cfg.LogLevel),
			logging.WithIsJSON(true),
			logging.WithSetDefault(true),
			logging.WithLogFilePath(cfg.PatchLog),
		)
	default:
		logger = logging.NewLogger(
			logging.WithLevel(cfg.LogLevel),
			logging.WithIsJSON(true),
			logging.WithSetDefault(true),
			logging.WithLogFilePath(cfg.PatchLog),
		)
	}

	logger = logger.With(
		slog.Group("program_info",
			slog.Int("num_goroutines", runtime.NumGoroutine()),
			slog.Int("pid", os.Getpid()),
		))

	return logger
}

func setupPrettySlog() *logging.Logger {
	opts := logpretty.PrettyHandlerOptions{
		SlogOpts: &logging.HandlerOptions{
			Level: logging.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return logging.New(handler)
}
