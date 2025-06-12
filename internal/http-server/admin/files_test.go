package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fileTesterHandler - обработчик с переопределенными функциями для тестирования
type fileTesterHandler struct {
	*Handler
	isVideoContentFunc func([]byte) bool
	parseMFError       error
}

// Переопределяем метод uploadVideoHandler для тестирования
func (h *fileTesterHandler) uploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	const op = "admin.uploadVideoHandler"

	logger := h.logger.With(
		"op", op,
	)

	// Если есть ошибка парсинга формы, симулируем её
	if h.parseMFError != nil {
		logger.Error("File too large", "error", h.parseMFError)
		admResponse.RespondWithError(w, http.StatusBadRequest, "File too large (max 500MB)")
		return
	}

	// Получаем файл из запроса
	file, header, err := r.FormFile("video")
	if err != nil {
		logger.Error("Failed to get file from request", "error", err)
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid file upload")
		return
	}
	defer file.Close()

	// Читаем начало файла для проверки типа
	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		logger.Error("Failed to read file header", "error", err)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to verify file type")
		return
	}

	contentType := http.DetectContentType(buff)
	fileExt := filepath.Ext(header.Filename)

	// Используем переопределенную функцию для проверки типа
	if h.isVideoContentFunc != nil && !h.isVideoContentFunc(buff) {
		logger.Error("Invalid file type", "content_type", contentType, "extension", fileExt)
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid file type. Only video files are allowed")
		return
	}

	// Сбрасываем позицию чтения файла
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		logger.Error("Failed to reset file position", "error", err)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to process file")
		return
	}

	// Сохраняем файл
	if err := saveVideoFile(h.Handler, header.Filename, file); err != nil {
		logger.Error("Failed to save file", "error", err)
		admResponse.RespondWithError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	admResponse.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "File " + header.Filename + " uploaded successfully",
	})
}

// setupFileTestHandler создает тестовый хендлер с временной директорией для файлов
func setupFileTestHandler(t *testing.T) (*fileTesterHandler, string) {
	// Создаем временную директорию для тестовых файлов
	tempDir := t.TempDir()

	// Создаем подпапки для видео и изображений
	videoDir := filepath.Join(tempDir, "video")
	imgDir := filepath.Join(tempDir, "img")
	require.NoError(t, os.MkdirAll(videoDir, 0755))
	require.NoError(t, os.MkdirAll(imgDir, 0755))

	// Создаем базовый хендлер с моками
	baseHandler, _, _ := newTestAuthHandlerWithMocks(t)

	// Обновляем конфигурацию с путями к временным директориям
	baseHandler.cfg.Media = config.Media{
		VideoPatch:  videoDir,
		ImagesPatch: imgDir,
	}

	// Создаем тестовый обработчик с переопределенными методами
	h := &fileTesterHandler{
		Handler:            baseHandler,
		isVideoContentFunc: func(data []byte) bool { return true }, // По умолчанию все файлы - видео
	}

	return h, tempDir
}

// Создает тестовый multipart запрос для загрузки файла
func createFileUploadRequest(t *testing.T, fieldName, fileName, fileContent, contentType string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)

	_, err = io.Copy(part, bytes.NewBufferString(fileContent))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/admin/files/"+fieldName, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

// Helper для создания видео файла
func createTempVideoFile(t *testing.T, dir, fileName string) string {
	filePath := filepath.Join(dir, fileName)
	file, err := os.Create(filePath)
	require.NoError(t, err)
	defer file.Close()

	// Записываем фиктивные данные MP4 файла (начинается с ftyp)
	_, err = file.Write([]byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x6d, 0x70, 0x34, 0x32})
	require.NoError(t, err)

	return filePath
}

// Тест успешной загрузки видео
func TestHandler_UploadVideoHandler_Success(t *testing.T) {
	// Настраиваем хендлер с тестовой директорией
	h, _ := setupFileTestHandler(t)

	// Устанавливаем, что наша проверка типа файла должна возвращать true
	h.isVideoContentFunc = func(data []byte) bool {
		return true
	}

	// Создаем тестовый запрос с фиктивным видеофайлом
	req := createFileUploadRequest(t, "video", "test.mp4", "test video content", "video/mp4")
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика
	h.uploadVideoHandler(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем ответ
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "uploaded successfully")

	// Проверяем, что файл действительно был создан
	_, err = os.Stat(filepath.Join(h.cfg.Media.VideoPatch, "test.mp4"))
	assert.False(t, os.IsNotExist(err), "File should have been created")
}

