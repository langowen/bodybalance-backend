package metrics

import (
	"fmt"
	"github.com/langowen/bodybalance-backend/internal/port/metrics"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func Middleware(handlerType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Оборачиваем ResponseWriter для перехвата статуса ответа
			ww := NewWrapResponseWriter(w)

			// Увеличиваем счетчик активных запросов
			metrics.ActiveRequests.WithLabelValues(handlerType).Inc()
			defer metrics.ActiveRequests.WithLabelValues(handlerType).Dec()

			// Вызываем следующий обработчик
			next.ServeHTTP(ww, r)

			// Получаем путь и очищаем от параметров ID
			path := getCleanPath(r)

			// Записываем метрики в зависимости от типа обработчика
			duration := time.Since(start).Seconds()
			status := fmt.Sprintf("%d", ww.Status())

			// Записываем метрики количества запросов
			if handlerType == "admin" {
				metrics.AdminRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			} else if handlerType == "api" {
				metrics.APIRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			}

			// Записываем метрики длительности
			metrics.RequestDuration.WithLabelValues(handlerType, r.Method, path).Observe(duration)
		})
	}
}

// Очищает путь запроса от конкретных ID для лучшей группировки метрик
func getCleanPath(r *http.Request) string {
	// Получаем текущий маршрут из Chi (если используется)
	routeContext := chi.RouteContext(r.Context())
	if routeContext != nil && routeContext.RoutePath != "" {
		return routeContext.RoutePath
	}

	// Универсальная очистка ID из URL (например, /users/123 -> /users/{id})
	parts := strings.Split(r.URL.Path, "/")
	for i, part := range parts {
		// Проверяем, похож ли сегмент на ID
		if i > 0 && isNumeric(part) {
			parts[i] = "{id}"
		}
	}
	return strings.Join(parts, "/")
}

// Проверяет, является ли строка числовой (предположительно ID)
func isNumeric(s string) bool {
	// Простая эвристика: если строка состоит только из цифр и имеет длину > 0
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// WrapResponseWriter используется для перехвата статуса ответа
type WrapResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func NewWrapResponseWriter(w http.ResponseWriter) *WrapResponseWriter {
	return &WrapResponseWriter{ResponseWriter: w}
}

func (w *WrapResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *WrapResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *WrapResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
