package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/theartofdevel/logging"
)

type Handler struct {
	logger  *logging.Logger
	storage storage.ApiStorage
	cfg     *config.Config
}

func New(router *chi.Mux, logger *logging.Logger, storage storage.ApiStorage, cfg *config.Config) {
	h := &Handler{
		logger:  logger,
		storage: storage,
		cfg:     cfg,
	}

	router.Route("/v1", func(r chi.Router) {
		r.Get("/video_categories", h.getVideosByCategoryAndType)
		r.Get("/video", h.getVideo)
		r.Get("/category", h.getCategoriesByType)
		r.Get("/login", h.checkAccount)
	})
}
