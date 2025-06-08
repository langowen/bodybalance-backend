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

func newTestCategoryHandler() (*Handler, *MockApiStorage, *MockRedisApi) {
	storage := &MockApiStorage{}
	redis := &MockRedisApi{}
	logger := logging.NewLogger()
	return &Handler{storage: storage, redis: redis, logger: logger}, storage, redis
}

func TestGetCategoriesByType_EmptyType(t *testing.T) {
	h, _, _ := newTestCategoryHandler()
	req := httptest.NewRequest(http.MethodGet, "/category?type=", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCategoriesByType_FromCache(t *testing.T) {
	h, _, redis := newTestCategoryHandler()
	cats := []response.CategoryResponse{{ID: 1, Name: "cat1", ImgURL: "img1"}}
	redis.On("GetCategories", mock.Anything, "1").Return(cats, nil)

	req := httptest.NewRequest(http.MethodGet, "/category?type=1", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got []response.CategoryResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, cats, got)
}

func TestGetCategoriesByType_FromStorageAndCacheSet(t *testing.T) {
	h, storage, redis := newTestCategoryHandler()
	redis.On("GetCategories", mock.Anything, "2").Return(nil, nil)
	cats := []response.CategoryResponse{{ID: 2, Name: "cat2", ImgURL: "img2"}}
	storage.On("GetCategories", mock.Anything, "2").Return(cats, nil)
	redis.On("SetCategories", mock.Anything, "2", cats, mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/category?type=2", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got []response.CategoryResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, cats, got)
}

func TestGetCategoriesByType_ContentTypeNotFound(t *testing.T) {
	h, storage, redis := newTestCategoryHandler()
	redis.On("GetCategories", mock.Anything, "404").Return(nil, nil)
	storage.On("GetCategories", mock.Anything, "404").Return(nil, st.ErrContentTypeNotFound)

	req := httptest.NewRequest(http.MethodGet, "/category?type=404", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetCategoriesByType_NoCategoriesFound(t *testing.T) {
	h, storage, redis := newTestCategoryHandler()
	redis.On("GetCategories", mock.Anything, "404cat").Return(nil, nil)
	storage.On("GetCategories", mock.Anything, "404cat").Return(nil, st.ErrNoCategoriesFound)

	req := httptest.NewRequest(http.MethodGet, "/category?type=404cat", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetCategoriesByType_StorageError(t *testing.T) {
	h, storage, redis := newTestCategoryHandler()
	redis.On("GetCategories", mock.Anything, "err").Return(nil, nil)
	storage.On("GetCategories", mock.Anything, "err").Return(nil, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/category?type=err", nil)
	w := httptest.NewRecorder()

	h.getCategoriesByType(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
