//go:generate go run github.com/swaggo/swag/cmd/swag@latest init --output ../../docs --parseDepth 3 --parseDependency --parseInternal
package main

import (
	"context"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/port/http-server"
	"github.com/theartofdevel/logging"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

	// Инициализируем app
	apps := app.NewApp(cfg)

	// Инициализируем лог
	apps.GetLogger()
	logger := apps.Logger

	logger.With(
		"Config params", cfg,
		"go_version", runtime.Version(),
		"git_commit", GitCommit,
		"version", Version,
	).Info("starting application")

	// Добавляем логер в контекст
	ctx, cancel := context.WithCancel(logging.ContextWithLogger(context.Background(), logger))

	// Обработчик сигналов для graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Инициализируем хранилище PostgresSQL
	apps.GetStorage(ctx)

	// Инициализируем хранилище Redis
	apps.GetRedis(ctx)

	// Инициализируем сервис API
	apps.GetService()

	// Инициализируем HTTP сервер
	srv := http_server.NewServer(apps)
	serverDone := srv.StartServer(ctx)
	logger.Info("server started")

	// Ждем сигнала для graceful shutdown
	<-done
	logger.Info("Gracefully shutting down")

	// Отменяем контекст
	cancel()
	logger.Info("stopping server")

	// Ждем завершения работы сервера
	<-serverDone

	logger.Info("server stopped")
}
