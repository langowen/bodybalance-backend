package metrics

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/metrics"
	"net/http"
)

// Источники данных
const (
	SourceRedis = "redis"
	SourceSQL   = "sql"
)

type contextKey string

const dataSourceContextKey = contextKey("data_source")

// WithDataSource добавляет информацию об источнике данных в контекст запроса
func WithDataSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, dataSourceContextKey, source)
}

// GetDataSource возвращает источник данных из контекста, если он был установлен
func GetDataSource(ctx context.Context) (string, bool) {
	source, ok := ctx.Value(dataSourceContextKey).(string)
	return source, ok
}

// RecordDataSource увеличивает счётчик для указанного источника данных
func RecordDataSource(r *http.Request, source string) {
	method := r.Method
	path := r.URL.Path

	metrics.DataSourceRequests.WithLabelValues(method, path, source).Inc()
}
