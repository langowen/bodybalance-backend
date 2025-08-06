package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

// @Summary Аутентификация администратора
// @Description Вход в систему с логином и паролем администратора
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param input body dto.SignInRequest true "Данные для входа"
// @Success 200 {object} dto.SuccessResponse "Успешная аутентификация"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/signin [post]
func (h *Handler) signing(w http.ResponseWriter, r *http.Request) {
	const op = "admin.signing"

	logger := h.logger.With(
		"handler", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req dto.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", sl.Err(err))
		dto.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	user, err := h.service.Signing(ctx, req.Login, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrEmptyUsername):
			dto.RespondWithError(w, http.StatusBadRequest, "Login are required")
			return
		case errors.Is(err, admin.ErrEmptyPassword):
			dto.RespondWithError(w, http.StatusBadRequest, "Password are required")
			return
		case errors.Is(err, admin.ErrUserNotFound):
			dto.RespondWithError(w, http.StatusUnauthorized, "Invalid login")
			return
		case errors.Is(err, admin.ErrUserNotAdmin):
			dto.RespondWithError(w, http.StatusForbidden, "Access denied")
			return
		default:
			logger.Error("failed to authenticate user", sl.Err(err))
			dto.RespondWithError(w, http.StatusInternalServerError, "Failed to authenticate")
			return
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.cfg.HTTPServer.TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	})

	tokenString, err := token.SignedString([]byte(h.cfg.HTTPServer.SigningKey))
	if err != nil {
		logger.Error("failed to generate token", sl.Err(err))
		dto.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	h.setAuthCookie(w, tokenString)

	res := dto.SuccessResponse{
		ID:      user.ID,
		Message: "Authentication successful",
	}

	dto.RespondWithJSON(w, http.StatusOK, res)
}

// @Summary Выход из системы
// @Description Завершает сеанс администратора, удаляя токен аутентификации
// @Tags Auth
// @Produce  json
// @Success 200 {object} dto.SuccessResponse "Успешный выход"
// @Router /admin/logout [post]
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	secure := false
	httpOnly := false
	sameSite := http.SameSiteLaxMode

	if h.cfg.Env == "prod" {
		secure = true
		httpOnly = false
		sameSite = http.SameSiteStrictMode

	}

	// Удаляем cookie с токеном
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/admin",
		MaxAge:   -1, // Удалить cookie
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: sameSite,
	})

	res := dto.SuccessResponse{
		Message: "Logged out successfully",
	}
	dto.RespondWithJSON(w, http.StatusOK, res)
}

func (h *Handler) setAuthCookie(w http.ResponseWriter, token string) {

	secure := false
	httpOnly := false
	sameSite := http.SameSiteLaxMode

	if h.cfg.Env == "prod" {
		secure = true
		httpOnly = false //user reverse proxy
		sameSite = http.SameSiteStrictMode

	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/admin",
		MaxAge:   int(h.cfg.HTTPServer.TokenTTL),
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: sameSite,
	})
}
