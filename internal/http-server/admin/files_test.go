package admin

import (
	"bytes"
	"encoding/json"
	"github.com/langowen/bodybalance-backend/internal/config"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/stretchr/testify/assert"
	"github.com/theartofdevel/logging"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestUploadVideoHandler тестирует загрузку видео
func TestUploadVideoHandler(t *testing.T) {
	// 1. Создаем временный файл
	tmpFile, err := os.CreateTemp("", "test_video_*.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(tmpFile.Name()); err != nil {
		t.Log(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Log(err)
	}

	// 2. Используем реальные данные MP4 (минимум 512 байт)
	tmpFile, err = os.CreateTemp("", "test_video_*.mp4")
	if err != nil {
		t.Fatal(err)
	}
	testVideoData := makeValidMP4Header()
	if _, err = tmpFile.Write(testVideoData); err != nil {
		t.Fatal(err)
	}
	if _, err = tmpFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	// 3. Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("video", "test_video.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(part, tmpFile); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Log(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Log(err)
	}

	// 4. Настраиваем моки
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: t.TempDir(),
		},
	}

	// 5. Создаем Handler
	handler := New(logger, mockStorage, cfg, mockRedis)

	// 6. Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/admin/files/video", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// 7. Вызываем обработчик
	handler.uploadVideoHandler(w, req)

	// 8. Проверяем
	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", w.Code)
	}
}

// TestUploadMOVVideoHandler тестирует загрузку видео в формате MOV
func TestUploadMOVVideoHandler(t *testing.T) {
	// 1. Создаем временный файл для MOV
	tmpFile, err := os.CreateTemp("", "test_video_*.mov")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(tmpFile.Name()); err != nil {
		t.Log(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Log(err)
	}

	// 2. Используем данные MOV с правильным заголовком
	tmpFile, err = os.CreateTemp("", "test_video_*.mov")
	if err != nil {
		t.Fatal(err)
	}
	testVideoData := makeValidMOVHeader()
	if _, err = tmpFile.Write(testVideoData); err != nil {
		t.Fatal(err)
	}
	if _, err = tmpFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	// 3. Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("video", "test_video.mov")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(part, tmpFile); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Log(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Log(err)
	}

	// 4. Настраиваем моки
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: t.TempDir(),
		},
	}

	// 5. Создаем Handler
	handler := New(logger, mockStorage, cfg, mockRedis)

	// 6. Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/admin/files/video", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// 7. Вызываем обработчик
	handler.uploadVideoHandler(w, req)

	// 8. Проверяем, что MOV файл успешно загружен
	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус 200 для MOV файла, получен %d", w.Code)
	}
}

// TestIsVideoContent проверяет работу функции isVideoContent
func TestIsVideoContent(t *testing.T) {
	// Проверка MP4
	mp4Header := makeValidMP4Header()
	assert.True(t, isVideoContent(mp4Header), "MP4 должен определяться как видео")

	// Проверка MOV
	movHeader := makeValidMOVHeader()
	assert.True(t, isVideoContent(movHeader), "MOV должен определяться как видео")

	// Проверка WebM
	webmHeader := []byte{
		0x1A, 0x45, 0xDF, 0xA3, // EBML signature
		0x01, 0x00, 0x00, 0x00, // some WebM data
	}
	paddedWebm := append(webmHeader, make([]byte, 504)...) // до 512 байт
	assert.True(t, isVideoContent(paddedWebm), "WebM должен определяться как видео")

	// Проверка Ogg
	oggHeader := []byte{
		0x4F, 0x67, 0x67, 0x53, // 'OggS'
		0x00, 0x00, 0x00, 0x00, // some Ogg data
	}
	paddedOgg := append(oggHeader, make([]byte, 504)...) // до 512 байт
	assert.True(t, isVideoContent(paddedOgg), "Ogg должен определяться как видео")

	// Проверка не-видео файла
	textData := []byte("This is a plain text file, not a video file")
	paddedText := append(textData, make([]byte, 512-len(textData))...) // до 512 байт
	assert.False(t, isVideoContent(paddedText), "Текстовый файл не должен определяться как видео")
}

func makeValidMP4Header() []byte {
	// Минимальный валидный MP4 заголовок (512 байт)
	header := []byte{
		0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, // ftyp
		0x69, 0x73, 0x6F, 0x6D, // isom
		0x00, 0x00, 0x02, 0x00, 0x6D, 0x70, 0x34, 0x31, // mp41
		0x6D, 0x70, 0x34, 0x32, 0x00, 0x00, 0x00, 0x08, // mp42
		0x77, 0x69, 0x64, 0x65, // wide
	}
	// Дополняем до 512 байт
	if len(header) < 512 {
		padding := make([]byte, 512-len(header))
		header = append(header, padding...)
	}
	return header
}

