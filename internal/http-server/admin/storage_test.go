package admin

import (
	"context"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/storage/postgres/admin"
	"testing"
	"time"

	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient представляет мок Redis клиента
type MockRedisClient struct {
	mock.Mock
}

// GetCategoriesWithFallback мок метода
func (m *MockRedisClient) GetCategoriesWithFallback(ctx context.Context, contentType string, ttl time.Duration, fallback func() ([]response.CategoryResponse, error)) ([]response.CategoryResponse, error) {
	args := m.Called(ctx, contentType, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.CategoryResponse), args.Error(1)
}

// GetCategories мок метода
func (m *MockRedisClient) GetCategories(ctx context.Context, contentType string) ([]response.CategoryResponse, error) {
	args := m.Called(ctx, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.CategoryResponse), args.Error(1)
}

// SetCategories мок метода
func (m *MockRedisClient) SetCategories(ctx context.Context, contentType string, categories []response.CategoryResponse, ttl time.Duration) error {
	args := m.Called(ctx, contentType, categories, ttl)
	return args.Error(0)
}

// DeleteCategories мок метода
func (m *MockRedisClient) DeleteCategories(ctx context.Context, contentType string) error {
	args := m.Called(ctx, contentType)
	return args.Error(0)
}

// GetAccount мок метода
func (m *MockRedisClient) GetAccount(ctx context.Context, username string) (*response.AccountResponse, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AccountResponse), args.Error(1)
}

// SetAccount мок метода
func (m *MockRedisClient) SetAccount(ctx context.Context, username string, account *response.AccountResponse, ttl time.Duration) error {
	args := m.Called(ctx, username, account, ttl)
	return args.Error(0)
}

// DeleteAccount мок метода
func (m *MockRedisClient) DeleteAccount(ctx context.Context, username string) error {
	args := m.Called(ctx, username)
	return args.Error(0)
}

// GetVideo мок метода
func (m *MockRedisClient) GetVideo(ctx context.Context, videoID string) (*response.VideoResponse, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.VideoResponse), args.Error(1)
}

// SetVideo мок метода
func (m *MockRedisClient) SetVideo(ctx context.Context, videoID string, video *response.VideoResponse, ttl time.Duration) error {
	args := m.Called(ctx, videoID, video, ttl)
	return args.Error(0)
}

// DeleteVideo мок метода
func (m *MockRedisClient) DeleteVideo(ctx context.Context, videoID string) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

// GetVideosByCategoryAndType мок метода
func (m *MockRedisClient) GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]response.VideoResponse, error) {
	args := m.Called(ctx, contentType, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.VideoResponse), args.Error(1)
}

// SetVideosByCategoryAndType мок метода
func (m *MockRedisClient) SetVideosByCategoryAndType(ctx context.Context, contentType, category string, videos []response.VideoResponse, ttl time.Duration) error {
	args := m.Called(ctx, contentType, category, videos, ttl)
	return args.Error(0)
}

// DeleteVideosByCategoryAndType мок метода
func (m *MockRedisClient) DeleteVideosByCategoryAndType(ctx context.Context, contentType, category string) error {
	args := m.Called(ctx, contentType, category)
	return args.Error(0)
}

// MockAdmStorage представляет собой моки для интерфейса AdmStorage.
type MockAdmStorage struct {
	mock.Mock
}

// Реализация интерфейса AdmStorage.

// Видео
func (m *MockAdmStorage) AddVideo(ctx context.Context, video admResponse.VideoRequest) (int64, error) {
	args := m.Called(ctx, video)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAdmStorage) GetVideo(ctx context.Context, id int64) (admResponse.VideoResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(admResponse.VideoResponse), args.Error(1)
}

func (m *MockAdmStorage) GetVideos(ctx context.Context) ([]admResponse.VideoResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]admResponse.VideoResponse), args.Error(1)
}

func (m *MockAdmStorage) UpdateVideo(ctx context.Context, id int64, video admResponse.VideoRequest) error {
	args := m.Called(ctx, id, video)
	return args.Error(0)
}

func (m *MockAdmStorage) DeleteVideo(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAdmStorage) DeleteVideoCategories(ctx context.Context, videoID int64) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockAdmStorage) GetVideoCategories(ctx context.Context, videoID int64) ([]admResponse.CategoryResponse, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).([]admResponse.CategoryResponse), args.Error(1)
}

func (m *MockAdmStorage) AddVideoCategories(ctx context.Context, videoID int64, categoryIDs []int64) error {
	args := m.Called(ctx, videoID, categoryIDs)
	return args.Error(0)
}

