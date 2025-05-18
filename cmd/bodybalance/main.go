package main

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/server"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres"
	"github.com/theartofdevel/logging"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// Информация о сборке
var (
	Version   = "unknown"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Инициализируем конфиг
	cfg := config.MustGetConfig()

	// Инициализируем лог
	logger := newLogger(cfg)

	// Добавляем логер в контекст
	ctx, cancel := context.WithCancel(logging.ContextWithLogger(context.Background(), logger))
	defer cancel()

	// Инициализируем хранилище PostgresSQL
	pgStorage, err := postgres.New(ctx, cfg, logger)
	if err != nil {
		log.Fatalln("Failed to initialize PostgresSQL storage", sl.Err(err))
	}
	defer func(pgStorage *postgres.Storage) {
		err = pgStorage.Close()
		if err != nil {
			logger.Error("can't close pgStorage", sl.Err(err))
		}
	}(pgStorage)

	logger.With(
		"Config params", cfg,
		"go_version", runtime.Version(),
		"build_time", BuildTime,
		"git_commit", GitCommit,
		"version", Version,
	).Info("starting server")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := server.Init(cfg, logger, pgStorage)

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
	logger := logging.NewLogger(
		logging.WithLevel(cfg.LogLevel),
		logging.WithIsJSON(true),
		logging.WithSetDefault(true),
	)

	logger = logger.With(
		slog.Group("program_info",
			slog.Int("num_goroutines", runtime.NumGoroutine()),
			slog.Int("pid", os.Getpid()),
		))

	return logger
}
