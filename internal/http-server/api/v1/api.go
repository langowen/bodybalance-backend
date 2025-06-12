// Package v1 содержит обработчики API версии 1
// @title BodyBalance API
// @version 1.0
// @description API для управления видео-контентом BodyBalance.
// @contact.name Sergei
// @contact.email info@7375.org
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host body.7375.org
// @BasePath /v1
// @schemes https
package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/app"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/storage/redis"
	"github.com/theartofdevel/logging"
)

type Handler struct {
	logger  *logging.Logger
	storage ApiStorage
	redis   *redis.Storage
	cfg     *config.Config
}

func New(app *app.App) *Handler {
	return &Handler{
		logger:  app.Logger,
		storage: app.Storage.Api,
		redis:   app.Redis,
		cfg:     app.Cfg,
	}
}

func (h *Handler) Router(r ...chi.Router) chi.Router {
	var router chi.Router
	if len(r) > 0 {
		router = r[0]
	} else {
		router = chi.NewRouter()
	}

	router.Get("/video_categories", h.getVideosByCategoryAndType)
	router.Get("/video", h.getVideo)
	router.Get("/category", h.getCategoriesByType)
	router.Get("/login", h.checkAccount)

	return router
}
