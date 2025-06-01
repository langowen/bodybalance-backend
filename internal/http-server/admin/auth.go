package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"net/http"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

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

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	// Удаляем cookie с токеном
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/admin", // Должен совпадать с путем установки
		MaxAge:   -1,       // Удалить cookie
		HttpOnly: true,
		Secure:   false,                //h.cfg.Env == "prod"
		SameSite: http.SameSiteLaxMode, ///http.SameSiteStrictMode
	})

	admResponse.RespondWithJSON(w, http.StatusOK, admResponse.SignInResponse{
		Message: "Logged out successfully",
	})
}

func (h *Handler) setAuthCookie(w http.ResponseWriter, token string) {
	//secure := h.cfg.Env == "prod" // h.cfg.Env == "prod"
	sameSite := http.SameSiteLaxMode //http.SameSiteStrictMode
	if h.cfg.Env == "dev" {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/admin",
		MaxAge:   int(h.cfg.HTTPServer.TokenTTL),
		HttpOnly: true,
		Secure:   false,
		SameSite: sameSite,
	})
}
