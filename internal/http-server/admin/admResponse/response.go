package admResponse

import (
	"encoding/json"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"time"
)

// VideoResponse представляет собой структуру возврата видео
type VideoResponse struct {
	ID          float64            `json:"id"`          // ID из БД
	URL         string             `json:"url"`         // URL адрес до файла
	Name        string             `json:"name"`        // Название видео
	Description string             `json:"description"` // Описание видео
	ImgURL      string             `json:"img_url"`     // Превью картинка для видео
	Categories  []CategoryResponse `json:"categories"`
	DateCreated string             `json:"created_at"`
}

// VideoRequest представляет структуру запроса для видео
type VideoRequest struct {
	URL         string  `json:"url"`          // URL адрес до файла
	Name        string  `json:"name"`         // Название видео
	Description string  `json:"description"`  // Описание видео
	ImgURL      string  `json:"img_url"`      // Превью картинка для видео
	CategoryIDs []int64 `json:"category_ids"` // ID категорий видео
}

type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

type TypeRequest struct {
	Name string `json:"name"` // Название типа
}

type TypeResponse struct {
	ID          float64 `json:"id"`   // ID из БД
	Name        string  `json:"name"` // Название типа
	DateCreated string  `json:"created_at,omitempty"`
}

type CategoryRequest struct {
	Name    string  `json:"name"`
	ImgURL  string  `json:"img_url"`
	TypeIDs []int64 `json:"type_ids"`
}

type CategoryResponse struct {
	ID          float64        `json:"id"`
	Name        string         `json:"name"`
	ImgURL      string         `json:"img_url,omitempty"`
	Types       []TypeResponse `json:"types,omitempty"`
	DateCreated string         `json:"date_created,omitempty"`
}

type UserRequest struct {
	Username      string `json:"username"`
	ContentTypeID string `json:"content_type_id"`
	Admin         bool   `json:"admin"`
	Password      string `json:"password"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	ContentTypeID string `json:"content_type_id"`
	ContentType   string `json:"content_type_name"`
	Admin         bool   `json:"admin"`
	DateCreated   string `json:"date_created"` // В формате 02.01.2006
}

type SignInRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type SignInResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func RespondWithError(w http.ResponseWriter, code int, message string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Error: message,
	}

	if len(details) > 0 {
		response.Details = details[0]
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode response", sl.Err(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
