package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/theartofdevel/logging"
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	TypeID   int    `json:"type_id"`
	TypeName string `json:"type_name"`
}

type Handler struct {
	logger  *logging.Logger
	storage ApiStorage
	redis   RedisApi
	cfg     *config.Config
}

func New(logger *logging.Logger, storage ApiStorage, redis RedisApi, cfg *config.Config) *Handler {
	return &Handler{
		logger:  logger,
		storage: storage,
		redis:   redis,
		cfg:     cfg,
	}
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	// Публичные маршруты
	r.Get("/login", h.checkAccount)

	// Защищенные маршруты (требуют аутентификации)
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)

		r.Get("/video_categories", h.getVideosByCategoryAndType)
		r.Get("/video", h.getVideo)
		r.Get("/category", h.getCategoriesByType)
	})

	return r
}
