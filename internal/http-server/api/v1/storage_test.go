package v1

import (
	"context"
	"testing"
	"time"

	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockApiStorage реализует мок для интерфейса ApiStorage
type MockApiStorage struct {
	mock.Mock
}

func (m *MockApiStorage) GetVideosByCategoryAndType(ctx context.Context, contentType, categoryName string) ([]response.VideoResponse, error) {
	args := m.Called(ctx, contentType, categoryName)
	return args.Get(0).([]response.VideoResponse), args.Error(1)
}

func (m *MockApiStorage) GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error) {
	args := m.Called(ctx, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.CategoryResponse), args.Error(1)
}

func (m *MockApiStorage) CheckAccount(ctx context.Context, username string) (response.AccountResponse, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(response.AccountResponse), args.Error(1)
}

func (m *MockApiStorage) GetVideo(ctx context.Context, videoID string) (response.VideoResponse, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(response.VideoResponse), args.Error(1)
}

type MockRedisApi struct {
	mock.Mock
}

// GetCategoriesWithFallback мок метода
func (m *MockRedisApi) GetCategoriesWithFallback(ctx context.Context, contentType string, ttl time.Duration, fallback func() ([]response.CategoryResponse, error)) ([]response.CategoryResponse, error) {
	args := m.Called(ctx, contentType, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.CategoryResponse), args.Error(1)
}

// GetCategories мок метода
func (m *MockRedisApi) GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error) {
	args := m.Called(ctx, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.CategoryResponse), args.Error(1)
}

// SetCategories мок метода
func (m *MockRedisApi) SetCategories(ctx context.Context, contentType string, categories []response.CategoryResponse, ttl time.Duration) error {
	args := m.Called(ctx, contentType, categories, ttl)
	return args.Error(0)
}

// DeleteCategories мок метода
func (m *MockRedisApi) DeleteCategories(ctx context.Context, contentType string) error {
	args := m.Called(ctx, contentType)
	return args.Error(0)
}

// GetAccount мок метода
func (m *MockRedisApi) GetAccount(ctx context.Context, username string) (*response.AccountResponse, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AccountResponse), args.Error(1)
}

// SetAccount мок метода
func (m *MockRedisApi) SetAccount(ctx context.Context, username string, account *response.AccountResponse, ttl time.Duration) error {
	args := m.Called(ctx, username, account, ttl)
	return args.Error(0)
}

// DeleteAccount мок метода
func (m *MockRedisApi) DeleteAccount(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

// GetVideo мок метода
func (m *MockRedisApi) GetVideo(ctx context.Context, videoID string) (*response.VideoResponse, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.VideoResponse), args.Error(1)
}

// SetVideo мок метода
func (m *MockRedisApi) SetVideo(ctx context.Context, videoID string, video *response.VideoResponse, ttl time.Duration) error {
	args := m.Called(ctx, videoID, video, ttl)
	return args.Error(0)
}

// DeleteVideo мок метода
func (m *MockRedisApi) DeleteVideo(ctx context.Context, videoID string) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

// GetVideosByCategoryAndType мок метода
func (m *MockRedisApi) GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]response.VideoResponse, error) {
	args := m.Called(ctx, contentType, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.VideoResponse), args.Error(1)
}

// SetVideosByCategoryAndType мок метода
func (m *MockRedisApi) SetVideosByCategoryAndType(ctx context.Context, contentType, category string, videos []response.VideoResponse, ttl time.Duration) error {
	args := m.Called(ctx, contentType, category, videos, ttl)
	return args.Error(0)
}

// DeleteVideosByCategoryAndType мок метода
func (m *MockRedisApi) DeleteVideosByCategoryAndType(ctx context.Context, contentType, category string) error {
	args := m.Called(ctx, contentType, category)
	return args.Error(0)
}

