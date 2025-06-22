package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redismock/v9"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/logdiscart"
	pgapi "github.com/langowen/bodybalance-backend/internal/storage/postgres/api"
	"github.com/langowen/bodybalance-backend/internal/storage/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// mockStorage - мок для хранилища с реализацией метода Feedback
type mockStorage struct {
	pgapi.Storage
	feedbackErr error
}

// Feedback - реализация метода Feedback для мока
func (m *mockStorage) Feedback(ctx context.Context, feedback *response.FeedbackResponse) error {
	return m.feedbackErr
}

// newTestHandlerWithFeedbackMock создает тестовый обработчик с моком для метода Feedback
func newTestHandlerWithFeedbackMock(t *testing.T, feedbackErr error) *Handler {
	// Создаем конфиг
	cfg := &config.Config{
		Redis: config.Redis{
			CacheTTL: time.Hour,
			Enable:   true,
		},
		Env: "test",
	}

	// Создаем Redis-клиент с моком
	redisCli, _ := redismock.NewClientMock()
	redisStorage := redis.NewStorage(redisCli, cfg)

	// Создаем мок для хранилища
	mockPgStorage := &mockStorage{
		feedbackErr: feedbackErr,
	}

	// Создаем логгер
	logger := logdiscart.NewDiscardLogger()

	// Создаем хендлер
	handler := &Handler{
		storage: mockPgStorage,
		redis:   redisStorage,
		logger:  logger,
		cfg:     cfg,
	}

	return handler
}

func TestFeedback_ValidRequestWithEmail(t *testing.T) {
	// Создаем мок хендлера без ошибок при сохранении фидбека
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка запроса
	feedback := response.FeedbackResponse{
		Name:    "Тест Тестович",
		Email:   "test@example.com",
		Message: "Отличное приложение!",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем ответ сервера - должна быть строка
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "Feedback successfully saved")
}

func TestFeedback_ValidRequestWithTelegram(t *testing.T) {
	// Создаем мок хендлера без ошибок при сохранении фидбека
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка запроса
	feedback := response.FeedbackResponse{
		Name:     "Тест Тестович",
		Telegram: "@test_user",
		Message:  "Отличное приложение!",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем ответ сервера - должна быть строка
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "Feedback successfully saved")
}

func TestFeedback_NoMessage(t *testing.T) {
	// Создаем мок хендлера (без ошибок, т.к. до вызова storage.Feedback не дойдет)
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка запроса без сообщения
	feedback := response.FeedbackResponse{
		Name:  "Тест Тестович",
		Email: "test@example.com",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)
	assert.Contains(t, respBody["error"], "Message is required")
}

func TestFeedback_InvalidEmail(t *testing.T) {
	// Создаем мок хендлера (без ошибок, т.к. до вызова storage.Feedback не дойдет)
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка запроса с некорректным email
	feedback := response.FeedbackResponse{
		Name:    "Тест Тестович",
		Email:   "invalid-email",
		Message: "Отличное приложение!",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)
	assert.Contains(t, respBody["error"], "Invalid email format")
}

func TestFeedback_InvalidTelegram(t *testing.T) {
	// Создаем мок хендлера (без ошибок, т.к. до вызова storage.Feedback не дойдет)
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка запроса с некорректным telegram (без @)
	feedback := response.FeedbackResponse{
		Name:     "Тест Тестович",
		Telegram: "invalid_telegram",
		Message:  "Отличное приложение!",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)
	assert.Contains(t, respBody["error"], "Telegram must start with @")
}

func TestFeedback_TooShortTelegram(t *testing.T) {
	// Создаем мок хендлера (без ошибок, т.к. до вызова storage.Feedback не дойдет)
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка запроса с слишком коротким telegram
	feedback := response.FeedbackResponse{
		Name:     "Тест Тестович",
		Telegram: "@abc",
		Message:  "Отличное приложение!",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)
	assert.Contains(t, respBody["error"], "Telegram must start with @ and contain 5-32 characters")
}

func TestFeedback_NoContactMethod(t *testing.T) {
	// Создаем мок хендлера (без ошибок, т.к. до вызова storage.Feedback не дойдет)
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка запроса без контактной информации
	feedback := response.FeedbackResponse{
		Name:    "Тест Тестович",
		Message: "Отличное приложение!",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)
	assert.Contains(t, respBody["error"], "Either email or telegram must be provided")
}

func TestFeedback_InvalidJSON(t *testing.T) {
	// Создаем мок хендлера (без ошибок, т.к. до вызова storage.Feedback не дойдет)
	h := newTestHandlerWithFeedbackMock(t, nil)

	// Подготовка некорректного JSON
	invalidJSON := []byte(`{"name": "Test", "email": "test@example.com", "message":`)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	require.NoError(t, err)
	assert.Contains(t, respBody["error"], "Invalid request format")
}

func TestFeedback_DatabaseError(t *testing.T) {
	// Создаем мок хендлера с симуляцией ошибки базы данных
	h := newTestHandlerWithFeedbackMock(t, errors.New("db connection error"))

	// Подготовка запроса
	feedback := response.FeedbackResponse{
		Name:    "Тест Тестович",
		Email:   "test@example.com",
		Message: "Отличное приложение!",
	}
	body, err := json.Marshal(feedback)
	require.NoError(t, err)

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Вызываем метод
	h.feedback(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Проверяем ответ сервера - должен содержать информацию об ошибке
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "Server Error")
}
