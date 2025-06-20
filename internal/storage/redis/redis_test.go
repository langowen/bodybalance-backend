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
		contentType := int64(1)
		cacheKey := "categories:1"

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
		contentType := int64(1)
		cacheKey := "categories:1"

		expectedCategories := []response.CategoryResponse{
			{ID: 1, Name: "Test Category", ImgURL: "https://example.com/image.jpg"},
		}

		data, _ := json.Marshal(expectedCategories)

		mock.ExpectGet(cacheKey).SetVal(string(data))

		// Act
		categories, err := storage.GetCategories(ctx, contentType)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, categories)
		require.Equal(t, expectedCategories, categories) // Сравниваем сам слайс, а не указатель
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
		contentType := int64(1)
		cacheKey := "categories:1"

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
		contentType := int64(1)
		cacheKey := "categories:1"
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
		contentType := int64(1)
		cacheKey := "categories:1"
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

// TestGetAccount и TestSetAccount тестируют операции с аккаунтами
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
		videoID := int64(123)
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
		videoID := int64(123)
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
		videoID := int64(123)
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
		contentType := int64(123)
		category := int64(123)
		cacheKey := "videos:123:123"

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
		contentType := int64(123)
		category := int64(1234)
		cacheKey := "videos:123:1234"

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
		require.NotNil(t, videos)
		require.Equal(t, expectedVideos, *videos) // Сравниваем сам слайс, а не указатель
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
		contentType := int64(123)
		category := int64(1234)
		cacheKey := "videos:123:1234"
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
}

// TestInvalidateCacheByPattern тестирует удаление кэша по шаблону
func TestInvalidateCacheByPattern(t *testing.T) {
	t.Run("successfully invalidate cache by pattern", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		pattern := "videos:*"

		// Первая итерация SCAN
		mock.ExpectScan(0, pattern, 100).SetVal([]string{"videos:1:1", "videos:1:2"}, 1)
		mock.ExpectDel("videos:1:1", "videos:1:2").SetVal(2)

		// Вторая итерация SCAN
		mock.ExpectScan(1, pattern, 100).SetVal([]string{"videos:2:1"}, 0)
		mock.ExpectDel("videos:2:1").SetVal(1)

		// Act
		err := storage.InvalidateCacheByPattern(ctx, pattern)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error during scan", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		pattern := "videos:*"

		mock.ExpectScan(0, pattern, 100).SetErr(errors.New("scan error"))

		// Act
		err := storage.InvalidateCacheByPattern(ctx, pattern)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to scan redis keys")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error during delete", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		pattern := "videos:*"

		mock.ExpectScan(0, pattern, 100).SetVal([]string{"videos:1:1", "videos:1:2"}, 0)
		mock.ExpectDel("videos:1:1", "videos:1:2").SetErr(errors.New("delete error"))

		// Act
		err := storage.InvalidateCacheByPattern(ctx, pattern)

		// Assert
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to delete redis keys")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no keys found", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		pattern := "videos:*"

		mock.ExpectScan(0, pattern, 100).SetVal([]string{}, 0)

		// Act
		err := storage.InvalidateCacheByPattern(ctx, pattern)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestInvalidateAllCache тестирует удаление всего кэша
func TestInvalidateAllCache(t *testing.T) {
	t.Run("successfully invalidate all cache", func(t *testing.T) {
		// Arrange
		db, mock := redismock.NewClientMock()
		storage := &Storage{
			redis: db,
			cfg:   &config.Config{},
		}

		ctx := context.Background()
		pattern := "*"

		mock.ExpectScan(0, pattern, 100).SetVal([]string{"videos:1:1", "account:user1"}, 0)
		mock.ExpectDel("videos:1:1", "account:user1").SetVal(2)

		// Act
		err := storage.InvalidateAllCache(ctx)

		// Assert
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestInvalidateTypeSpecificCache тестирует удаление кэша определенного типа
func TestInvalidateTypeSpecificCache(t *testing.T) {
	testCases := []struct {
		name    string
		pattern string
		keys    []string
	}{
		{
			name:    "invalidate videos cache",
			pattern: "videos:*",
			keys:    []string{"videos:1:1", "videos:2:3"},
		},
		{
			name:    "invalidate categories cache",
			pattern: "categories:*",
			keys:    []string{"categories:1", "categories:2"},
		},
		{
			name:    "invalidate accounts cache",
			pattern: "account:*",
			keys:    []string{"account:user1", "account:user2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			db, mock := redismock.NewClientMock()
			storage := &Storage{
				redis: db,
				cfg:   &config.Config{},
			}

			ctx := context.Background()

			mock.ExpectScan(0, tc.pattern, 100).SetVal(tc.keys, 0)
			mock.ExpectDel(tc.keys...).SetVal(int64(len(tc.keys)))

			// Подменяем соответствующий метод на наш тестовый
			var err error
			switch tc.pattern {
			case "videos:*":
				err = storage.InvalidateVideosCache(ctx)
			case "categories:*":
				err = storage.InvalidateCategoriesCache(ctx)
			case "account:*":
				err = storage.InvalidateAccountsCache(ctx)
			}

			// Assert
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