// Пример тестов для ApiStorage
func TestApiStorage_GetVideosByCategoryAndType(t *testing.T) {
	mockStorage := &MockApiStorage{}
	ctx := context.Background()
	contentType := "type1"
	category := "cat1"
	videos := []response.VideoResponse{{ID: 1, Name: "Video1"}}
	mockStorage.On("GetVideosByCategoryAndType", ctx, contentType, category).Return(videos, nil)

	result, err := mockStorage.GetVideosByCategoryAndType(ctx, contentType, category)
	assert.NoError(t, err)
	assert.Equal(t, videos, result)
	mockStorage.AssertExpectations(t)
}

func TestApiStorage_GetCategories(t *testing.T) {
	mockStorage := &MockApiStorage{}
	ctx := context.Background()
	contentType := "type1"
	categories := []response.CategoryResponse{{ID: 1, Name: "Cat1"}}
	mockStorage.On("GetCategories", ctx, contentType).Return(categories, nil)

	result, err := mockStorage.GetCategories(ctx, contentType)
	assert.NoError(t, err)
	assert.Equal(t, categories, result)
	mockStorage.AssertExpectations(t)
}

func TestApiStorage_CheckAccount(t *testing.T) {
	mockStorage := &MockApiStorage{}
	ctx := context.Background()
	username := "user1"
	account := response.AccountResponse{TypeID: 1, TypeName: "user1"}
	mockStorage.On("CheckAccount", ctx, username).Return(account, nil)

	result, err := mockStorage.CheckAccount(ctx, username)
	assert.NoError(t, err)
	assert.Equal(t, account, result)
	mockStorage.AssertExpectations(t)
}

func TestApiStorage_GetVideo(t *testing.T) {
	mockStorage := &MockApiStorage{}
	ctx := context.Background()
	videoID := "1"
	video := response.VideoResponse{ID: 1, Name: "Video1"}
	mockStorage.On("GetVideo", ctx, videoID).Return(video, nil)

	result, err := mockStorage.GetVideo(ctx, videoID)
	assert.NoError(t, err)
	assert.Equal(t, video, result)
	mockStorage.AssertExpectations(t)
}

// Пример тестов для RedisApi
func TestRedisApi_GetCategoriesWithFallback(t *testing.T) {
	mockRedis := &MockRedisApi{}
	ctx := context.Background()
	contentType := "type1"
	ttl := time.Minute
	categories := []response.CategoryResponse{{ID: 1, Name: "Cat1"}}
	mockRedis.On("GetCategoriesWithFallback", ctx, contentType, ttl).Return(categories, nil)

	result, err := mockRedis.GetCategoriesWithFallback(ctx, contentType, ttl, nil)
	assert.NoError(t, err)
	assert.Equal(t, categories, result)
	mockRedis.AssertExpectations(t)
}

func TestRedisApi_GetCategories(t *testing.T) {
	mockRedis := &MockRedisApi{}
	ctx := context.Background()
	contentType := "type1"
	categories := []response.CategoryResponse{{ID: 1, Name: "Cat1"}}
	mockRedis.On("GetCategories", ctx, contentType).Return(categories, nil)

	result, err := mockRedis.GetCategories(ctx, contentType)
	assert.NoError(t, err)
	assert.Equal(t, categories, result)
	mockRedis.AssertExpectations(t)
}

func TestRedisApi_SetCategories(t *testing.T) {
	mockRedis := &MockRedisApi{}
	ctx := context.Background()
	contentType := "type1"
	categories := []response.CategoryResponse{{ID: 1, Name: "Cat1"}}
	ttl := time.Minute
	mockRedis.On("SetCategories", ctx, contentType, categories, ttl).Return(nil)

	err := mockRedis.SetCategories(ctx, contentType, categories, ttl)
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}

func TestRedisApi_DeleteCategories(t *testing.T) {
	mockRedis := &MockRedisApi{}
	ctx := context.Background()
	contentType := "type1"
	mockRedis.On("DeleteCategories", ctx, contentType).Return(nil)

	err := mockRedis.DeleteCategories(ctx, contentType)
	assert.NoError(t, err)
	mockRedis.AssertExpectations(t)
}
