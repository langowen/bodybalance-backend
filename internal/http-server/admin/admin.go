package admin

import (
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/handler/docs"
	"github.com/theartofdevel/logging"
	"net/http"
)

type Handler struct {
	logger  *logging.Logger
	storage AdmStorage
	cfg     *config.Config
}

func New(logger *logging.Logger, storage AdmStorage, cfg *config.Config) *Handler {
	return &Handler{
		logger:  logger,
		storage: storage,
		cfg:     cfg,
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
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/signin", h.signing)
	r.Post("/logout", h.logout)

	// Документация
	docs.RegisterRoutes(r, docs.Config{
		User:     h.cfg.Docs.User,
		Password: h.cfg.Docs.Password,
	})

	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware) // Защищенные роуты

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
	r.Handle("/web/*", fs)
	r.Handle("/web", fs)

	return r
}
