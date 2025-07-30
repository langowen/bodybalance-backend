package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/admResponse"
	"net/http"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

// @Summary Аутентификация администратора
// @Description Вход в систему с логином и паролем администратора
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param input body admResponse.SignInRequest true "Данные для входа"
// @Success 200 {object} admResponse.SignInResponse
// @Failure 400 {object} admResponse.ErrorResponse
// @Failure 401 {object} admResponse.ErrorResponse
// @Failure 403 {object} admResponse.ErrorResponse
// @Failure 500 {object} admResponse.ErrorResponse
// @Router /admin/signin [post]
func (h *Handler) signing(w http.ResponseWriter, r *http.Request) {
	const op = "admin.signing"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req admResponse.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Login == "" || req.Password == "" {
		logger.Error("empty login or password")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Login and password are required")
		return
	}

	ctx := r.Context()
	user, err := h.storage.GetAdminUser(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("invalid login credentials")
			admResponse.RespondWithError(w, http.StatusUnauthorized, "Invalid login or password")
			return
		}
		logger.Error("failed to get user", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to authenticate")
		return
	}

	if !user.IsAdmin {
		logger.Warn("user is not admin", "username", req.Login)
		admResponse.RespondWithError(w, http.StatusForbidden, "Access denied")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.cfg.HTTPServer.TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Username: req.Login,
		IsAdmin:  true,
	})

	tokenString, err := token.SignedString([]byte(h.cfg.HTTPServer.SigningKey))
	if err != nil {
		logger.Error("failed to generate token", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Используем единый метод установки cookie
	h.setAuthCookie(w, tokenString)

	admResponse.RespondWithJSON(w, http.StatusOK, admResponse.SignInResponse{
		Message: "Authentication successful",
	})
}

// @Summary Выход из системы
// @Description Завершает сеанс администратора, удаляя токен аутентификации
// @Tags Auth
// @Produce  json
// @Success 200 {object} admResponse.SignInResponse
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

	admResponse.RespondWithJSON(w, http.StatusOK, admResponse.SignInResponse{
		Message: "Logged out successfully",
	})
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
