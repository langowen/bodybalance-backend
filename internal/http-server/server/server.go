package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1"
	mwLogger "github.com/langowen/bodybalance-backend/internal/http-server/middleware/logger"
	"github.com/langowen/bodybalance-backend/internal/http-server/server/handler"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres"
	"github.com/theartofdevel/logging"
	"net/http"
)

func Init(cfg *config.Config, logger *logging.Logger, pgStorage *postgres.Storage) *http.Server {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(logger)) //TODO подумать над заменой логера из пакета https://github.com/go-chi/httplog/blob/master/httplog.go
	router.Use(middleware.Recoverer)

	v1.New(router, logger, pgStorage, cfg)

	router.Get("/video/{filename}", handler.ServeVideoFile(cfg, logger))

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	return srv
}
