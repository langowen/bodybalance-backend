package admin

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/port/http-server/admin/dto"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"net/http"
	"strings"
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

// AuthMiddleware проверяет аутентификацию и права администратора
// @security AdminAuth
// @description Требуется JWT токен администратора в cookie с именем "token"
// @dto 401 {object} dto.ErrorResponse "Требуется аутентификация (нет токена)"
// @dto 403 {object} dto.ErrorResponse "Доступ запрещен (недостаточно прав)"
// @dto 400 {object} dto.ErrorResponse "Неверный токен"
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Проверяем HTTPS в production
		if h.cfg.Env == "prod" && r.Header.Get("X-Forwarded-Proto") != "https" {
			h.logger.Warn("HTTPS required", "url", r.URL)
			dto.RespondWithError(w, http.StatusForbidden, "HTTPS required")
			return
		}

		// 2. Проверяем cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			dto.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
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
			dto.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// 4. Проверка администраторских прав
		if !claims.IsAdmin {
			h.logger.Warn("IsAdmin access required", "user", claims.Username)
			dto.RespondWithError(w, http.StatusForbidden, "IsAdmin access required")
			return
		}

		// 5. Обработка CORS
		if h.cfg.Env == "prod" && r.Header.Get("Origin") != "" {

			requestOrigin := r.Header.Get("Origin")

			baseURLDomain := ""
			if baseURL := h.cfg.Media.BaseURL; baseURL != "" {

				if idx := strings.Index(baseURL, "://"); idx != -1 {
					baseURLDomain = baseURL[idx+3:]
				} else {
					baseURLDomain = baseURL
				}
			}

			if r.Host == requestOrigin || (baseURLDomain != "" && strings.Contains(requestOrigin, baseURLDomain)) {
				// Устанавливаем CORS заголовки для того же домена
				w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			} else {
				h.logger.Info("Cross-origin request from external client", "origin", requestOrigin)
			}

			// Для preflight запросов (OPTIONS)
			if r.Method == http.MethodOptions {
				// Разрешаем нужные HTTP методы
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

				// Разрешаем нужные заголовки
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

				// Время кэширования preflight ответа
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 часа

				w.WriteHeader(http.StatusOK)
				return
			}
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

		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' data: https://code.jquery.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://unpkg.com; " +
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://fonts.googleapis.com; " +
			"img-src 'self' data: https://*; " +
			"font-src 'self' data: https://cdn.jsdelivr.net https://fonts.gstatic.com; " +
			"connect-src 'self' https://*; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"frame-ancestors 'none'"

		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}