// makeValidMOVHeader создает валидный заголовок для MOV-файла
func makeValidMOVHeader() []byte {
	// Минимальный валидный MOV заголовок (512 байт) с сигнатурой ftyp
	header := []byte{
		0x00, 0x00, 0x00, 0x14, 0x66, 0x74, 0x79, 0x70, // ftyp
		0x71, 0x74, 0x20, 0x20, 0x20, 0x04, 0x06, 0x00, // qt
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Далее идёт блок mdat или moov
		0x00, 0x00, 0x00, 0x08, 0x6D, 0x6F, 0x6F, 0x76, // moov
	}
	// Дополняем до 512 байт
	if len(header) < 512 {
		padding := make([]byte, 512-len(header))
		header = append(header, padding...)
	}
	return header
}

// TestUploadVideoHandler_InvalidFileType тестирует загрузку файла неверного типа
func TestUploadVideoHandler_InvalidFileType(t *testing.T) {
	// Создаем временный файл для мока multipart/form-data
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Заполняем файл тестовыми данными
	if _, err = tmpFile.WriteString("This is not a video file"); err != nil {
		t.Fatal(err)
	}
	if _, err = tmpFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	// Создаем multipart/form-data запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("video", "test_file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(part, tmpFile); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	// Настраиваем мок-хендлер
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}

	// Корректно инициализируем конфигурацию
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: t.TempDir(),
		},
	}

	handler := New(logger, mockStorage, cfg, mockRedis)

	// Создаем HTTP запрос
	req := httptest.NewRequest(http.MethodPost, "/admin/files/video", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.uploadVideoHandler(w, req)

	// Проверяем результат (ожидаем ошибку из-за неверного типа файла)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestListVideoFilesHandler тестирует получение списка видеофайлов
func TestListVideoFilesHandler(t *testing.T) {
	// Настраиваем мок-хендлер
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}

	// Создаем временную директорию и тестовые файлы
	tempDir := t.TempDir()
	testFiles := []struct {
		name string
		data []byte
	}{
		{"video1.mp4", []byte("test video 1")},
		{"video2.mp4", []byte("test video 2")},
		{"file.txt", []byte("not a video")}, // не должен попасть в список
	}
	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		if err := os.WriteFile(filePath, tf.data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tempDir,
		},
	}

	handler := New(logger, mockStorage, cfg, mockRedis)

	req := httptest.NewRequest(http.MethodGet, "/admin/files/video", nil)
	w := httptest.NewRecorder()

	handler.listVideoFilesHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []admResponse.FileInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	// Должны быть только mp4-файлы
	if len(response) != 2 {
		t.Errorf("ожидалось 2 видеофайла, получено %d", len(response))
	}
	videoNames := map[string]bool{}
	for _, f := range response {
		videoNames[f.Name] = true
	}
	if !videoNames["video1.mp4"] || !videoNames["video2.mp4"] {
		t.Error("В списке нет ожидаемых видеофайлов")
	}
}

// TestListVideoFilesHandler_Error тестирует ошибку при получении списка видеофайлов
func TestListVideoFilesHandler_Error(t *testing.T) {
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	// Указываем несуществующую директорию для видео
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: "/nonexistent_dir_for_test_error_case",
		},
	}

	handler := New(logger, mockStorage, cfg, mockRedis)

	req := httptest.NewRequest(http.MethodGet, "/admin/files/video", nil)
	w := httptest.NewRecorder()

	handler.listVideoFilesHandler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestUploadImageHandler тестирует загрузку изображения
func TestUploadImageHandler(t *testing.T) {
	// Создаем временный файл для мока multipart/form-data
	tmpFile, err := os.CreateTemp("", "test_image_*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Заполняем файл тестовыми данными, имитирующими JPEG
	fakeJPEGHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}
	fakeJPEGData := append(fakeJPEGHeader, bytes.Repeat([]byte{0}, 1000)...) // 1000 байт
	if _, err = tmpFile.Write(fakeJPEGData); err != nil {
		t.Fatal(err)
	}
	if _, err = tmpFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	// Создаем multipart/form-data запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "test_image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.Copy(part, tmpFile); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{
		Media: config.Media{
			ImagesPatch: t.TempDir(),
		},
	}

	handler := New(logger, mockStorage, cfg, mockRedis)

	req := httptest.NewRequest(http.MethodPost, "/admin/files/img", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.uploadImageHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "uploaded successfully")
}

// TestListImageFilesHandler тестирует получение списка изображений
func TestListImageFilesHandler(t *testing.T) {
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}

	tempDir := t.TempDir()
	testFiles := []struct {
		name string
		data []byte
	}{
		{"image1.jpg", []byte("test image 1")},
		{"image2.png", []byte("test image 2")},
		{"file.txt", []byte("not an image")},
	}
	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		if err := os.WriteFile(filePath, tf.data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{
		Media: config.Media{
			ImagesPatch: tempDir,
		},
	}

	handler := New(logger, mockStorage, cfg, mockRedis)

	req := httptest.NewRequest(http.MethodGet, "/admin/files/img", nil)
	w := httptest.NewRecorder()

	handler.listImageFilesHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []admResponse.FileInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	if len(response) != 2 {
		t.Errorf("ожидалось 2 изображения, получено %d", len(response))
	}
	imgNames := map[string]bool{}
	for _, f := range response {
		imgNames[f.Name] = true
	}
	if !imgNames["image1.jpg"] || !imgNames["image2.png"] {
		t.Error("В списке нет ожидаемых изображений")
	}
}

