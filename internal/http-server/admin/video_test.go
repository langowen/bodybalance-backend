package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theartofdevel/logging"
)

func newTestVideoHandlerWithMocks() (*Handler, *MockAdmStorage, *MockRedisClient) {
	storage := &MockAdmStorage{}
	redis := &MockRedisClient{}
	logger := logging.NewLogger()
	return &Handler{storage: storage, redis: redis, logger: logger}, storage, redis
}

func muxSetURLParamVideo(r *http.Request, key, val string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

func TestAddVideo_Success(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	reqBody := admResponse.VideoRequest{URL: "url", Name: "name", CategoryIDs: []int64{1, 2}}
	storage.On("AddVideo", mock.Anything, reqBody).Return(int64(10), nil)
	storage.On("AddVideoCategories", mock.Anything, int64(10), []int64{1, 2}).Return(nil)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addVideo(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	storage.AssertExpectations(t)
}

func TestAddVideo_InvalidJSON(t *testing.T) {
	h, _, _ := newTestVideoHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewReader([]byte("{")))
	w := httptest.NewRecorder()

	h.addVideo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddVideo_EmptyFields(t *testing.T) {
	h, _, _ := newTestVideoHandlerWithMocks()
	reqBody := admResponse.VideoRequest{URL: "", Name: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addVideo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddVideo_StorageError(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	reqBody := admResponse.VideoRequest{URL: "url", Name: "name"}
	storage.On("AddVideo", mock.Anything, reqBody).Return(int64(0), errors.New("fail"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addVideo(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAddVideo_CategoriesError(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	reqBody := admResponse.VideoRequest{URL: "url", Name: "name", CategoryIDs: []int64{1, 2}}
	storage.On("AddVideo", mock.Anything, reqBody).Return(int64(10), nil)
	storage.On("AddVideoCategories", mock.Anything, int64(10), []int64{1, 2}).Return(errors.New("fail"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/video", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addVideo(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetVideo_Success(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	videoID := int64(10)
	videoResp := admResponse.VideoResponse{ID: videoID, Name: "name"}
	categories := []admResponse.CategoryResponse{{ID: 1, Name: "cat1"}}
	storage.On("GetVideo", mock.Anything, videoID).Return(videoResp, nil)
	storage.On("GetVideoCategories", mock.Anything, videoID).Return(categories, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/video/10", nil)
	req = muxSetURLParamVideo(req, "id", "10")
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	storage.AssertExpectations(t)
}

func TestGetVideo_InvalidID(t *testing.T) {
	h, _, _ := newTestVideoHandlerWithMocks()
	req := httptest.NewRequest(http.MethodGet, "/admin/video/abc", nil)
	req = muxSetURLParamVideo(req, "id", "abc")
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetVideo_NotFound(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	storage.On("GetVideo", mock.Anything, int64(10)).Return(admResponse.VideoResponse{}, sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/admin/video/10", nil)
	req = muxSetURLParamVideo(req, "id", "10")
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetVideo_StorageError(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	storage.On("GetVideo", mock.Anything, int64(10)).Return(admResponse.VideoResponse{}, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/admin/video/10", nil)
	req = muxSetURLParamVideo(req, "id", "10")
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetVideo_CategoriesError(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	videoID := int64(10)
	videoResp := admResponse.VideoResponse{ID: videoID, Name: "name"}
	storage.On("GetVideo", mock.Anything, videoID).Return(videoResp, nil)
	storage.On("GetVideoCategories", mock.Anything, videoID).Return([]admResponse.CategoryResponse{}, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/admin/video/10", nil)
	req = muxSetURLParamVideo(req, "id", "10")
	w := httptest.NewRecorder()

	h.getVideo(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetVideos_Success(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	videos := []admResponse.VideoResponse{{ID: 1, Name: "v1"}, {ID: 2, Name: "v2"}}
	categories := []admResponse.CategoryResponse{{ID: 1, Name: "cat1"}}
	storage.On("GetVideos", mock.Anything).Return(videos, nil)
	storage.On("GetVideoCategories", mock.Anything, int64(1)).Return(categories, nil)
	storage.On("GetVideoCategories", mock.Anything, int64(2)).Return(categories, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/video", nil)
	w := httptest.NewRecorder()

	h.getVideos(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	storage.AssertExpectations(t)
}

func TestGetVideos_StorageError(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	storage.On("GetVideos", mock.Anything).Return([]admResponse.VideoResponse{}, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/admin/video", nil)
	w := httptest.NewRecorder()

	h.getVideos(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetVideos_CategoriesError(t *testing.T) {
	h, storage, _ := newTestVideoHandlerWithMocks()
	videos := []admResponse.VideoResponse{{ID: 1, Name: "v1"}, {ID: 2, Name: "v2"}}
	storage.On("GetVideos", mock.Anything).Return(videos, nil)
	storage.On("GetVideoCategories", mock.Anything, int64(1)).Return([]admResponse.CategoryResponse{}, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/admin/video", nil)
	w := httptest.NewRecorder()

	h.getVideos(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateVideo_Success(t *testing.T) {
	h, storage, mockRedis := newTestVideoHandlerWithMocks()
	videoID := int64(10)
	reqBody := admResponse.VideoRequest{URL: "url", Name: "name", CategoryIDs: []int64{1, 2}}
	storage.On("UpdateVideo", mock.Anything, videoID, reqBody).Return(nil)
	storage.On("DeleteVideoCategories", mock.Anything, videoID).Return(nil)
	storage.On("AddVideoCategories", mock.Anything, videoID, []int64{1, 2}).Return(nil)
	// Ожидание для всех асинхронных вызовов GetVideo
	storage.On("GetVideo", mock.Anything, videoID).Return(admResponse.VideoResponse{ID: videoID, Categories: []admResponse.CategoryResponse{}}, nil).Maybe()
	// Ожидание для всех асинхронных вызовов DeleteVideo
	mockRedis.On("DeleteVideo", mock.Anything, "10").Return(nil).Maybe()
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/video/10", bytes.NewReader(body))
	req = muxSetURLParamVideo(req, "id", "10")
	w := httptest.NewRecorder()

	h.updateVideo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	storage.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

func TestUpdateVideo_InvalidID(t *testing.T) {
	h, _, _ := newTestVideoHandlerWithMocks()
	req := httptest.NewRequest(http.MethodPut, "/admin/video/abc", nil)
	req = muxSetURLParamVideo(req, "id", "abc")
	w := httptest.NewRecorder()

	h.updateVideo(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
