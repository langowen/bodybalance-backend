package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupVideoTestDB вспомогательная функция для настройки тестового окружения
func setupVideoTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Storage) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	storage := &Storage{
		db: db,
	}

	return db, mock, storage
}

// TestAddVideo тестирует функцию AddVideo
func TestAddVideo(t *testing.T) {
	t.Run("successful add", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		request := admResponse.VideoRequest{
			URL:         "https://example.com/video.mp4",
			Name:        "Test Video",
			Description: "Test description",
			ImgURL:      "https://example.com/image.jpg",
		}

		// Ожидаем запрос на добавление видео
		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery(`INSERT INTO videos \(url, name, description, img_url, deleted\) VALUES \(\$1, \$2, \$3, \$4, FALSE\) RETURNING id`).
			WithArgs(request.URL, request.Name, request.Description, request.ImgURL).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		id, err := storage.AddVideo(ctx, &request)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, int64(1), id)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		request := admResponse.VideoRequest{
			URL:         "https://example.com/video.mp4",
			Name:        "Test Video",
			Description: "Test description",
			ImgURL:      "https://example.com/image.jpg",
		}

		// Ожидаем ошибку при выполнении запроса
		mock.ExpectQuery(`INSERT INTO videos \(url, name, description, img_url, deleted\) VALUES \(\$1, \$2, \$3, \$4, FALSE\) RETURNING id`).
			WithArgs(request.URL, request.Name, request.Description, request.ImgURL).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.AddVideo(ctx, &request)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "database error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestAddVideoCategories тестирует функцию AddVideoCategories
func TestAddVideoCategories(t *testing.T) {
	t.Run("successful add", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		categoryIDs := []int64{1, 2, 3}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем подготовку запроса
		mock.ExpectPrepare(`INSERT INTO video_categories`)

		// Ожидаем выполнение запросов для каждой категории
		for _, catID := range categoryIDs {
			mock.ExpectExec(`INSERT INTO video_categories`).
				WithArgs(videoID, catID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// Ожидаем успешное завершение транзакции
		mock.ExpectCommit()

		// Вызов тестируемого метода
		err := storage.AddVideoCategories(ctx, videoID, categoryIDs)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("begin transaction error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		categoryIDs := []int64{1, 2, 3}

		// Ожидаем ошибку при начале транзакции
		mock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))

		// Вызов тестируемого метода
		err := storage.AddVideoCategories(ctx, videoID, categoryIDs)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "begin transaction error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("prepare statement error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		categoryIDs := []int64{1, 2, 3}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем ошибку при подготовке запроса
		mock.ExpectPrepare(`INSERT INTO video_categories`).
			WillReturnError(errors.New("prepare statement error"))

		// Ожидаем откат транзакции
		mock.ExpectRollback()

		// Вызов тестируемого метода
		err := storage.AddVideoCategories(ctx, videoID, categoryIDs)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "prepare statement error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec statement error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		categoryIDs := []int64{1, 2, 3}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем подготовку запроса
		mock.ExpectPrepare(`INSERT INTO video_categories`)

		// Ожидаем успешное выполнение первого запроса
		mock.ExpectExec(`INSERT INTO video_categories`).
			WithArgs(videoID, categoryIDs[0]).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Ожидаем ошибку при выполнении второго запроса
		mock.ExpectExec(`INSERT INTO video_categories`).
			WithArgs(videoID, categoryIDs[1]).
			WillReturnError(errors.New("exec statement error"))

		// Ожидаем откат транзакции
		mock.ExpectRollback()

		// Вызов тестируемого метода
		err := storage.AddVideoCategories(ctx, videoID, categoryIDs)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "exec statement error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		categoryIDs := []int64{1, 2, 3}

		// Ожидаем начало транзакции
		mock.ExpectBegin()

		// Ожидаем подготовку запроса
		mock.ExpectPrepare(`INSERT INTO video_categories`)

		// Ожидаем выполнение запросов для каждой категории
		for _, catID := range categoryIDs {
			mock.ExpectExec(`INSERT INTO video_categories`).
				WithArgs(videoID, catID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		// Ожидаем ошибку при закрытии транзакции
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		// Вызов тестируемого метода
		err := storage.AddVideoCategories(ctx, videoID, categoryIDs)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "commit error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetVideo тестирует функцию GetVideo
func TestGetVideo(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		createdAt := time.Now()

		// Ожидаем запрос на получение видео
		rows := sqlmock.NewRows([]string{"id", "url", "name", "description", "img_url", "created_at"}).
			AddRow(videoID, "https://example.com/video.mp4", "Test Video", "Test description", "https://example.com/image.jpg", createdAt)

		mock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(videoID).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		video, err := storage.GetVideo(ctx, videoID)

		// Проверка результатов
		require.NoError(t, err)
		assert.Equal(t, videoID, video.ID)
		assert.Equal(t, "https://example.com/video.mp4", video.URL)
		assert.Equal(t, "Test Video", video.Name)
		assert.Equal(t, "Test description", video.Description)
		assert.Equal(t, "https://example.com/image.jpg", video.ImgURL)
		assert.Equal(t, createdAt.Format("02.01.2006"), video.DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("video not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(999)

		// Ожидаем запрос без результатов
		mock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(videoID).
			WillReturnError(sql.ErrNoRows)

		// Вызов тестируемого метода
		_, err := storage.GetVideo(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем ошибку при выполнении запроса
		mock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(videoID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetVideo(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "database error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetVideoCategories тестирует функцию GetVideoCategories
func TestGetVideoCategories(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем запрос на получение категорий видео
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Category 1").
			AddRow(2, "Category 2")

		mock.ExpectQuery(`SELECT c\.id, c\.name FROM categories c JOIN video_categories vc ON c\.id = vc\.category_id WHERE vc\.video_id = \$1`).
			WithArgs(videoID).
			WillReturnRows(rows)

		// Ожидаем запрос на получение типов контента для 1-й категории
		typesRows1 := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Type 1").
			AddRow(2, "Type 2")

		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types cct ON ct\.id = cct\.content_type_id WHERE cct\.category_id = \$1`).
			WithArgs(1).
			WillReturnRows(typesRows1)

		// Ожидаем запрос на получение типов контента для 2-й категории
		typesRows2 := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(3, "Type 3").
			AddRow(4, "Type 4")

		mock.ExpectQuery(`SELECT ct\.id, ct\.name FROM content_types ct JOIN category_content_types cct ON ct\.id = cct\.content_type_id WHERE cct\.category_id = \$1`).
			WithArgs(2).
			WillReturnRows(typesRows2)

		// Вызов тестируемого метода
		categories, err := storage.GetVideoCategories(ctx, videoID)

		// Проверка результатов
		require.NoError(t, err)
		require.NotNil(t, categories)
		catTest := *categories
		assert.Len(t, catTest, 2)
		assert.Equal(t, int64(1), catTest[0].ID)
		assert.Equal(t, "Category 1", catTest[0].Name)
		assert.Equal(t, int64(2), catTest[1].ID)
		assert.Equal(t, "Category 2", catTest[1].Name)

		// Проверка, что у категорий есть типы контента
		assert.Len(t, catTest[0].Types, 2)
		assert.Equal(t, int64(1), catTest[0].Types[0].ID)
		assert.Equal(t, "Type 1", catTest[0].Types[0].Name)

		assert.Len(t, catTest[1].Types, 2)
		assert.Equal(t, int64(3), catTest[1].Types[0].ID)
		assert.Equal(t, "Type 3", catTest[1].Types[0].Name)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем запрос без результатов
		rows := sqlmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery(`SELECT c\.id, c\.name FROM categories c JOIN video_categories vc ON c\.id = vc\.category_id WHERE vc\.video_id = \$1`).
			WithArgs(videoID).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		categories, err := storage.GetVideoCategories(ctx, videoID)

		// Проверка результатов
		require.NoError(t, err)
		assert.Empty(t, categories)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем ошибку при выполнении запроса
		mock.ExpectQuery(`SELECT c\.id, c\.name FROM categories c JOIN video_categories vc ON c\.id = vc\.category_id WHERE vc\.video_id = \$1`).
			WithArgs(videoID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetVideoCategories(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "database error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Создаем строки с некорректными данными для вызова ошибки сканирования
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(nil, "Category 1") // nil для поля id вызовет ошибку

		mock.ExpectQuery(`SELECT c\.id, c\.name FROM categories c JOIN video_categories vc ON c\.id = vc\.category_id WHERE vc\.video_id = \$1`).
			WithArgs(videoID).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		_, err := storage.GetVideoCategories(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		errMsg := strings.ToLower(err.Error())
		assert.True(t, strings.Contains(errMsg, "scan"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestGetVideos тестирует функцию GetVideos
func TestGetVideos(t *testing.T) {
	t.Run("successful get all", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		createdAt1 := time.Now()
		createdAt2 := time.Now().Add(-24 * time.Hour)

		// Ожидаем запрос на получение всех видео
		rows := sqlmock.NewRows([]string{"id", "url", "name", "description", "img_url", "created_at"}).
			AddRow(1, "https://example.com/video1.mp4", "Video 1", "Description 1", "https://example.com/image1.jpg", createdAt1).
			AddRow(2, "https://example.com/video2.mp4", "Video 2", "Description 2", "https://example.com/image2.jpg", createdAt2)

		mock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE deleted IS NOT TRUE ORDER BY id`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		videos, err := storage.GetVideos(ctx)

		// Проверка результатов
		require.NoError(t, err)
		require.NotNil(t, videos)
		videoTest := *videos
		assert.Len(t, videoTest, 2)

		// Первое видео
		assert.Equal(t, int64(1), videoTest[0].ID)
		assert.Equal(t, "https://example.com/video1.mp4", videoTest[0].URL)
		assert.Equal(t, "Video 1", videoTest[0].Name)
		assert.Equal(t, "Description 1", videoTest[0].Description)
		assert.Equal(t, "https://example.com/image1.jpg", videoTest[0].ImgURL)
		assert.Equal(t, createdAt1.Format("02.01.2006"), videoTest[0].DateCreated)

		// Второе видео
		assert.Equal(t, int64(2), videoTest[1].ID)
		assert.Equal(t, "https://example.com/video2.mp4", videoTest[1].URL)
		assert.Equal(t, "Video 2", videoTest[1].Name)
		assert.Equal(t, "Description 2", videoTest[1].Description)
		assert.Equal(t, "https://example.com/image2.jpg", videoTest[1].ImgURL)
		assert.Equal(t, createdAt2.Format("02.01.2006"), videoTest[1].DateCreated)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем запрос без результатов
		rows := sqlmock.NewRows([]string{"id", "url", "name", "description", "img_url", "created_at"})

		mock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE deleted IS NOT TRUE ORDER BY id`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		videos, err := storage.GetVideos(ctx)

		// Проверка результатов
		require.NoError(t, err)
		assert.Empty(t, videos)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Ожидаем ошибку при выполнении запроса
		mock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE deleted IS NOT TRUE ORDER BY id`).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		_, err := storage.GetVideos(ctx)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "database error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()

		// Создаем строки с некорректными данными для вызова ошибки сканирования
		rows := sqlmock.NewRows([]string{"id", "url", "name", "description", "img_url", "created_at"}).
			AddRow(nil, "https://example.com/video1.mp4", "Video 1", "Description 1", "https://example.com/image1.jpg", time.Now()) // nil для поля id вызовет ошиб

		mock.ExpectQuery(`SELECT id, url, name, description, img_url, created_at FROM videos WHERE deleted IS NOT TRUE ORDER BY id`).
			WillReturnRows(rows)

		// Вызов тестируемого метода
		_, err := storage.GetVideos(ctx)

		// Проверка результатов
		require.Error(t, err)
		errMsg := strings.ToLower(err.Error())
		assert.True(t, strings.Contains(errMsg, "scan"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestUpdateVideo тестирует функцию UpdateVideo
func TestUpdateVideo(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		request := admResponse.VideoRequest{
			URL:         "https://example.com/updated.mp4",
			Name:        "Updated Video",
			Description: "Updated description",
			ImgURL:      "https://example.com/updated-image.jpg",
		}

		// Ожидаем запрос на обновление видео
		mock.ExpectExec(`UPDATE videos SET url = \$1, name = \$2, description = \$3, img_url = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(request.URL, request.Name, request.Description, request.ImgURL, videoID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вызов тестируемого метода
		err := storage.UpdateVideo(ctx, videoID, &request)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("video not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(999)
		request := admResponse.VideoRequest{
			URL:         "https://example.com/updated.mp4",
			Name:        "Updated Video",
			Description: "Updated description",
			ImgURL:      "https://example.com/updated-image.jpg",
		}

		// Ожидаем запрос, который не затрагивает ни одной строки
		mock.ExpectExec(`UPDATE videos SET url = \$1, name = \$2, description = \$3, img_url = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(request.URL, request.Name, request.Description, request.ImgURL, videoID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Вызов тестируемого метода
		err := storage.UpdateVideo(ctx, videoID, &request)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		request := admResponse.VideoRequest{
			URL:         "https://example.com/updated.mp4",
			Name:        "Updated Video",
			Description: "Updated description",
			ImgURL:      "https://example.com/updated-image.jpg",
		}

		// Ожидаем ошибку при выполнении запроса
		mock.ExpectExec(`UPDATE videos SET url = \$1, name = \$2, description = \$3, img_url = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(request.URL, request.Name, request.Description, request.ImgURL, videoID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.UpdateVideo(ctx, videoID, &request)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "database error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)
		request := admResponse.VideoRequest{
			URL:         "https://example.com/updated.mp4",
			Name:        "Updated Video",
			Description: "Updated description",
			ImgURL:      "https://example.com/updated-image.jpg",
		}

		// Ожидаем ошибку при получении информации о затронутых строках
		result := sqlmock.NewErrorResult(errors.New("rows affected error"))
		mock.ExpectExec(`UPDATE videos SET url = \$1, name = \$2, description = \$3, img_url = \$4 WHERE id = \$5 AND deleted IS NOT TRUE`).
			WithArgs(request.URL, request.Name, request.Description, request.ImgURL, videoID).
			WillReturnResult(result)

		// Вызов тестируемого метода
		err := storage.UpdateVideo(ctx, videoID, &request)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "rows affected error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDeleteVideoCategories тестирует функцию DeleteVideoCategories
func TestDeleteVideoCategories(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем запрос на удаление категорий видео
		mock.ExpectExec(`DELETE FROM video_categories WHERE video_id = \$1`).
			WithArgs(videoID).
			WillReturnResult(sqlmock.NewResult(0, 2)) // 2 удаленных строки

		// Вызов тестируемого метода
		err := storage.DeleteVideoCategories(ctx, videoID)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no categories to delete", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем запрос, который не затрагивает ни одной строки
		mock.ExpectExec(`DELETE FROM video_categories WHERE video_id = \$1`).
			WithArgs(videoID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 удаленных строк

		// Вызов тестируемого метода
		err := storage.DeleteVideoCategories(ctx, videoID)

		// Проверка результатов (функция не должна возвращать ошибку, даже если нет строк для удаления)
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем ошибку при выполнении запроса
		mock.ExpectExec(`DELETE FROM video_categories WHERE video_id = \$1`).
			WithArgs(videoID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.DeleteVideoCategories(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "database error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestDeleteVideo тестирует функцию DeleteVideo
func TestDeleteVideo(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем запрос на пометку видео как удаленного
		mock.ExpectExec(`UPDATE videos SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(videoID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вызов тестируемого метода
		err := storage.DeleteVideo(ctx, videoID)

		// Проверка результатов
		require.NoError(t, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("video not found", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(999)

		// Ожидаем запрос, который не затрагивает ни одной строки
		mock.ExpectExec(`UPDATE videos SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(videoID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		// Вызов тестируемого метода
		err := storage.DeleteVideo(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем ошибку при выполнении запроса
		mock.ExpectExec(`UPDATE videos SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(videoID).
			WillReturnError(errors.New("database error"))

		// Вызов тестируемого метода
		err := storage.DeleteVideo(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "database error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		// Подготовка тестового окружения
		_, mock, storage := setupVideoTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Ожидаем ошибку при получении информации о затронутых строках
		result := sqlmock.NewErrorResult(errors.New("rows affected error"))
		mock.ExpectExec(`UPDATE videos SET deleted = TRUE WHERE id = \$1 AND deleted IS NOT TRUE`).
			WithArgs(videoID).
			WillReturnResult(result)

		// Вызов тестируемого метода
		err := storage.DeleteVideo(ctx, videoID)

		// Проверка результатов
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "rows affected error"))

		// Проверка, что все ожидания были выполнены
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