// Администратор
func (m *MockAdmStorage) GetAdminUser(ctx context.Context, login, passwordHash string) (*admin.AdmUser, error) {
	args := m.Called(ctx, login, passwordHash)
	if args.Get(0) != nil {
		return args.Get(0).(*admin.AdmUser), args.Error(1)
	}
	return nil, args.Error(1)
}

// Типы
func (m *MockAdmStorage) AddType(ctx context.Context, req admResponse.TypeRequest) (admResponse.TypeResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(admResponse.TypeResponse), args.Error(1)
}

func (m *MockAdmStorage) GetType(ctx context.Context, id int64) (admResponse.TypeResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(admResponse.TypeResponse), args.Error(1)
}

func (m *MockAdmStorage) GetTypes(ctx context.Context) ([]admResponse.TypeResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]admResponse.TypeResponse), args.Error(1)
}

func (m *MockAdmStorage) UpdateType(ctx context.Context, id int64, req admResponse.TypeRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockAdmStorage) DeleteType(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Пользователи
func (m *MockAdmStorage) AddUser(ctx context.Context, req admResponse.UserRequest) (admResponse.UserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(admResponse.UserResponse), args.Error(1)
}

func (m *MockAdmStorage) GetUser(ctx context.Context, id int64) (admResponse.UserResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(admResponse.UserResponse), args.Error(1)
}

func (m *MockAdmStorage) GetUsers(ctx context.Context) ([]admResponse.UserResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]admResponse.UserResponse), args.Error(1)
}

func (m *MockAdmStorage) UpdateUser(ctx context.Context, id int64, req admResponse.UserRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockAdmStorage) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Категории
func (m *MockAdmStorage) AddCategory(ctx context.Context, req admResponse.CategoryRequest) (admResponse.CategoryResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(admResponse.CategoryResponse), args.Error(1)
}

func (m *MockAdmStorage) GetCategory(ctx context.Context, id int64) (admResponse.CategoryResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(admResponse.CategoryResponse), args.Error(1)
}

func (m *MockAdmStorage) GetCategories(ctx context.Context) ([]admResponse.CategoryResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]admResponse.CategoryResponse), args.Error(1)
}

