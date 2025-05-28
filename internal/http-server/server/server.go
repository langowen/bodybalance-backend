package server

import (
	"crypto/tls"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1"
	"github.com/langowen/bodybalance-backend/internal/http-server/handler/docs"
	"github.com/langowen/bodybalance-backend/internal/http-server/handler/img"
	"github.com/langowen/bodybalance-backend/internal/http-server/handler/video"
	mwLogger "github.com/langowen/bodybalance-backend/internal/http-server/middleware/logger"
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

	// Настраиваем TLS только для production
	if app.Cfg.Env == "prod" {
		serverConfig.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			// Дополнительные настройки безопасности для production
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	return serverConfig
}

func (s *Server) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New(s.app.Logger))
	r.Use(middleware.Recoverer)

	// Документация
	docs.RegisterRoutes(r, docs.Config{
		User:     s.app.Cfg.Docs.User,
		Password: s.app.Cfg.Docs.Password,
	})

	// API v1
	r.Mount("/v1", v1.New(s.app.Logger, s.app.Storage.Api).Router())

	// Админ интерфейс управления БД
	r.Mount("/admin", admin.New(
		s.app.Logger,
		s.app.Storage.Admin,
		s.app.Cfg,
	).Router())

	// Статические файлы
	r.Get("/video/{filename}", video.ServeVideoFile(s.app.Cfg, s.app.Logger))
	r.Get("/img/{filename}", img.ServeImgFile(s.app.Cfg, s.app.Logger))

	return r
}
