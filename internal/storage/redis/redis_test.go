package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew тестирует создание нового хранилища Redis
func TestNew(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()

		// Переопределяем функцию создания клиента
		originalRedisNewClient := redisNewClient
		defer func() { redisNewClient = originalRedisNewClient }()

		redisNewClient = func(options *redis.Options) redis.UniversalClient {
			return db
		}

		// Настраиваем ожидание для Ping
		mock.ExpectPing().SetVal("PONG")

		cfg := &config.Config{
			Redis: config.Redis{
				Host:     "localhost:6379",
				Password: "password",
				DB:       0,
			},
		}

		// Act
		storage, err := New(cfg)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, storage)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("connection failed", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()

		originalRedisNewClient := redisNewClient
		defer func() { redisNewClient = originalRedisNewClient }()

		redisNewClient = func(options *redis.Options) redis.UniversalClient {
			return db
		}

		mock.ExpectPing().SetErr(errors.New("connection refused"))

		cfg := &config.Config{
			Redis: config.Redis{
				Host:     "localhost:6379",
				Password: "password",
				DB:       0,
			},
		}

		// Act
		storage, err := New(cfg)

		// Assert
		require.Error(t, err)
		require.Nil(t, storage)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetSetCategories тестирует получение и установку категорий
func TestGetSetCategories(t *testing.T) {
	t.Run("get non-existent categories", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"

		mock.ExpectGet(cacheKey).SetErr(redis.Nil)

		// Act
		categories, err := storage.GetCategories(ctx, contentType)

		// Assert
		require.NoError(t, err)
		require.Nil(t, categories)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get existing categories", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"

		expectedCategories := []response.CategoryResponse{
			{ID: 1, Name: "Test Category", ImgURL: "https://example.com/image.jpg"},
		}

		data, _ := json.Marshal(expectedCategories)

		mock.ExpectGet(cacheKey).SetVal(string(data))

		// Act
		categories, err := storage.GetCategories(ctx, contentType)

		// Assert
		require.NoError(t, err)
		require.Equal(t, expectedCategories, categories)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error getting categories", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"

		mock.ExpectGet(cacheKey).SetErr(errors.New("redis error"))

		// Act
		categories, err := storage.GetCategories(ctx, contentType)

		// Assert
		require.Error(t, err)
		require.Nil(t, categories)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("set categories success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"
		ttl := 30 * time.Minute

		categories := []response.CategoryResponse{
			{ID: 1, Name: "Test Category", ImgURL: "https://example.com/image.jpg"},
		}

		data, _ := json.Marshal(categories)

		mock.ExpectSet(cacheKey, data, ttl).SetVal("OK")

		// Act
		err := storage.SetCategories(ctx, contentType, categories, ttl)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("set categories fails", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"
		ttl := 30 * time.Minute

		categories := []response.CategoryResponse{
			{ID: 1, Name: "Test Category", ImgURL: "https://example.com/image.jpg"},
		}

		data, _ := json.Marshal(categories)

		mock.ExpectSet(cacheKey, data, ttl).SetErr(errors.New("redis error"))

		// Act
		err := storage.SetCategories(ctx, contentType, categories, ttl)

		// Assert
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetCategoriesWithFallback тестирует получение категорий с фолбэком
func TestGetCategoriesWithFallback(t *testing.T) {
	t.Run("categories exist in cache", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"
		ttl := 30 * time.Minute

		expectedCategories := []response.CategoryResponse{
			{ID: 1, Name: "Test Category", ImgURL: "https://example.com/image.jpg"},
		}

		data, _ := json.Marshal(expectedCategories)

		mock.ExpectGet(cacheKey).SetVal(string(data))

		fallbackCalled := false
		fallback := func() ([]response.CategoryResponse, error) {
			fallbackCalled = true
			return nil, nil
		}

		// Act
		categories, err := storage.GetCategoriesWithFallback(ctx, contentType, ttl, fallback)

		// Assert
		require.NoError(t, err)
		require.Equal(t, expectedCategories, categories)
		assert.False(t, fallbackCalled)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("categories not in cache, fallback success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"
		ttl := 30 * time.Minute

		mock.ExpectGet(cacheKey).SetErr(redis.Nil)

		fallbackCategories := []response.CategoryResponse{
			{ID: 1, Name: "Fallback Category", ImgURL: "https://example.com/fallback.jpg"},
		}

		fallback := func() ([]response.CategoryResponse, error) {
			return fallbackCategories, nil
		}

		data, _ := json.Marshal(fallbackCategories)
		mock.ExpectSet(cacheKey, data, ttl).SetVal("OK")

		// Act
		categories, err := storage.GetCategoriesWithFallback(ctx, contentType, ttl, fallback)

		// Assert
		require.NoError(t, err)
		require.Equal(t, fallbackCategories, categories)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("categories not in cache, fallback fails", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"
		ttl := 30 * time.Minute

		mock.ExpectGet(cacheKey).SetErr(redis.Nil)

		fallback := func() ([]response.CategoryResponse, error) {
			return nil, errors.New("fallback error")
		}

		// Act
		categories, err := storage.GetCategoriesWithFallback(ctx, contentType, ttl, fallback)

		// Assert
		require.Error(t, err)
		require.Nil(t, categories)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDeleteCategories тестирует удаление категорий
func TestDeleteCategories(t *testing.T) {
	t.Run("delete categories success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"

		mock.ExpectDel(cacheKey).SetVal(1)

		// Act
		err := storage.DeleteCategories(ctx, contentType)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete categories fails", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		cacheKey := "categories:video"

		mock.ExpectDel(cacheKey).SetErr(errors.New("redis error"))

		// Act
		err := storage.DeleteCategories(ctx, contentType)

		// Assert
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestAccountOperations тестирует операции с аккаунтами
func TestAccountOperations(t *testing.T) {
	t.Run("get non-existent account", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		username := "testuser"
		cacheKey := "account:testuser"

		mock.ExpectGet(cacheKey).SetErr(redis.Nil)

		// Act
		account, err := storage.GetAccount(ctx, username)

		// Assert
		require.NoError(t, err)
		require.Nil(t, account)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get existing account", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		username := "testuser"
		cacheKey := "account:testuser"

		expectedAccount := &response.AccountResponse{
			TypeID:   1,
			TypeName: "premium",
		}

		data, _ := json.Marshal(expectedAccount)

		mock.ExpectGet(cacheKey).SetVal(string(data))

		// Act
		account, err := storage.GetAccount(ctx, username)

		// Assert
		require.NoError(t, err)
		require.Equal(t, expectedAccount, account)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("set account success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		username := "testuser"
		cacheKey := "account:testuser"
		ttl := 30 * time.Minute

		account := &response.AccountResponse{
			TypeID:   1,
			TypeName: "premium",
		}

		data, _ := json.Marshal(account)

		mock.ExpectSet(cacheKey, data, ttl).SetVal("OK")

		// Act
		err := storage.SetAccount(ctx, username, account, ttl)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete account success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		username := "testuser"
		cacheKey := "account:testuser"

		mock.ExpectDel(cacheKey).SetVal(1)

		// Act
		err := storage.DeleteAccount(ctx, username)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestVideoOperations тестирует операции с видео
func TestVideoOperations(t *testing.T) {
	t.Run("get non-existent video", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		videoID := "123"
		cacheKey := "video:123"

		mock.ExpectGet(cacheKey).SetErr(redis.Nil)

		// Act
		video, err := storage.GetVideo(ctx, videoID)

		// Assert
		require.NoError(t, err)
		require.Nil(t, video)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get existing video", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		videoID := "123"
		cacheKey := "video:123"

		expectedVideo := &response.VideoResponse{
			ID:   123,
			Name: "Test Video",
			URL:  "https://example.com/video.mp4",
		}

		data, _ := json.Marshal(expectedVideo)

		mock.ExpectGet(cacheKey).SetVal(string(data))

		// Act
		video, err := storage.GetVideo(ctx, videoID)

		// Assert
		require.NoError(t, err)
		require.Equal(t, expectedVideo, video)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("set video success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		videoID := "123"
		cacheKey := "video:123"
		ttl := 30 * time.Minute

		video := &response.VideoResponse{
			ID:   123,
			Name: "Test Video",
			URL:  "https://example.com/video.mp4",
		}

		data, _ := json.Marshal(video)

		mock.ExpectSet(cacheKey, data, ttl).SetVal("OK")

		// Act
		err := storage.SetVideo(ctx, videoID, video, ttl)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete video success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		videoID := "123"
		cacheKey := "video:123"

		mock.ExpectDel(cacheKey).SetVal(1)

		// Act
		err := storage.DeleteVideo(ctx, videoID)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestVideosByCategoryAndTypeOperations тестирует операции с видео по категории и типу
func TestVideosByCategoryAndTypeOperations(t *testing.T) {
	t.Run("get non-existent videos by category and type", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		category := "sports"
		cacheKey := "videos:video:sports"

		mock.ExpectGet(cacheKey).SetErr(redis.Nil)

		// Act
		videos, err := storage.GetVideosByCategoryAndType(ctx, contentType, category)

		// Assert
		require.NoError(t, err)
		require.Nil(t, videos)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get existing videos by category and type", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		category := "sports"
		cacheKey := "videos:video:sports"

		expectedVideos := []response.VideoResponse{
			{ID: 123, Name: "Test Video 1", URL: "https://example.com/video1.mp4"},
			{ID: 124, Name: "Test Video 2", URL: "https://example.com/video2.mp4"},
		}

		data, _ := json.Marshal(expectedVideos)

		mock.ExpectGet(cacheKey).SetVal(string(data))

		// Act
		videos, err := storage.GetVideosByCategoryAndType(ctx, contentType, category)

		// Assert
		require.NoError(t, err)
		require.Equal(t, expectedVideos, videos)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("set videos by category and type success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		category := "sports"
		cacheKey := "videos:video:sports"
		ttl := 30 * time.Minute

		videos := []response.VideoResponse{
			{ID: 123, Name: "Test Video 1", URL: "https://example.com/video1.mp4"},
			{ID: 124, Name: "Test Video 2", URL: "https://example.com/video2.mp4"},
		}

		data, _ := json.Marshal(videos)

		mock.ExpectSet(cacheKey, data, ttl).SetVal("OK")

		// Act
		err := storage.SetVideosByCategoryAndType(ctx, contentType, category, videos, ttl)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete videos by category and type success", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		contentType := "video"
		category := "sports"
		cacheKey := "videos:video:sports"

		mock.ExpectDel(cacheKey).SetVal(1)

		// Act
		err := storage.DeleteVideosByCategoryAndType(ctx, contentType, category)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
