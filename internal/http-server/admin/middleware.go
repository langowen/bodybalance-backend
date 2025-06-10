package admin

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"net/http"
)

// AuthMiddleware проверяет аутентификацию и права администратора
// @security AdminAuth
// @description Требуется JWT токен администратора в cookie с именем "token"
// @response 401 {object} admResponse.ErrorResponse "Требуется аутентификация (нет токена)"
// @response 403 {object} admResponse.ErrorResponse "Доступ запрещен (недостаточно прав)"
// @response 400 {object} admResponse.ErrorResponse "Неверный токен"
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Проверяем HTTPS в production
		if h.cfg.Env == "prod" && r.Header.Get("X-Forwarded-Proto") != "https" {
			h.logger.Warn("HTTPS required", "url", r.URL)
			admResponse.RespondWithError(w, http.StatusForbidden, "HTTPS required")
			return
		}

		// 2. Проверяем cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			admResponse.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// 3. Валидация токена
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(h.cfg.HTTPServer.SigningKey), nil
		})

		if err != nil || !token.Valid {
			h.logger.Warn("Invalid token", sl.Err(err))
			admResponse.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// 4. Проверка администраторских прав
		if !claims.IsAdmin {
			h.logger.Warn("Admin access required", "user", claims.Username)
			admResponse.RespondWithError(w, http.StatusForbidden, "Admin access required")
			return
		}

		// 5. Проверка SameSite (дополнительно)
		if h.cfg.Env == "prod" && r.Header.Get("Origin") != "" {
			h.logger.Warn("Cross-site request attempted", "origin", r.Header.Get("Origin"))
			admResponse.RespondWithError(w, http.StatusForbidden, "Cross-site requests not allowed")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware добавляет стандартные security headers ко всем ответам
func (h *Handler) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		// CSP с разрешенными внешними ресурсами
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://code.jquery.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; " +
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; " +
			"img-src 'self' data:; " +
			"font-src 'self' https://cdn.jsdelivr.net; " +
			"connect-src 'self'; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'none'"

		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}
