package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/storage/redis"
	"github.com/theartofdevel/logging"
)

type Handler struct {
	logger  *logging.Logger
	storage ApiStorage
	redis   *redis.Storage
}

func New(logger *logging.Logger, storage ApiStorage, redis *redis.Storage) *Handler {
	return &Handler{
		logger:  logger,
		storage: storage,
		redis:   redis,
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
