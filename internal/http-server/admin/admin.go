package admin

import (
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/handler/docs"
	"github.com/langowen/bodybalance-backend/internal/storage/redis"
	"github.com/theartofdevel/logging"
	"net/http"
)

type Handler struct {
	logger  *logging.Logger
	storage AdmStorage
	cfg     *config.Config
	redis   *redis.Storage
}

func New(app *app.App) *Handler {
	return &Handler{
		logger:  app.Logger,
		storage: app.Storage.Admin,
		cfg:     app.Cfg,
		redis:   app.Redis,
	}
}

// @title Admin API
// @version 1.0
// @description API для административной панели BodyBalance
// @securityDefinitions.apikey AdminAuth
// @in cookie
// @name token
// @description JWT токен аутентификации администратора (доступен после /admin/signin)

// @BasePath /admin
func (h *Handler) Router(r ...chi.Router) chi.Router {
	var router chi.Router
	if len(r) > 0 {
		router = r[0]
	} else {
		router = chi.NewRouter()
	}

	if h.cfg.Env == "prod" {
		router.Use(h.SecurityHeadersMiddleware)
	}

	router.Post("/signin", h.signing)
	router.Post("/logout", h.logout)

	// Документация
	docs.RegisterRoutes(router, docs.Config{
		User:     h.cfg.Docs.User,
		Password: h.cfg.Docs.Password,
	})

	router.Group(func(r chi.Router) {
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

	fs := http.StripPrefix("/admin/web", http.FileServer(http.Dir("./web")))
	router.Handle("/web/*", fs)
	router.Handle("/web", fs)

	return router
}
