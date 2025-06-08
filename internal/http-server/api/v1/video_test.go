package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	st "github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theartofdevel/logging"
)

func newTestVideoHandler() (*Handler, *MockApiStorage, *MockRedisApi) {
	storage := &MockApiStorage{}
	redis := &MockRedisApi{}
	logger := logging.NewLogger()
	return &Handler{storage: storage, redis: redis, logger: logger}, storage, redis
}

func TestGetVideo_EmptyID(t *testing.T) {
	h, _, _ := newTestVideoHandler()
	req := httptest.NewRequest(http.MethodGet, "/video?video_id=", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetVideo_FromCache(t *testing.T) {
	h, _, redis := newTestVideoHandler()
	vid := &response.VideoResponse{ID: 1, Name: "video1"}
	redis.On("GetVideo", mock.Anything, "1").Return(vid, nil)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=1", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got response.VideoResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, *vid, got)
}

func TestGetVideo_FromStorageAndCacheSet(t *testing.T) {
	h, storage, redis := newTestVideoHandler()
	redis.On("GetVideo", mock.Anything, "2").Return(nil, nil)
	vid := response.VideoResponse{ID: 2, Name: "video2"}
	storage.On("GetVideo", mock.Anything, "2").Return(vid, nil)
	redis.On("SetVideo", mock.Anything, "2", &vid, mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=2", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got response.VideoResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, vid, got)
}

func TestGetVideo_NotFound(t *testing.T) {
	h, storage, redis := newTestVideoHandler()
	redis.On("GetVideo", mock.Anything, "404").Return(nil, nil)
	storage.On("GetVideo", mock.Anything, "404").Return(response.VideoResponse{}, st.ErrVideoNotFound)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=404", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetVideo_StorageError(t *testing.T) {
	h, storage, redis := newTestVideoHandler()
	redis.On("GetVideo", mock.Anything, "err").Return(nil, nil)
	storage.On("GetVideo", mock.Anything, "err").Return(response.VideoResponse{}, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=err", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestGetVideo_InvalidID проверяет обработку некорректного ID видео
func TestGetVideo_InvalidID(t *testing.T) {
	h, _, redis := newTestVideoHandler()

	// Настраиваем мок Redis
	redis.On("GetVideo", mock.Anything, "abc").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=abc", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Читаем содержимое ответа напрямую как строку
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "invalid video ID")
}

// TestGetVideo_RedisError проверяет обработку ошибки Redis при получении видео из кеша
func TestGetVideo_RedisError(t *testing.T) {
	h, storage, redis := newTestVideoHandler()
	redis.On("GetVideo", mock.Anything, "3").Return(nil, errors.New("redis error"))
	vid := response.VideoResponse{ID: 3, Name: "video3"}
	storage.On("GetVideo", mock.Anything, "3").Return(vid, nil)
	redis.On("SetVideo", mock.Anything, "3", &vid, mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=3", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got response.VideoResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, vid, got)
}

// TestGetVideo_RedisSetError проверяет обработку ошибки Redis при установке видео в кеш
func TestGetVideo_RedisSetError(t *testing.T) {
	h, storage, redis := newTestVideoHandler()
	redis.On("GetVideo", mock.Anything, "4").Return(nil, nil)
	vid := response.VideoResponse{ID: 4, Name: "video4"}
	storage.On("GetVideo", mock.Anything, "4").Return(vid, nil)
	redis.On("SetVideo", mock.Anything, "4", &vid, mock.Anything).Return(errors.New("redis set error"))

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=4", nil)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var got response.VideoResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Equal(t, vid, got)
}

// TestGetVideo_ContextCancellation проверяет обработку отмены контекста
func TestGetVideo_ContextCancellation(t *testing.T) {
	h, storage, redis := newTestVideoHandler()

	// Создаем контекст и сразу его отменяем
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	redis.On("GetVideo", mock.Anything, "5").Return(nil, context.Canceled)
	// Добавляем настройку мока для хранилища, которая не должна выполниться из-за отмены контекста,
	// но должна быть настроена для предотвращения паники в тесте
	storage.On("GetVideo", mock.Anything, "5").Return(response.VideoResponse{}, context.Canceled)

	req := httptest.NewRequest(http.MethodGet, "/video?video_id=5", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Проверяем содержимое ответа как строку
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "context canceled")
}

// TestGetVideosByCategoryAndType тестирует функцию получения видео по категориям и типам
func TestGetVideosByCategoryAndType(t *testing.T) {
	h, storage, redis := newTestVideoHandler()

	videos := []response.VideoResponse{
		{ID: 1, Name: "Video 1"},
		{ID: 2, Name: "Video 2"},
	}

	// Настраиваем мок Redis - возвращаем nil, имитируя отсутствие данных в кэше
	redis.On("GetVideosByCategoryAndType", mock.Anything, "1", "2").Return(nil, nil)

	// Настраиваем мок хранилища
	storage.On("GetVideosByCategoryAndType", mock.Anything, "1", "2").Return(videos, nil)

	// Настраиваем мок для установки кэша (вызывается в фоновой горутине)
	redis.On("SetVideosByCategoryAndType", mock.Anything, "1", "2", videos, mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/video_categories?type=1&category=2", nil)
	w := httptest.NewRecorder()

	h.getVideosByCategoryAndType(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []response.VideoResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, videos, resp)
}

// TestGetVideosByCategoryAndType_InvalidType тестирует обработку некорректного типа контента
func TestGetVideosByCategoryAndType_InvalidType(t *testing.T) {
	h, _, _ := newTestVideoHandler()

	// Используем некорректный тип контента
	req := httptest.NewRequest(http.MethodGet, "/video_categories?type=abc&category=2", nil)
	w := httptest.NewRecorder()

	h.getVideosByCategoryAndType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Проверка содержимого ответа
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "Invalid content type")
}

// TestGetVideosByCategoryAndType_EmptyParams тестирует запрос без указания параметров
func TestGetVideosByCategoryAndType_EmptyParams(t *testing.T) {
	h, _, _ := newTestVideoHandler()

	req := httptest.NewRequest(http.MethodGet, "/video_categories", nil)
	w := httptest.NewRecorder()

	h.getVideosByCategoryAndType(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тут не нужно настраивать мок для GetVideosByCategoryAndType,
	// так как из-за пустых параметров метод не должен быть вызван вообще

	// Проверка содержимого ответа
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "Bad Request")
}

// TestGetVideosByCategoryAndType_Error тестирует обработку ошибки получения видео
func TestGetVideosByCategoryAndType_Error(t *testing.T) {
	h, storage, redis := newTestVideoHandler()

	// Настраиваем мок Redis - возвращаем nil, имитируя отсутствие данных в кэше
	redis.On("GetVideosByCategoryAndType", mock.Anything, "1", "2").Return(nil, nil)

	// Настраиваем мок хранилища для возврата ошибки
	storage.On("GetVideosByCategoryAndType", mock.Anything, "1", "2").Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/video_categories?type=1&category=2", nil)
	w := httptest.NewRecorder()

	h.getVideosByCategoryAndType(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Проверка содержимого ответа
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "database error")
}
