package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theartofdevel/logging"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestHandlerWithMockStorage() (*Handler, *MockAdmStorage) {
	mockStorage := &MockAdmStorage{}
	logger := logging.NewLogger()
	cfg := &config.Config{}
	return New(logger, mockStorage, cfg, nil), mockStorage
}

func TestAddType_Success(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	reqBody := admResponse.TypeRequest{Name: "TestType"}
	mockStorage.On("AddType", mock.Anything, reqBody).Return(admResponse.TypeResponse{ID: 1, Name: "TestType"}, nil)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addType(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestAddType_InvalidJSON(t *testing.T) {
	h, _ := newTestHandlerWithMockStorage()
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewReader([]byte("{")))
	w := httptest.NewRecorder()

	h.addType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddType_EmptyName(t *testing.T) {
	h, _ := newTestHandlerWithMockStorage()
	reqBody := admResponse.TypeRequest{Name: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddType_StorageError(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	reqBody := admResponse.TypeRequest{Name: "TestType"}
	mockStorage.On("AddType", mock.Anything, reqBody).Return(admResponse.TypeResponse{}, errors.New("fail"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/type", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addType(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetType_Success(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	resp := admResponse.TypeResponse{ID: 1, Name: "TestType"}
	mockStorage.On("GetType", mock.Anything, int64(1)).Return(resp, nil)
	req := httptest.NewRequest(http.MethodGet, "/admin/type/1", nil)
	req = muxSetURLParam(req, "id", "1")
	w := httptest.NewRecorder()

	h.getType(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestGetType_InvalidID(t *testing.T) {
	h, _ := newTestHandlerWithMockStorage()
	req := httptest.NewRequest(http.MethodGet, "/admin/type/abc", nil)
	req = muxSetURLParam(req, "id", "abc")
	w := httptest.NewRecorder()

	h.getType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetType_NotFound(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	mockStorage.On("GetType", mock.Anything, int64(2)).Return(admResponse.TypeResponse{}, sql.ErrNoRows)
	req := httptest.NewRequest(http.MethodGet, "/admin/type/2", nil)
	req = muxSetURLParam(req, "id", "2")
	w := httptest.NewRecorder()

	h.getType(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetType_StorageError(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	mockStorage.On("GetType", mock.Anything, int64(3)).Return(admResponse.TypeResponse{}, errors.New("fail"))
	req := httptest.NewRequest(http.MethodGet, "/admin/type/3", nil)
	req = muxSetURLParam(req, "id", "3")
	w := httptest.NewRecorder()

	h.getType(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateType_Success(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	reqBody := admResponse.TypeRequest{Name: "UpdatedType"}
	mockStorage.On("UpdateType", mock.Anything, int64(1), reqBody).Return(nil)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/type/1", bytes.NewReader(body))
	req = muxSetURLParam(req, "id", "1")
	w := httptest.NewRecorder()

	h.updateType(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateType_InvalidID(t *testing.T) {
	h, _ := newTestHandlerWithMockStorage()
	req := httptest.NewRequest(http.MethodPut, "/admin/type/abc", nil)
	req = muxSetURLParam(req, "id", "abc")
	w := httptest.NewRecorder()

	h.updateType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateType_DecodeError(t *testing.T) {
	h, _ := newTestHandlerWithMockStorage()
	req := httptest.NewRequest(http.MethodPut, "/admin/type/1", bytes.NewReader([]byte("{")))
	req = muxSetURLParam(req, "id", "1")
	w := httptest.NewRecorder()

	h.updateType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateType_NotFound(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	reqBody := admResponse.TypeRequest{Name: "UpdatedType"}
	mockStorage.On("UpdateType", mock.Anything, int64(2), reqBody).Return(sql.ErrNoRows)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/type/2", bytes.NewReader(body))
	req = muxSetURLParam(req, "id", "2")
	w := httptest.NewRecorder()

	h.updateType(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateType_StorageError(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	reqBody := admResponse.TypeRequest{Name: "UpdatedType"}
	mockStorage.On("UpdateType", mock.Anything, int64(3), reqBody).Return(errors.New("fail"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/type/3", bytes.NewReader(body))
	req = muxSetURLParam(req, "id", "3")
	w := httptest.NewRecorder()

	h.updateType(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteType_Success(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	mockStorage.On("DeleteType", mock.Anything, int64(1)).Return(nil)
	req := httptest.NewRequest(http.MethodDelete, "/admin/type/1", nil)
	req = muxSetURLParam(req, "id", "1")
	w := httptest.NewRecorder()

	h.deleteType(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteType_InvalidID(t *testing.T) {
	h, _ := newTestHandlerWithMockStorage()
	req := httptest.NewRequest(http.MethodDelete, "/admin/type/abc", nil)
	req = muxSetURLParam(req, "id", "abc")
	w := httptest.NewRecorder()

	h.deleteType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteType_NotFound(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	mockStorage.On("DeleteType", mock.Anything, int64(2)).Return(sql.ErrNoRows)
	req := httptest.NewRequest(http.MethodDelete, "/admin/type/2", nil)
	req = muxSetURLParam(req, "id", "2")
	w := httptest.NewRecorder()

	h.deleteType(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteType_StorageError(t *testing.T) {
	h, mockStorage := newTestHandlerWithMockStorage()
	mockStorage.On("DeleteType", mock.Anything, int64(3)).Return(errors.New("fail"))
	req := httptest.NewRequest(http.MethodDelete, "/admin/type/3", nil)
	req = muxSetURLParam(req, "id", "3")
	w := httptest.NewRecorder()

	h.deleteType(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// muxSetURLParam позволяет подменить chi URLParam для тестов
func muxSetURLParam(r *http.Request, key, val string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}
