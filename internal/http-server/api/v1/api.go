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

type CategoryResponse struct {
	ID        int             `json:"id"`         // ID из БД
	Category  string          `json:"category"`   // Название категории
	VideoList []storage.Video `json:"videoItems"` // Список видео
}

func New(router *chi.Mux, logger *logging.Logger, storage storage.ApiStorage, cfg *config.Config) {
	h := &Handler{
		logger:  logger,
		storage: storage,
		cfg:     cfg,
	}

	router.Route("/v1", func(r chi.Router) {
		r.Get("/video", h.getVideosByCategoryAndType)
		r.Get("/category", h.getCategoriesByType)
		r.Get("/login", h.checkAccount)
	})
}
