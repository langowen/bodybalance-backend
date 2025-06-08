package v1

import (
	"encoding/json"
	"errors"
	"github.com/theartofdevel/logging"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/stretchr/testify/assert"

	stor "github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/stretchr/testify/mock"
)

func newTestAccountHandler() (*Handler, *MockApiStorage, *MockRedisApi) {
	storage := &MockApiStorage{}
	redis := &MockRedisApi{}
	logger := logging.NewLogger()
	return &Handler{storage: storage, redis: redis, logger: logger}, storage, redis
}

func TestCheckAccount_FromCache(t *testing.T) {
	h, _, redis := newTestAccountHandler()
	acc := &response.AccountResponse{TypeID: 1, TypeName: "user"}
	redis.On("GetAccount", mock.Anything, "user1").Return(acc, nil)

	req := httptest.NewRequest(http.MethodGet, "/login?username=user1", nil)
	w := httptest.NewRecorder()

	h.checkAccount(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got response.AccountResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, *acc, got)
}

func TestCheckAccount_FromStorageAndCacheSet(t *testing.T) {
	h, storage, redis := newTestAccountHandler()
	redis.On("GetAccount", mock.Anything, "user2").Return(nil, nil)
	acc := response.AccountResponse{TypeID: 2, TypeName: "user2"}
	storage.On("CheckAccount", mock.Anything, "user2").Return(acc, nil)
	redis.On("SetAccount", mock.Anything, "user2", &acc, mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/login?username=user2", nil)
	w := httptest.NewRecorder()

	h.checkAccount(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got response.AccountResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, acc, got)
}

func TestCheckAccount_EmptyUsername(t *testing.T) {
	h, _, _ := newTestAccountHandler()
	req := httptest.NewRequest(http.MethodGet, "/login?username=", nil)
	w := httptest.NewRecorder()

	h.checkAccount(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCheckAccount_NotFound(t *testing.T) {
	h, storage, redis := newTestAccountHandler()
	redis.On("GetAccount", mock.Anything, "nouser").Return(nil, nil)
	storage.On("CheckAccount", mock.Anything, "nouser").Return(response.AccountResponse{}, stor.ErrAccountNotFound)

	req := httptest.NewRequest(http.MethodGet, "/login?username=nouser", nil)
	w := httptest.NewRecorder()

	h.checkAccount(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCheckAccount_StorageError(t *testing.T) {
	h, storage, redis := newTestAccountHandler()
	redis.On("GetAccount", mock.Anything, "erruser").Return(nil, nil)
	storage.On("CheckAccount", mock.Anything, "erruser").Return(response.AccountResponse{}, errors.New("fail"))

	req := httptest.NewRequest(http.MethodGet, "/login?username=erruser", nil)
	w := httptest.NewRecorder()

	h.checkAccount(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCheckAccount_RedisGetError(t *testing.T) {
	h, storage, redis := newTestAccountHandler()
	redis.On("GetAccount", mock.Anything, "user3").Return(nil, errors.New("redis fail"))
	acc := response.AccountResponse{TypeID: 3, TypeName: "user3"}
	storage.On("CheckAccount", mock.Anything, "user3").Return(acc, nil)
	redis.On("SetAccount", mock.Anything, "user3", &acc, mock.Anything).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/login?username=user3", nil)
	w := httptest.NewRecorder()

	h.checkAccount(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got response.AccountResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, acc, got)
}
