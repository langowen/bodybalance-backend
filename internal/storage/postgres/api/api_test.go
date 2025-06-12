package api

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/langowen/bodybalance-backend/internal/config"
	st "github.com/langowen/bodybalance-backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Вспомогательная функция для настройки тестового окружения
func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Storage) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	// Создаем конфигурацию для тестов
	cfg := &config.Config{
		Media: config.Media{
			BaseURL: "http://localhost:8080",
		},
	}

	// Создаем экземпляр хранилища с моком DB
	storage := &Storage{
		db:  db,
		cfg: cfg,
	}

	return db, mock, storage
}

// Тест для функции GetVideosByCategoryAndType
func TestGetVideosByCategoryAndType(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)
		category := int64(2)

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Мокаем проверку категории
		categoryExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(category).
			WillReturnRows(categoryExists)

		// Мокаем основной запрос
		rows := sqlmock.NewRows([]string{"id", "url", "name", "description", "category", "img_url"}).
			AddRow(1, "video1.mp4", "Video 1", "Description 1", "Category 1", "img1.jpg").
			AddRow(2, "video2.mp4", "Video 2", "Description 2", "Category 2", "img2.jpg")

		mock.ExpectQuery(`SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url.*ORDER BY`).
			WithArgs(contentType, category).
			WillReturnRows(rows)

		// Вызываем тестируемую функцию
		videos, err := storage.GetVideosByCategoryAndType(ctx, contentType, category)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Len(t, *videos, 2)
		videosList := *videos
		assert.Equal(t, int64(1), videosList[0].ID)
		assert.Equal(t, "http://localhost:8080/video/video1.mp4", videosList[0].URL)
		assert.Equal(t, "Video 1", videosList[0].Name)
		assert.Equal(t, "Description 1", videosList[0].Description)
		assert.Equal(t, "Category 1", videosList[0].Category)
		assert.Equal(t, "http://localhost:8080/img/img1.jpg", videosList[0].ImgURL)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("content type not found", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(999)
		category := int64(1)

		// Мокаем проверку типа контента - тип не существует
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Вызываем тестируемую функцию
		videos, err := storage.GetVideosByCategoryAndType(ctx, contentType, category)

		// Проверяем результат
		assert.Error(t, err)
		assert.Nil(t, videos)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrContentTypeNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("category not found", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)
		category := int64(999)

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Мокаем проверку категории - категория не существует
		categoryExists := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(category).
			WillReturnRows(categoryExists)

		// Вызываем тестируемую функцию
		videos, err := storage.GetVideosByCategoryAndType(ctx, contentType, category)

		// Проверяем результат
		assert.Error(t, err)
		assert.Nil(t, videos)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrNoCategoriesFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no videos found", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)
		category := int64(2)

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Мокаем проверку категории
		categoryExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(category).
			WillReturnRows(categoryExists)

		// Мокаем основной запрос - пустой результат
		emptyRows := sqlmock.NewRows([]string{"id", "url", "name", "description", "category", "img_url"})

		mock.ExpectQuery(`SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url.*ORDER BY`).
			WithArgs(contentType, category).
			WillReturnRows(emptyRows)

		// Вызываем тестируемую функцию
		videos, err := storage.GetVideosByCategoryAndType(ctx, contentType, category)

		// Проверяем результат
		assert.Error(t, err)
		assert.Nil(t, videos)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrVideoNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)
		category := int64(2)

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Мокаем проверку категории
		categoryExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(category).
			WillReturnRows(categoryExists)

		// Мокаем основной запрос - возвращаем ошибку
		dbErr := errors.New("database error")
		mock.ExpectQuery(`SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url.*ORDER BY`).
			WithArgs(contentType, category).
			WillReturnError(dbErr)

		// Вызываем тестируемую функцию
		videos, err := storage.GetVideosByCategoryAndType(ctx, contentType, category)

		// Проверяем результат
		assert.Error(t, err)
		assert.Nil(t, videos)
		assert.ErrorContains(t, err, "query failed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Тест для функции CheckAccount
func TestCheckAccount(t *testing.T) {
	t.Run("successful check", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		username := "testuser"

		// Мокаем запрос
		rows := sqlmock.NewRows([]string{"content_type_id", "name"}).
			AddRow(1, "premium")

		mock.ExpectQuery(`SELECT a.content_type_id, ct.name.*FROM accounts a.*WHERE a.username = \$1`).
			WithArgs(username).
			WillReturnRows(rows)

		// Вызываем тестируемую функцию
		account, err := storage.CheckAccount(ctx, username)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, int64(1), account.TypeID)
		assert.Equal(t, "premium", account.TypeName)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("account not found", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		username := "nonexistentuser"

		// Мокаем запрос - аккаунт не найден
		mock.ExpectQuery(`SELECT a.content_type_id, ct.name.*FROM accounts a.*WHERE a.username = \$1`).
			WithArgs(username).
			WillReturnError(sql.ErrNoRows)

		// Вызываем тестируемую функцию
		account, err := storage.CheckAccount(ctx, username)

		// Проверяем результат
		assert.Error(t, err)
		assert.Zero(t, account)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrAccountNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		username := "testuser"

		// Мокаем запрос - возвращаем ошибку
		dbErr := errors.New("database error")
		mock.ExpectQuery(`SELECT a.content_type_id, ct.name.*FROM accounts a.*WHERE a.username = \$1`).
			WithArgs(username).
			WillReturnError(dbErr)

		// Вызываем тестируемую функцию
		account, err := storage.CheckAccount(ctx, username)

		// Проверяем результат
		assert.Error(t, err)
		assert.Zero(t, account)
		assert.ErrorContains(t, err, "query failed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Тест для функции GetCategories
func TestGetCategories(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Мокаем основной запрос
		rows := sqlmock.NewRows([]string{"id", "name", "img_url"}).
			AddRow(1, "Category 1", "cat1.jpg").
			AddRow(2, "Category 2", "cat2.jpg")

		mock.ExpectQuery(`SELECT c.id, c.name, c.img_url.*FROM categories c.*ORDER BY`).
			WithArgs(contentType).
			WillReturnRows(rows)

		// Вызываем тестируемую функцию
		categories, err := storage.GetCategories(ctx, contentType)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NotNil(t, categories)
		categoriesList := *categories
		assert.Len(t, categoriesList, 2)
		assert.Equal(t, int64(1), categoriesList[0].ID)
		assert.Equal(t, "Category 1", categoriesList[0].Name)
		assert.Equal(t, "http://localhost:8080/img/cat1.jpg", categoriesList[0].ImgURL)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("content type not found", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(999)

		// Мокаем проверку типа контента - тип не существует
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Вызываем тестируемую функцию
		categories, err := storage.GetCategories(ctx, contentType)

		// Проверяем результат
		assert.Error(t, err)
		assert.Nil(t, categories)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrContentTypeNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no categories found", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Мокаем основной запрос - пустой результат
		emptyRows := sqlmock.NewRows([]string{"id", "name", "img_url"})

		mock.ExpectQuery(`SELECT c.id, c.name, c.img_url.*FROM categories c.*ORDER BY`).
			WithArgs(contentType).
			WillReturnRows(emptyRows)

		// Вызываем тестируемую функцию
		categories, err := storage.GetCategories(ctx, contentType)

		// Проверяем результат
		assert.Error(t, err)
		assert.Nil(t, categories)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrNoCategoriesFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Мокаем основной запрос - возвращаем ошибку
		dbErr := errors.New("database error")
		mock.ExpectQuery(`SELECT c.id, c.name, c.img_url.*FROM categories c.*ORDER BY`).
			WithArgs(contentType).
			WillReturnError(dbErr)

		// Вызываем тестируемую функцию
		categories, err := storage.GetCategories(ctx, contentType)

		// Проверяем результат
		assert.Error(t, err)
		assert.Nil(t, categories)
		assert.ErrorContains(t, err, "query failed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Тест для функции GetVideo
func TestGetVideo(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Мокаем запрос
		rows := sqlmock.NewRows([]string{"id", "url", "name", "description", "category", "img_url"}).
			AddRow(1, "video1.mp4", "Video 1", "Description 1", "Category 1", "img1.jpg")

		mock.ExpectQuery(`SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url.*FROM videos v.*WHERE v.id = \$1`).
			WithArgs(videoID).
			WillReturnRows(rows)

		// Вызываем тестируемую функцию
		video, err := storage.GetVideo(ctx, videoID)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Equal(t, int64(1), video.ID)
		assert.Equal(t, "http://localhost:8080/video/video1.mp4", video.URL)
		assert.Equal(t, "Video 1", video.Name)
		assert.Equal(t, "Description 1", video.Description)
		assert.Equal(t, "Category 1", video.Category)
		assert.Equal(t, "http://localhost:8080/img/img1.jpg", video.ImgURL)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("video not found", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(999)

		// Мокаем запрос - видео не найдено
		mock.ExpectQuery(`SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url.*FROM videos v.*WHERE v.id = \$1`).
			WithArgs(videoID).
			WillReturnError(sql.ErrNoRows)

		// Вызываем тестируемую функцию
		video, err := storage.GetVideo(ctx, videoID)

		// Проверяем результат
		assert.Error(t, err)
		assert.Zero(t, video)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrVideoNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		videoID := int64(1)

		// Мокаем запрос - возвращаем ошибку
		dbErr := errors.New("database error")
		mock.ExpectQuery(`SELECT v.id, v.url, v.name, v.description, c.name as category, v.img_url.*FROM videos v.*WHERE v.id = \$1`).
			WithArgs(videoID).
			WillReturnError(dbErr)

		// Вызываем тестируемую функцию
		video, err := storage.GetVideo(ctx, videoID)

		// Проверяем результат
		assert.Error(t, err)
		assert.Zero(t, video)
		assert.ErrorContains(t, err, "query failed")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Тесты для вспомогательных функций
func TestHelperFunctions(t *testing.T) {
	t.Run("constructFullMediaURL", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, _, storage := setupTestDB(t)

		// Тестовые случаи
		testCases := []struct {
			relativePath string
			expected     string
		}{
			{"video.mp4", "http://localhost:8080/video/video.mp4"},
			{"/video.mp4", "http://localhost:8080/video/video.mp4"},
			{"", ""},
		}

		for _, tc := range testCases {
			result := storage.constructFullMediaURL(tc.relativePath)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("constructFullImgURL", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, _, storage := setupTestDB(t)

		// Тестовые случаи
		testCases := []struct {
			relativePath string
			expected     string
		}{
			{"image.jpg", "http://localhost:8080/img/image.jpg"},
			{"/image.jpg", "http://localhost:8080/img/image.jpg"},
			{"", ""},
		}

		for _, tc := range testCases {
			result := storage.constructFullImgURL(tc.relativePath)
			assert.Equal(t, tc.expected, result)
		}
	})
}

// Тесты для проверки типа контента и категории
func TestCheckFunctions(t *testing.T) {
	t.Run("chekType success", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(1)
		op := "test"

		// Мокаем проверку типа контента
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Вызываем тестируемую функцию
		err := storage.chekType(ctx, contentType, op)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("chekType failure", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		contentType := int64(999)
		op := "test"

		// Мокаем проверку типа контента - тип не существует
		contentTypeExists := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM content_types WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(contentType).
			WillReturnRows(contentTypeExists)

		// Вызываем тестируемую функцию
		err := storage.chekType(ctx, contentType, op)

		// Проверяем результат
		assert.Error(t, err)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrContentTypeNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("chekCategory success", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		category := int64(1)
		op := "test"

		// Мокаем проверку категории
		categoryExists := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(category).
			WillReturnRows(categoryExists)

		// Вызываем тестируемую функцию
		err := storage.chekCategory(ctx, category, op)

		// Проверяем результат
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("chekCategory failure", func(t *testing.T) {
		// Подготавливаем тестовое окружение
		_, mock, storage := setupTestDB(t)
		defer mock.ExpectClose()

		ctx := context.Background()
		category := int64(999)
		op := "test"

		// Мокаем проверку категории - категория не существует
		categoryExists := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM categories WHERE id = \$1 AND deleted IS NOT TRUE\)`).
			WithArgs(category).
			WillReturnRows(categoryExists)

		// Вызываем тестируемую функцию
		err := storage.chekCategory(ctx, category, op)

		// Проверяем результат
		assert.Error(t, err)
		assert.ErrorIs(t, errors.Unwrap(err), st.ErrNoCategoriesFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
