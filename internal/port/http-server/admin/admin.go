// Package admin содержит обработчики административного API
// @title BodyBalance IsAdmin API
// @version 1.0
// @description API для управления административной частью BodyBalance.
// @contact.name Sergei
// @contact.email info@7375.org
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host body.7375.org
// @BasePath /admin
// @schemes https
// @securityDefinitions.apikey AdminAuth
// @in cookie
// @name token
package admin

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/handler/docs"
	"github.com/theartofdevel/logging"
)

//go:embed web/*
var staticFiles embed.FS

type Handler struct {
	logger  *logging.Logger
	cfg     *config.Config
	service Service
}

func New(app *app.App) *Handler {
	return &Handler{
		logger:  app.Logger,
		cfg:     app.Cfg,
		service: app.ServiceAdmin,
	}
}

func (h *Handler) Router(r chi.Router) chi.Router {
	if h.cfg.Env == "prod" {
		r.Use(h.SecurityHeadersMiddleware)
	}

	r.Post("/signin", h.signing)
	r.Post("/logout", h.logout)

	// Документация
	docsRouter := docs.Routes()
	r.Mount("/swagger", docsRouter)

	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)

		// API для работы с файлами
		r.Route("/files", func(r chi.Router) {
			r.Post("/video", h.uploadVideoHandler)
			r.Get("/video", h.listVideoFilesHandler)
			r.Post("/img", h.uploadImageHandler)
			r.Get("/img", h.listImageFilesHandler)
		})
		// API для работы с видео
		r.Route("/video", func(r chi.Router) {
			r.Post("/", h.addVideo)
			r.Get("/{id}", h.getVideo)
			r.Get("/", h.getVideos)
			r.Put("/{id}", h.updateVideo)
			r.Delete("/{id}", h.deleteVideo)
		})

		// API для работы с типами
		r.Route("/type", func(r chi.Router) {
			r.Post("/", h.addType)
			r.Get("/{id}", h.getType)
			r.Get("/", h.getTypes)
			r.Put("/{id}", h.updateType)
			r.Delete("/{id}", h.deleteType)
		})

		// API для работы с пользователями
		r.Route("/users", func(r chi.Router) {
			r.Post("/", h.addUser)
			r.Get("/{id}", h.getUser)
			r.Get("/", h.getUsers)
			r.Put("/{id}", h.updateUser)
			r.Delete("/{id}", h.deleteUser)
		})

		// API для работы с категориями
		r.Route("/category", func(r chi.Router) {
			r.Post("/", h.addCategory)
			r.Get("/{id}", h.getCategory)
			r.Get("/", h.getCategories)
			r.Put("/{id}", h.updateCategory)
			r.Delete("/{id}", h.deleteCategory)
		})
	})

	webFS, err := fs.Sub(staticFiles, "web")
	if err != nil {
		err = fmt.Errorf("admin: failed to initialize embedded filesystem: %w", err)
		panic(err)
	}

	fileServer := http.FileServer(http.FS(webFS))

	// Группа для статических файлов
	r.Route("/web", func(r chi.Router) {
		r.Handle("/*", http.StripPrefix("/admin/web", fileServer))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			data, err := fs.ReadFile(webFS, "index.html")
			if err != nil {
				http.Error(w, "Admin UI not found", http.StatusNotFound)
				return
			}

			w.Write(data)
		})
	})

	return r
}
