package v1

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"net/http"
	"time"
)

// AuthMiddleware проверяет аутентификацию пользователя
// @description Требуется JWT токен пользователя в cookie с именем "user_token"
// @response 401 {object} response.ErrorResponse "Требуется аутентификация (нет токена)"
// @response 400 {object} response.ErrorResponse "Неверный токен"
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем cookie
		cookie, err := r.Cookie("user_token")
		if err != nil {
			response.RespondWithError(w, http.StatusUnauthorized, "Authentication required", "No valid token found")
			return
		}

		// Валидация токена
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(h.cfg.HTTPServer.SigningKey), nil
		})

		if err != nil || !token.Valid {
			h.logger.Warn("Invalid token", sl.Err(err))
			response.RespondWithError(w, http.StatusUnauthorized, "Invalid token", "Authentication failed")
			return
		}

		// Устанавливаем данные пользователя в контекст для дальнейшего использования
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", claims)

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetAuthCookie устанавливает JWT токен в cookie
func (h *Handler) SetAuthCookie(w http.ResponseWriter, username string, typeID float64, typeName string) error {
	// Создаем токен с данными пользователя
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.cfg.HTTPServer.TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Username: username,
		TypeID:   int(typeID),
		TypeName: typeName,
	})

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(h.cfg.HTTPServer.SigningKey))
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Определяем настройки безопасности в зависимости от окружения
	secure := false
	httpOnly := true
	sameSite := http.SameSiteLaxMode

	if h.cfg.Env == "prod" {
		secure = true
		sameSite = http.SameSiteStrictMode
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_token",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   int(h.cfg.HTTPServer.TokenTTL.Seconds()),
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: sameSite,
	})

	return nil
}
