package v1

import (
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