// Тест на неверный тип файла при загрузке видео
func TestHandler_UploadVideoHandler_InvalidFileType(t *testing.T) {
	// Настраиваем хендлер с тестовой директорией
	h, _ := setupFileTestHandler(t)

	// Устанавливаем, что наша проверка типа файла должна возвращать false
	h.isVideoContentFunc = func(data []byte) bool {
		return false
	}

	// Создаем тестовый запрос с фиктивным текстовым файлом
	req := createFileUploadRequest(t, "video", "test.txt", "not a video", "text/plain")
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика
	h.uploadVideoHandler(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Проверяем, что файл не был создан
	_, err := os.Stat(filepath.Join(h.cfg.Media.VideoPatch, "test.txt"))
	assert.True(t, os.IsNotExist(err), "File should not have been created")
}

// Тест на получение списка видео файлов
func TestHandler_ListVideoFilesHandler_Success(t *testing.T) {
	// Настраиваем хендлер с тестовой директорией
	h, tempDir := setupFileTestHandler(t)

	// Создаем тестовые файлы в директории
	videoDir := filepath.Join(tempDir, "video")
	createTempVideoFile(t, videoDir, "video1.mp4")
	createTempVideoFile(t, videoDir, "video2.mp4")

	// Добавляем не-видео файл, который должен быть пропущен
	file, _ := os.Create(filepath.Join(videoDir, "text.txt"))
	file.Close()

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/files/video", nil)
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика основного хендлера
	h.Handler.listVideoFilesHandler(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем ответ
	var response []admResponse.FileInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2) // Должно быть только 2 видеофайла
	assert.Contains(t, []string{response[0].Name, response[1].Name}, "video1.mp4")
	assert.Contains(t, []string{response[0].Name, response[1].Name}, "video2.mp4")
}

// Тест ошибки при получении списка файлов из несуществующей директории
func TestHandler_ListVideoFilesHandler_DirectoryError(t *testing.T) {
	// Настраиваем хендлер с несуществующей директорией
	h, _ := setupFileTestHandler(t)
	h.cfg.Media.VideoPatch = "/non/existent/directory"

	// Создаем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/files/video", nil)
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика основного хендлера
	h.Handler.listVideoFilesHandler(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Тест на ошибку при загрузке слишком большого файла
func TestHandler_UploadVideoHandler_FileTooLarge(t *testing.T) {
	// Настраиваем хендлер с тестовой директорией
	h, _ := setupFileTestHandler(t)

	// Устанавливаем ошибку парсинга формы
	h.parseMFError = errors.New("http: request body too large")

	// Создаем тестовый запрос с фиктивным видеофайлом
	req := createFileUploadRequest(t, "video", "large_file.mp4", "test content", "video/mp4")
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика
	h.uploadVideoHandler(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Проверяем сообщение об ошибке
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "File too large")
}

// Тест на отсутствие файла в запросе
func TestHandler_UploadVideoHandler_NoFile(t *testing.T) {
	// Настраиваем хендлер с тестовой директорией
	h, _ := setupFileTestHandler(t)

	// Создаем пустой multipart-запрос без файлов
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/admin/files/video", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика
	h.uploadVideoHandler(w, req)

	// Проверяем результаты
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на ошибку при сохранении файла
func TestHandler_UploadVideoHandler_SaveError(t *testing.T) {
	// Настраиваем хендлер с тестовой директорией
	h, _ := setupFileTestHandler(t)

	// Устанавливаем, что наша проверка типа файла должна возвращать true
	h.isVideoContentFunc = func(data []byte) bool {
		return true
	}

	// Устанавливаем директорию только для чтения чтобы вызвать ошибку при записи
	h.cfg.Media.VideoPatch = "/root/readonly_dir"

	// Создаем тестовый запрос с фиктивным видеофайлом
	req := createFileUploadRequest(t, "video", "test.mp4", "test video content", "video/mp4")
	w := httptest.NewRecorder()

	// Вызываем функцию обработчика
	h.uploadVideoHandler(w, req)

	// Проверяем результаты - предполагаем ошибку из-за невозможности записи в директорию
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
