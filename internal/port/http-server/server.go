package http_server

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/api/v1"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/handler/img"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/handler/video"
	mwLogger "github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/logger"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	"github.com/langowen/bodybalance-backend/internal/service/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/theartofdevel/logging"
	"net/http"
	"time"
)

type Server struct {
	app     *app.App
	service *api.ServiceApi
	logger  *logging.Logger
	cfg     *config.Config
}

func NewServer(app *app.App) *Server {
	return &Server{
		app:     app,
		service: app.Service,
		logger:  app.Logger,
		cfg:     app.Cfg,
	}
}

func (s *Server) StartServer(ctx context.Context) <-chan struct{} {
	router := s.setupRouter()

	srv := &http.Server{
		Addr:         ":" + s.cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  s.cfg.HTTPServer.Timeout,
		WriteTimeout: s.cfg.HTTPServer.Timeout,
		IdleTimeout:  s.cfg.HTTPServer.IdleTimeout,
	}

	doneChan := make(chan struct{})

	go func() {
		s.logger.Info("starting server", "port", s.cfg.HTTPServer.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to start server")
		}
	}()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Failed to stop server", "error", err)
		}

		close(doneChan)
	}()

	return doneChan
}

func (s *Server) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New(s.logger))
	r.Use(middleware.Recoverer)

	// API v1 с метриками
	apiRouter := chi.NewRouter()
	apiRouter.Use(metrics.Middleware("api"))
	r.Mount("/v1", v1.New(s.app).Router(apiRouter))

	// Админ интерфейс управления БД с метриками
	adminRouter := chi.NewRouter()
	adminRouter.Use(metrics.Middleware("admin"))
	r.Mount("/admin", admin.New(s.app).Router(adminRouter))

	// Статические файлы с метриками
	r.Get("/video/{filename}", metrics.WrapVideoHandler(video.ServeVideoFile(s.cfg, s.logger)))
	r.Get("/img/{filename}", metrics.WrapImgHandler(img.ServeImgFile(s.cfg, s.logger)))

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	return r
}
