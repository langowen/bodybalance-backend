package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// SimpleLimiter ограничивает только 4xx ошибки.
type SimpleLimiter struct {
	window      time.Duration
	maxErrors   int
	banDuration time.Duration

	mu     sync.RWMutex
	errors map[string][]time.Time
	banned map[string]time.Time
}

func NewSimpleLimiter(window time.Duration, maxErrors int, banDuration time.Duration) *SimpleLimiter {
	limiter := &SimpleLimiter{
		window:      window,
		maxErrors:   maxErrors,
		banDuration: banDuration,
		errors:      make(map[string][]time.Time),
		banned:      make(map[string]time.Time),
	}

	go limiter.cleanup()
	return limiter
}

func (l *SimpleLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		l.mu.RLock()
		if until, ok := l.banned[ip]; ok && time.Now().Before(until) {
			l.mu.RUnlock()
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		l.mu.RUnlock()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Учет ошибок 4xx
		if rec.status >= 400 && rec.status < 500 {
			l.recordError(ip)
		}
	})
}

func (l *SimpleLimiter) recordError(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Очистка старых ошибок для этого IP
	errors := l.errors[ip]
	j := 0
	for _, t := range errors {
		if t.After(cutoff) {
			errors[j] = t
			j++
		}
	}
	errors = errors[:j]

	// Добавление новой ошибки
	errors = append(errors, now)
	l.errors[ip] = errors

	// Проверка на превышение лимита
	if len(errors) >= l.maxErrors {
		l.banned[ip] = now.Add(l.banDuration)
		delete(l.errors, ip) // очищаем историю после бана
	}
}

func (l *SimpleLimiter) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.cleanupOldData()
		}
	}
}

func (l *SimpleLimiter) cleanupOldData() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	errorsCutoff := now.Add(-l.window * 2)
	bannedCutoff := now.Add(-l.banDuration)

	// Очистка старых банов
	for ip, until := range l.banned {
		if until.Before(bannedCutoff) {
			delete(l.banned, ip)
		}
	}

	// Очистка старых ошибок
	for ip, errors := range l.errors {
		j := 0
		for _, t := range errors {
			if t.After(errorsCutoff) {
				errors[j] = t
				j++
			}
		}
		if j == 0 {
			delete(l.errors, ip)
		} else {
			l.errors[ip] = errors[:j]
		}
	}
}
