package http_server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/api/v1"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/handler/img"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/handler/video"
	mwLogger "github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/logger"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Server struct {
	app *app.App
}

func Init(app *app.App) *http.Server {
	s := &Server{app: app}
	router := s.setupRouter()

	serverConfig := &http.Server{
		Addr:         ":" + app.Cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  app.Cfg.HTTPServer.Timeout,
		WriteTimeout: app.Cfg.HTTPServer.Timeout,
		IdleTimeout:  app.Cfg.HTTPServer.IdleTimeout,
	}

	return serverConfig
}

func (s *Server) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New(s.app.Logger))
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
	r.Get("/video/{filename}", metrics.WrapVideoHandler(video.ServeVideoFile(s.app.Cfg, s.app.Logger)))
	r.Get("/img/{filename}", metrics.WrapImgHandler(img.ServeImgFile(s.app.Cfg, s.app.Logger)))

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	return r
}
