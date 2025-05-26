package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/theartofdevel/logging"
)

type Handler struct {
	logger  *logging.Logger
	storage ApiStorage
}

func New(logger *logging.Logger, storage ApiStorage) *Handler {
	return &Handler{
		logger:  logger,
		storage: storage,
	}
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/video_categories", h.getVideosByCategoryAndType)
	r.Get("/video", h.getVideo)
	r.Get("/category", h.getCategoriesByType)
	r.Get("/login", h.checkAccount)

	return r
}