func (m *MockAdmStorage) UpdateCategory(ctx context.Context, id int64, req admResponse.CategoryRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockAdmStorage) DeleteCategory(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Тесты для RedisStorage
func TestRedisStorage_CategoryMethods(t *testing.T) {
	mockRedis := &MockRedisClient{}
	ctx := context.Background()
	contentType := "1"
	ttl := time.Hour

	categories := []response.CategoryResponse{
		{ID: 1, Name: "Category 1"},
		{ID: 2, Name: "Category 2"},
	}

	// GetCategories
	mockRedis.On("GetCategories", ctx, contentType).Return(categories, nil)
	result, err := mockRedis.GetCategories(ctx, contentType)
	assert.NoError(t, err)
	assert.Equal(t, categories, result)

	// SetCategories
	mockRedis.On("SetCategories", ctx, contentType, categories, ttl).Return(nil)
	err = mockRedis.SetCategories(ctx, contentType, categories, ttl)
	assert.NoError(t, err)

	// DeleteCategories
	mockRedis.On("DeleteCategories", ctx, contentType).Return(nil)
	err = mockRedis.DeleteCategories(ctx, contentType)
	assert.NoError(t, err)

	mockRedis.AssertExpectations(t)
}

func TestRedisStorage_AccountMethods(t *testing.T) {
	mockRedis := &MockRedisClient{}
	ctx := context.Background()
	username := "testuser"
	ttl := time.Hour

	account := &response.AccountResponse{
		TypeID:   1,
		TypeName: "testuser",
	}

	// GetAccount
	mockRedis.On("GetAccount", ctx, username).Return(account, nil)
	result, err := mockRedis.GetAccount(ctx, username)
	assert.NoError(t, err)
	assert.Equal(t, account, result)

	// SetAccount
	mockRedis.On("SetAccount", ctx, username, account, ttl).Return(nil)
	err = mockRedis.SetAccount(ctx, username, account, ttl)
	assert.NoError(t, err)

	// DeleteAccount
	mockRedis.On("DeleteAccount", ctx, username).Return(nil)
	err = mockRedis.DeleteAccount(ctx, username)
	assert.NoError(t, err)

	mockRedis.AssertExpectations(t)
}

func TestRedisStorage_VideoMethods(t *testing.T) {
	mockRedis := &MockRedisClient{}
	ctx := context.Background()
	videoID := "1"
	contentType := "1"
	category := "2"
	ttl := time.Hour

	video := &response.VideoResponse{
		ID:   1,
		Name: "Test Video",
	}

	videos := []response.VideoResponse{*video}

	// GetVideo
	mockRedis.On("GetVideo", ctx, videoID).Return(video, nil)
	resultVideo, err := mockRedis.GetVideo(ctx, videoID)
	assert.NoError(t, err)
	assert.Equal(t, video, resultVideo)

	// SetVideo
	mockRedis.On("SetVideo", ctx, videoID, video, ttl).Return(nil)
	err = mockRedis.SetVideo(ctx, videoID, video, ttl)
	assert.NoError(t, err)

	// DeleteVideo
	mockRedis.On("DeleteVideo", ctx, videoID).Return(nil)
	err = mockRedis.DeleteVideo(ctx, videoID)
	assert.NoError(t, err)

	// GetVideosByCategoryAndType
	mockRedis.On("GetVideosByCategoryAndType", ctx, contentType, category).Return(videos, nil)
	resultVideos, err := mockRedis.GetVideosByCategoryAndType(ctx, contentType, category)
	assert.NoError(t, err)
	assert.Equal(t, videos, resultVideos)

	// SetVideosByCategoryAndType
	mockRedis.On("SetVideosByCategoryAndType", ctx, contentType, category, videos, ttl).Return(nil)
	err = mockRedis.SetVideosByCategoryAndType(ctx, contentType, category, videos, ttl)
	assert.NoError(t, err)

	// DeleteVideosByCategoryAndType
	mockRedis.On("DeleteVideosByCategoryAndType", ctx, contentType, category).Return(nil)
	err = mockRedis.DeleteVideosByCategoryAndType(ctx, contentType, category)
	assert.NoError(t, err)

	mockRedis.AssertExpectations(t)
}

// Старые тесты...

func TestAdmStorage_VideoMethods(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	video := admResponse.VideoRequest{Name: "Test Video"}
	videoResponse := admResponse.VideoResponse{ID: 1, Name: "Test Video"}

	mockStorage.On("AddVideo", ctx, video).Return(int64(1), nil)
	mockStorage.On("GetVideo", ctx, int64(1)).Return(videoResponse, nil)
	mockStorage.On("DeleteVideo", ctx, int64(1)).Return(nil)

	// AddVideo
	videoID, err := mockStorage.AddVideo(ctx, video)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), videoID)

	// GetVideo
	result, err := mockStorage.GetVideo(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, videoResponse, result)

	// DeleteVideo
	err = mockStorage.DeleteVideo(ctx, 1)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_AdminUserMethods(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	adminUser := &admin.AdmUser{IsAdmin: true}

	mockStorage.On("GetAdminUser", ctx, "admin", "hashed_password").Return(adminUser, nil)

	// GetAdminUser
	result, err := mockStorage.GetAdminUser(ctx, "admin", "hashed_password")
	assert.NoError(t, err)
	assert.Equal(t, adminUser, result)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_CategoryMethods(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	category := admResponse.CategoryRequest{Name: "Test Category"}
	categoryResponse := admResponse.CategoryResponse{ID: 1, Name: "Test Category"}

	mockStorage.On("AddCategory", ctx, category).Return(categoryResponse, nil)
	mockStorage.On("GetCategory", ctx, int64(1)).Return(categoryResponse, nil)

	// AddCategory
	result, err := mockStorage.AddCategory(ctx, category)
	assert.NoError(t, err)
	assert.Equal(t, categoryResponse, result)

	// GetCategory
	resultCategory, err := mockStorage.GetCategory(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, categoryResponse, resultCategory)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_GetVideos(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	videos := []admResponse.VideoResponse{
		{ID: 1, Name: "Video 1"},
		{ID: 2, Name: "Video 2"},
	}

	mockStorage.On("GetVideos", ctx).Return(videos, nil)

	result, err := mockStorage.GetVideos(ctx)
	assert.NoError(t, err)
	assert.Equal(t, videos, result)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_UpdateVideo(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	video := admResponse.VideoRequest{Name: "Updated Video"}

	mockStorage.On("UpdateVideo", ctx, int64(1), video).Return(nil)

	err := mockStorage.UpdateVideo(ctx, 1, video)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_VideoCategoriesMethods(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	categories := []admResponse.CategoryResponse{
		{ID: 1, Name: "Category 1"},
		{ID: 2, Name: "Category 2"},
	}
	categoryIDs := []int64{1, 2}

	mockStorage.On("DeleteVideoCategories", ctx, int64(1)).Return(nil)
	mockStorage.On("GetVideoCategories", ctx, int64(1)).Return(categories, nil)
	mockStorage.On("AddVideoCategories", ctx, int64(1), categoryIDs).Return(nil)

	// DeleteVideoCategories
	err := mockStorage.DeleteVideoCategories(ctx, 1)
	assert.NoError(t, err)

	// GetVideoCategories
	result, err := mockStorage.GetVideoCategories(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, categories, result)

	// AddVideoCategories
	err = mockStorage.AddVideoCategories(ctx, 1, categoryIDs)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_TypeMethods(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	typeReq := admResponse.TypeRequest{Name: "Test Type"}
	typeResponse := admResponse.TypeResponse{ID: 1, Name: "Test Type"}
	types := []admResponse.TypeResponse{typeResponse}

	// AddType
	mockStorage.On("AddType", ctx, typeReq).Return(typeResponse, nil)
	result, err := mockStorage.AddType(ctx, typeReq)
	assert.NoError(t, err)
	assert.Equal(t, typeResponse, result)

	// GetType
	mockStorage.On("GetType", ctx, int64(1)).Return(typeResponse, nil)
	resultType, err := mockStorage.GetType(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, typeResponse, resultType)

	// GetTypes
	mockStorage.On("GetTypes", ctx).Return(types, nil)
	resultTypes, err := mockStorage.GetTypes(ctx)
	assert.NoError(t, err)
	assert.Equal(t, types, resultTypes)

	// UpdateType
	updatedType := admResponse.TypeRequest{Name: "Updated Type"}
	mockStorage.On("UpdateType", ctx, int64(1), updatedType).Return(nil)
	err = mockStorage.UpdateType(ctx, 1, updatedType)
	assert.NoError(t, err)

	// DeleteType
	mockStorage.On("DeleteType", ctx, int64(1)).Return(nil)
	err = mockStorage.DeleteType(ctx, 1)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_UserMethods(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	userReq := admResponse.UserRequest{Username: "Test User"}
	userResponse := admResponse.UserResponse{ID: "1", Username: "Test User"}
	users := []admResponse.UserResponse{userResponse}

	// AddUser
	mockStorage.On("AddUser", ctx, userReq).Return(userResponse, nil)
	result, err := mockStorage.AddUser(ctx, userReq)
	assert.NoError(t, err)
	assert.Equal(t, userResponse, result)

	// GetUser
	mockStorage.On("GetUser", ctx, int64(1)).Return(userResponse, nil)
	resultUser, err := mockStorage.GetUser(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, userResponse, resultUser)

	// GetUsers
	mockStorage.On("GetUsers", ctx).Return(users, nil)
	resultUsers, err := mockStorage.GetUsers(ctx)
	assert.NoError(t, err)
	assert.Equal(t, users, resultUsers)

	// UpdateUser
	updatedUser := admResponse.UserRequest{Username: "Updated User"}
	mockStorage.On("UpdateUser", ctx, int64(1), updatedUser).Return(nil)
	err = mockStorage.UpdateUser(ctx, 1, updatedUser)
	assert.NoError(t, err)

	// DeleteUser
	mockStorage.On("DeleteUser", ctx, int64(1)).Return(nil)
	err = mockStorage.DeleteUser(ctx, 1)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_GetCategories(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	categories := []admResponse.CategoryResponse{
		{ID: 1, Name: "Category 1"},
		{ID: 2, Name: "Category 2"},
	}

	mockStorage.On("GetCategories", ctx).Return(categories, nil)

	result, err := mockStorage.GetCategories(ctx)
	assert.NoError(t, err)
	assert.Equal(t, categories, result)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_UpdateCategory(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()
	category := admResponse.CategoryRequest{Name: "Updated Category"}

	mockStorage.On("UpdateCategory", ctx, int64(1), category).Return(nil)

	err := mockStorage.UpdateCategory(ctx, 1, category)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}

func TestAdmStorage_DeleteCategory(t *testing.T) {
	mockStorage := &MockAdmStorage{}
	ctx := context.Background()

	mockStorage.On("DeleteCategory", ctx, int64(1)).Return(nil)

	err := mockStorage.DeleteCategory(ctx, 1)
	assert.NoError(t, err)

	mockStorage.AssertExpectations(t)
}
