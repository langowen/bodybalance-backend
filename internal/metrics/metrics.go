package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AdminRequestsTotal счетчик запросов к административному API
	AdminRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bodybalance_admin_requests_total",
			Help: "Общее количество запросов к административному API",
		},
		[]string{"method", "endpoint", "status"},
	)

	// APIRequestsTotal счетчик запросов к API для мобильного приложения
	APIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bodybalance_api_requests_total",
			Help: "Общее количество запросов к API для мобильного приложения",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RequestDuration гистограмма времени обработки запросов
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bodybalance_request_duration_seconds",
			Help:    "Длительность обработки запросов в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"handler", "method", "endpoint"},
	)

	// ActiveRequests количество активных запросов
	ActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bodybalance_active_requests",
			Help: "Количество активных запросов",
		},
		[]string{"handler"},
	)

	// StaticFileRequests счетчик запросов к статическим файлам
	StaticFileRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bodybalance_static_file_requests_total",
			Help: "Общее количество запросов к статическим файлам",
		},
		[]string{"file_type", "filename", "status"},
	)

	// StaticFileSize размер отданных статических файлов в байтах
	StaticFileSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bodybalance_static_file_size_bytes",
			Help:    "Размер отправленных статических файлов в байтах",
			Buckets: prometheus.ExponentialBuckets(1024, 4, 10), // 1KB, 4KB, 16KB, 64KB, 256KB, 1MB, 4MB, 16MB, 64MB, 256MB
		},
		[]string{"file_type", "filename"},
	)

	// StaticFileTransferRate скорость передачи файла в байтах/секунду
	StaticFileTransferRate = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bodybalance_static_file_transfer_rate_bytes_per_second",
			Help:    "Скорость передачи статических файлов в байтах/секунду",
			Buckets: prometheus.ExponentialBuckets(1024*10, 2, 10), // 10KB/s, 20KB/s, 40KB/s и т.д.
		},
		[]string{"file_type", "filename"},
	)

	// StaticFileLoadTime время загрузки файла в секундах
	StaticFileLoadTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bodybalance_static_file_load_time_seconds",
			Help:    "Время загрузки статических файлов в секундах",
			Buckets: prometheus.LinearBuckets(0.01, 0.05, 20), // 0.01s, 0.06s, 0.11s и т.д. до ~1 секунды
		},
		[]string{"file_type", "filename"},
	)

	// StaticFileCacheHit счетчик попаданий в кэш (по заголовку If-None-Match/If-Modified-Since)
	StaticFileCacheHit = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bodybalance_static_file_cache_hits_total",
			Help: "Количество запросов к файлам, обработанных из кэша (304 Not Modified)",
		},
		[]string{"file_type", "filename"},
	)
)
