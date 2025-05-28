package admin

import (
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/config"
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

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/signin", h.signing)
	r.Post("/logout", h.logout)

	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware) // Защищенные роуты

		// API для работы с видеофайлами
		r.Route("/files", func(r chi.Router) {
			r.Post("/video", h.uploadVideoHandler)
			r.Get("/video", h.listVideoFilesHandler)
			r.Post("/img", h.uploadImageHandler)
			r.Get("/img", h.listImageFilesHandler)
		})

		r.Route("/video", func(r chi.Router) {
			r.Post("/", h.addVideo)
			r.Get("/{id}", h.getVideo)
			r.Get("/", h.getVideos)
			r.Put("/{id}", h.updateVideo)
			r.Delete("/{id}", h.deleteVideo)
		})
	})

	fs := http.StripPrefix("/admin/web", http.FileServer(http.Dir("./web")))
	r.Handle("/web/*", fs)
	r.Handle("/web", fs)

	return r
}

//r.Get("/image", h.getImage)
//r.Get("/category", h.getCategory)
//r.Get("/login", h.getAccount)
//r.Get("/type", h.getType)
