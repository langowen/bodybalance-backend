//go:generate go run github.com/swaggo/swag/cmd/swag@latest init --output ../../docs --parseDepth 3 --parseDependency --parseInternal
package main

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/server"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/logpretty"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres"
	"github.com/langowen/bodybalance-backend/internal/storage/redis"
	"github.com/theartofdevel/logging"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	_ "github.com/langowen/bodybalance-backend/docs"
	_ "github.com/langowen/bodybalance-backend/internal/http-server/admin"
	_ "github.com/langowen/bodybalance-backend/internal/http-server/api/v1"
)

var (
	Version   = "unknown"
	GitCommit = "unknown"
)

// BodyBalance API
// @title BodyBalance API
// @version 1.0
// @description API для управления видео-контентом BodyBalance.
// @host body.7375.org
// @BasePath /v1
// @schemes https
//
// @contact.name Sergei
// @contact.email info@7375.org
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// BodyBalance Admin API
// @title BodyBalance Admin API
// @version 1.0
// @description API для управления административной частью BodyBalance.
// @host body.7375.org
// @BasePath /admin
// @schemes https
//
// @contact.name Sergei
// @contact.email info@7375.org
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @securityDefinitions.apikey AdminAuth
// @in cookie
// @name token

func main() {
	// Инициализируем конфиг
	cfg := config.MustGetConfig()

	// Инициализируем лог
	logger := newLogger(cfg)

	if os.Getenv("GENERATE_SWAGGER") == "true" {
		return
	}

	// Добавляем логер в контекст
	ctx, cancel := context.WithCancel(logging.ContextWithLogger(context.Background(), logger))
	defer cancel()

	// Инициализируем хранилище PostgresSQL
	pgStorage, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Fatalln("Failed to initialize PostgresSQL storage", sl.Err(err))
	}
	defer func(pgStorage *postgres.Storage) {
		err = pgStorage.Close()
		if err != nil {
			logger.Error("can't close pgStorage", sl.Err(err))
		}
	}(pgStorage)

	redisStorage, err := redis.New(cfg)
	if err != nil {
		log.Fatalln("Failed to initialize RedisSQL storage", sl.Err(err))
	}

	logger.With(
		"Config params", cfg,
		"go_version", runtime.Version(),
		"git_commit", GitCommit,
		"version", Version,
	).Info("starting server")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	apps := app.New(cfg, logger, pgStorage, redisStorage)

	srv := server.Init(apps)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("failed to start server")
		}
	}()

	logger.Info("server started")

	<-done
	logger.Info("stopping server")

	ctxServer, cancelServer := context.WithTimeout(ctx, cfg.HTTPServer.Timeout)
	defer cancelServer()

	if err := srv.Shutdown(ctxServer); err != nil {
		logger.Error("failed to stop server", sl.Err(err))
	}

	logger.Info("server stopped")
}

func newLogger(cfg *config.Config) *logging.Logger {
	var logger *logging.Logger

	switch cfg.Env {
	case "local":
		logger = setupPrettySlog()
	case "prod", "dev", "test":
		logger = logging.NewLogger(
			logging.WithLevel(cfg.LogLevel),
			logging.WithIsJSON(true),
			logging.WithSetDefault(true),
			logging.WithLogFilePath("logs/body_balance.log"),
		)
	default:
		logger = logging.NewLogger(
			logging.WithLevel(cfg.LogLevel),
			logging.WithIsJSON(true),
			logging.WithSetDefault(true),
			logging.WithLogFilePath(""),
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
