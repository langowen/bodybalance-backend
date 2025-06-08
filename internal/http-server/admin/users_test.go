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

func newTestHandlerWithMockUserStorage() (*Handler, *MockAdmStorage, *MockRedisClient) {
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	logger := logging.NewLogger()
	cfg := &config.Config{}
	return New(logger, mockStorage, cfg, mockRedis), mockStorage, mockRedis
}

func muxSetURLParamUser(r *http.Request, key, val string) *http.Request {
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

func TestAddUser_Success(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	reqBody := admResponse.UserRequest{Username: "user1", Password: "pass"}
	userResp := admResponse.UserResponse{ID: "1", Username: "user1"}
	mockStorage.On("AddUser", mock.Anything, reqBody).Return(userResp, nil)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addUser(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestAddUser_InvalidJSON(t *testing.T) {
	h, _, _ := newTestHandlerWithMockUserStorage()
	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader([]byte("{")))
	w := httptest.NewRecorder()

	h.addUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddUser_EmptyFields(t *testing.T) {
	h, _, _ := newTestHandlerWithMockUserStorage()
	reqBody := admResponse.UserRequest{Username: "", Password: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddUser_Duplicate(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	reqBody := admResponse.UserRequest{Username: "user1", Password: "pass"}
	mockStorage.On("AddUser", mock.Anything, reqBody).Return(admResponse.UserResponse{}, errors.New("user already exists"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addUser(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestAddUser_StorageError(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	reqBody := admResponse.UserRequest{Username: "user1", Password: "pass"}
	mockStorage.On("AddUser", mock.Anything, reqBody).Return(admResponse.UserResponse{}, errors.New("fail"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/admin/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.addUser(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUser_Success(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	userResp := admResponse.UserResponse{ID: "1", Username: "user1"}
	mockStorage.On("GetUser", mock.Anything, int64(1)).Return(userResp, nil)
	req := httptest.NewRequest(http.MethodGet, "/admin/users/1", nil)
	req = muxSetURLParamUser(req, "id", "1")
	w := httptest.NewRecorder()

	h.getUser(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	mockStorage.AssertExpectations(t)
}

func TestGetUser_InvalidID(t *testing.T) {
	h, _, _ := newTestHandlerWithMockUserStorage()
	req := httptest.NewRequest(http.MethodGet, "/admin/users/abc", nil)
	req = muxSetURLParamUser(req, "id", "abc")
	w := httptest.NewRecorder()

	h.getUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUser_NotFound(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	mockStorage.On("GetUser", mock.Anything, int64(2)).Return(admResponse.UserResponse{}, sql.ErrNoRows)
	req := httptest.NewRequest(http.MethodGet, "/admin/users/2", nil)
	req = muxSetURLParamUser(req, "id", "2")
	w := httptest.NewRecorder()

	h.getUser(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUser_StorageError(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	mockStorage.On("GetUser", mock.Anything, int64(3)).Return(admResponse.UserResponse{}, errors.New("fail"))
	req := httptest.NewRequest(http.MethodGet, "/admin/users/3", nil)
	req = muxSetURLParamUser(req, "id", "3")
	w := httptest.NewRecorder()

	h.getUser(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateUser_Success(t *testing.T) {
	h, mockStorage, mockRedis := newTestHandlerWithMockUserStorage()
	reqBody := admResponse.UserRequest{Username: "user1"}
	mockStorage.On("UpdateUser", mock.Anything, int64(1), reqBody).Return(nil)
	mockRedis.On("DeleteAccount", mock.Anything, reqBody.Username).Return(nil)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/users/1", bytes.NewReader(body))
	req = muxSetURLParamUser(req, "id", "1")
	w := httptest.NewRecorder()
	done := make(chan struct{}, 1)
	req = req.WithContext(context.WithValue(req.Context(), "done", done))

	h.updateUser(w, req)
	<-done
	assert.Equal(t, http.StatusOK, w.Code)
	mockRedis.AssertExpectations(t)
}

func TestUpdateUser_InvalidID(t *testing.T) {
	h, _, _ := newTestHandlerWithMockUserStorage()
	req := httptest.NewRequest(http.MethodPut, "/admin/users/abc", nil)
	req = muxSetURLParamUser(req, "id", "abc")
	w := httptest.NewRecorder()

	h.updateUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUser_DecodeError(t *testing.T) {
	h, _, _ := newTestHandlerWithMockUserStorage()
	req := httptest.NewRequest(http.MethodPut, "/admin/users/1", bytes.NewReader([]byte("{")))
	req = muxSetURLParamUser(req, "id", "1")
	w := httptest.NewRecorder()

	h.updateUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUser_NotFound(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	reqBody := admResponse.UserRequest{Username: "user1"}
	mockStorage.On("UpdateUser", mock.Anything, int64(2), reqBody).Return(sql.ErrNoRows)
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/users/2", bytes.NewReader(body))
	req = muxSetURLParamUser(req, "id", "2")
	w := httptest.NewRecorder()
	done := make(chan struct{}, 1)
	req = req.WithContext(context.WithValue(req.Context(), "done", done))

	h.updateUser(w, req)
	<-done
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateUser_StorageError(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	reqBody := admResponse.UserRequest{Username: "user1"}
	mockStorage.On("UpdateUser", mock.Anything, int64(3), reqBody).Return(errors.New("fail"))
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/admin/users/3", bytes.NewReader(body))
	req = muxSetURLParamUser(req, "id", "3")
	w := httptest.NewRecorder()
	done := make(chan struct{}, 1)
	req = req.WithContext(context.WithValue(req.Context(), "done", done))

	h.updateUser(w, req)
	<-done
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteUser_Success(t *testing.T) {
	h, mockStorage, mockRedis := newTestHandlerWithMockUserStorage()
	userResp := admResponse.UserResponse{ID: "1", Username: "user1"}
	mockStorage.On("GetUser", mock.Anything, int64(1)).Return(userResp, nil)
	mockStorage.On("DeleteUser", mock.Anything, int64(1)).Return(nil)
	mockRedis.On("DeleteAccount", mock.Anything, userResp.Username).Return(nil)
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	req = muxSetURLParamUser(req, "id", "1")
	w := httptest.NewRecorder()
	done := make(chan struct{}, 1)
	req = req.WithContext(context.WithValue(req.Context(), "done", done))

	h.deleteUser(w, req)
	<-done
	assert.Equal(t, http.StatusOK, w.Code)
	mockRedis.AssertExpectations(t)
}

func TestDeleteUser_InvalidID(t *testing.T) {
	h, _, _ := newTestHandlerWithMockUserStorage()
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/abc", nil)
	req = muxSetURLParamUser(req, "id", "abc")
	w := httptest.NewRecorder()

	h.deleteUser(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUser_NotFound(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	mockStorage.On("GetUser", mock.Anything, int64(2)).Return(admResponse.UserResponse{}, sql.ErrNoRows)
	mockStorage.On("DeleteUser", mock.Anything, int64(2)).Return(sql.ErrNoRows)
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/2", nil)
	req = muxSetURLParamUser(req, "id", "2")
	w := httptest.NewRecorder()
	done := make(chan struct{}, 1)
	req = req.WithContext(context.WithValue(req.Context(), "done", done))

	h.deleteUser(w, req)
	<-done
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteUser_StorageError(t *testing.T) {
	h, mockStorage, _ := newTestHandlerWithMockUserStorage()
	userResp := admResponse.UserResponse{ID: "3", Username: "user3"}
	mockStorage.On("GetUser", mock.Anything, int64(3)).Return(userResp, nil)
	mockStorage.On("DeleteUser", mock.Anything, int64(3)).Return(errors.New("fail"))
	req := httptest.NewRequest(http.MethodDelete, "/admin/users/3", nil)
	req = muxSetURLParamUser(req, "id", "3")
	w := httptest.NewRecorder()
	done := make(chan struct{}, 1)
	req = req.WithContext(context.WithValue(req.Context(), "done", done))

	h.deleteUser(w, req)
	<-done
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