// TestSaveVideoFile тестирует сохранение видеофайла
func TestSaveVideoFile(t *testing.T) {
	// Создаем временный файл для тестирования
	tmpFile, err := os.CreateTemp("", "test_video_*.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Создаем тестовые данные
	testData := []byte("test video content")
	if _, err := tmpFile.Write(testData); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	// Настраиваем хендлер с временной директорией
	tempDir := t.TempDir()

	// Корректно инициализируем конфигурацию
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tempDir,
		},
	}

	handler := &Handler{cfg: cfg}

	// Вызываем функцию для сохранения файла
	filename := "test_output.mp4"
	err = saveVideoFile(handler, filename, tmpFile)
	assert.NoError(t, err)

	// Проверяем, что файл был создан
	outputPath := filepath.Join(tempDir, filename)
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)

	// Проверяем содержимое файла
	content, err := os.ReadFile(outputPath)
	assert.NoError(t, err)
	assert.Equal(t, testData, content)
}

// TestGetVideoFilesList тестирует получение списка видеофайлов
func TestGetVideoFilesList(t *testing.T) {
	// Пропускаем тест, если мы на CI или другой среде, где нет доступа к файловой системе
	if os.Getenv("CI") != "" {
		t.Skip("Skipping filesystem test in CI environment")
	}

	// Создаем временную директорию с тестовыми файлами
	tempDir := t.TempDir()

	// Создаем тестовые файлы
	testFiles := []struct {
		name string
		data []byte
	}{
		{"video1.mp4", []byte("test video 1")},
		{"video2.mp4", []byte("test video 2")},
		{"file.txt", []byte("not a video")}, // Этот файл не должен быть включен
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		if err := os.WriteFile(filePath, tf.data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Настраиваем хендлер
	cfg := &config.Config{
		Media: config.Media{
			VideoPatch: tempDir,
		},
	}

	handler := &Handler{cfg: cfg}

	// Получаем список файлов
	filesList, err := handler.getVideoFilesList()
	assert.NoError(t, err)

	// Проверяем количество файлов (только MP4)
	assert.Equal(t, 2, len(filesList))

	// Проверяем наличие видеофайлов в списке
	videoNames := make(map[string]bool)
	for _, f := range filesList {
		videoNames[f.Name] = true
	}

	assert.True(t, videoNames["video1.mp4"])
	assert.True(t, videoNames["video2.mp4"])
	assert.False(t, videoNames["file.txt"])
}

// TestHelperFunctions тестирует вспомогательные функции
func TestHelperFunctions(t *testing.T) {
	// Тест isValidFilename
	assert.True(t, isValidFilename("valid_filename.mp4"))
	assert.True(t, isValidFilename("valid-filename.mp4"))
	assert.False(t, isValidFilename("invalid/filename.mp4"))
	assert.False(t, isValidFilename("invalid\\filename.mp4"))
	assert.False(t, isValidFilename("invalid:filename.mp4"))

	// Тест isVideoExtension
	assert.True(t, isVideoExtension(".mp4"))
	assert.True(t, isVideoExtension(".webm"))
	assert.True(t, isVideoExtension(".ogg"))
	assert.True(t, isVideoExtension(".mov"))
	assert.True(t, isVideoExtension(".avi"))
	assert.False(t, isVideoExtension(".jpg"))
	assert.False(t, isVideoExtension(".txt"))

	// Тест isImageExtension
	assert.True(t, isImageExtension(".jpg"))
	assert.True(t, isImageExtension(".jpeg"))
	assert.True(t, isImageExtension(".png"))
	assert.True(t, isImageExtension(".gif"))
	assert.True(t, isImageExtension(".webp"))
	assert.False(t, isImageExtension(".mp4"))
	assert.False(t, isImageExtension(".txt"))
}

func TestUploadImageHandlerTooLarge(t *testing.T) {
	// Настраиваем хендлер
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	handler := New(logger, mockStorage, cfg, mockRedis)

	// Создаем HTTP запрос с заголовком Content-Length, превышающим лимит
	req := httptest.NewRequest(http.MethodPost, "/admin/files/img", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	req.ContentLength = maxImageUploadSize + 1
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.uploadImageHandler(w, req)

	// Проверяем результат (ожидаем ошибку из-за размера)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadVideoHandlerNoFile(t *testing.T) {
	// Настраиваем хендлер
	logger := logging.NewLogger()
	mockStorage := &MockAdmStorage{}
	mockRedis := &MockRedisClient{}
	cfg := &config.Config{}

	handler := New(logger, mockStorage, cfg, mockRedis)

	// Создаем пустой multipart запрос без файла
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	// Создаем HTTP запрос
	req := httptest.NewRequest(http.MethodPost, "/admin/files/video", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.uploadVideoHandler(w, req)

	// Проверяем результат (ожидаем ошибку из-за отсутствия файла)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
