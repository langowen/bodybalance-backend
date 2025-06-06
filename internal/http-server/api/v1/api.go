package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/theartofdevel/logging"
)

// Handler представляет обработчики API v1
// @description Основной обработчик API v1
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

// @title BodyBalance Backend API
// @version 1.0
// @description API for BodyBalance application

// @contact.name Sergei
// @contact.email info@7375.org

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.7375.org
// @BasePath /v1
// @schemes https
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/video_categories", h.getVideosByCategoryAndType)
	r.Get("/video", h.getVideo)
	r.Get("/category", h.getCategoriesByType)
	r.Get("/login", h.checkAccount)

	return r
}
