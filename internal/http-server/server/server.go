package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1"
	"github.com/langowen/bodybalance-backend/internal/http-server/handler/video"
	mwLogger "github.com/langowen/bodybalance-backend/internal/http-server/middleware/logger"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres"
	"github.com/theartofdevel/logging"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	cfg       *config.Config
	logger    *logging.Logger
	pgStorage *postgres.Storage
}

func Init(cfg *config.Config, logger *logging.Logger, pgStorage *postgres.Storage) *http.Server {
	router := (&Server{
		cfg:       cfg,
		logger:    logger,
		pgStorage: pgStorage,
	}).setupRouter()

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	return srv
}

func (s *Server) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New(s.logger)) //TODO подумать над заменой логера из пакета https://github.com/go-chi/httplog/blob/master/httplog.go
	r.Use(middleware.Recoverer)

	// Настройка Basic Auth
	docsAuth := middleware.BasicAuth("Restricted Docs", map[string]string{
		s.cfg.Docs.User: s.cfg.Docs.Password,
	})

	// Группа защищенных роутов для документации
	r.Route("/", func(r chi.Router) {
		r.Use(docsAuth)

		// Swagger JSON
		r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, filepath.Join(getProjectRoot(), "docs/swagger.json"))
		})

		// RapiDoc UI
		staticPath := filepath.Join(getProjectRoot(), "docs/static")
		r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(staticPath, "rapidoc.html"))
		})
		r.Handle("/docs/*", http.StripPrefix("/docs", http.FileServer(http.Dir(staticPath))))
	})

	v1.New(r, s.logger, s.pgStorage, s.cfg)

	r.Get("/video/{filename}", video.ServeVideoFile(s.cfg, s.logger))

	return r
}
func getProjectRoot() string {
	dir, _ := os.Getwd()
	return dir
}
