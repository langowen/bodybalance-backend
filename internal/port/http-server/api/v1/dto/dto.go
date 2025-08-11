package dto

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
)

// VideoResponse представляет информацию о видео
// @description Информация о видео, включая URL, описание и категорию
type VideoResponse struct {
	ID          int64  `json:"id"`          // ID из БД
	URL         string `json:"url"`         // URL адрес до файла
	Name        string `json:"name"`        // Название видео
	Description string `json:"description"` // Описание видео
	Category    string `json:"category"`    // Название категории
	ImgURL      string `json:"img_url"`     // Превью картинка для видео
}

// CategoryResponse представляет информацию о категории
// @description Информация о категории контента
type CategoryResponse struct {
	ID     int64  `json:"id"`      // ID из БД
	Name   string `json:"name"`    // Название категории
	ImgURL string `json:"img_url"` // Превью картинка для категории
}

// AccountResponse представляет информацию об аккаунте
// @description Информация о типе аккаунта пользователя
type AccountResponse struct {
	TypeID   int64  `json:"type_id"`   // ID type из БД
	TypeName string `json:"type_name"` // Название type
}

// FeedbackResponse представляет информацию о фидбэке от пользователя
// @description Информация об обратной связи, отправленная пользователем
type FeedbackResponse struct {
	Name     string `json:"name,omitempty"`     // Имя пользователя
	Email    string `json:"email,omitempty"`    // Email пользователя
	Telegram string `json:"telegram,omitempty"` // Telegram пользователя
	Message  string `json:"message"`            // Сообщение от пользователя
}

// HealthResponse предоставляет информацию о статусе сервера
// @description Информация о доступности сервисов
type HealthResponse struct {
	Status      string `json:"status"`       // Статус сервиса (ok, error)
	DbStatus    string `json:"db_status"`    // Статус подключения к БД (ok, error)
	RedisStatus string `json:"redis_status"` // Статус подключения к Redis (ok, disabled, error)
}

func RespondWithError(w http.ResponseWriter, code int, message string, details ...string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)

	errorText := message
	if len(details) > 0 {
		errorText += "\nDetails: " + details[0]
	}

	if _, err := w.Write([]byte(errorText)); err != nil {
		slog.Error("Failed to write error dto", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode dto", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
