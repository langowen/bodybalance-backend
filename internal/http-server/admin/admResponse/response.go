package admResponse

import (
	"encoding/json"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"time"
)

// VideoResponse представляет ответ с данными видео
// swagger:model videoResponse
type VideoResponse struct {
	ID          int64              `json:"id"`          // ID видео; example: 1
	URL         string             `json:"url"`         // URL видеофайла; example: https://example.com/video.mp4
	Name        string             `json:"name"`        // Название видео; example: Утренняя йога
	Description string             `json:"description"` // Описание видео; example: 30-минутный комплекс утренних упражнений
	ImgURL      string             `json:"img_url"`     // URL превью изображения; example: https://example.com/preview.jpg
	Categories  []CategoryResponse `json:"categories"`  // Список категорий видео
	DateCreated string             `json:"created_at"`  // Дата создания; example: 02.01.2006
}

// VideoRequest представляет запрос для создания/обновления видео
// swagger:model videoRequest
type VideoRequest struct {
	URL         string  `json:"url"`          // URL видеофайла; required: true; example: https://example.com/video.mp4
	Name        string  `json:"name"`         // Название видео; required: true; example: Утренняя йога
	Description string  `json:"description"`  // Описание видео; example: 30-минутный комплекс утренних упражнений
	ImgURL      string  `json:"img_url"`      // URL превью изображения; example: https://example.com/preview.jpg
	CategoryIDs []int64 `json:"category_ids"` // Список ID категорий; example: [1, 2, 3]
}

// FileInfo представляет информацию о файле
// swagger:model fileInfo
type FileInfo struct {
	Name    string    `json:"name"`     // Имя файла; example: video.mp4
	Size    int64     `json:"size"`     // Размер файла в байтах; example: 1024000
	ModTime time.Time `json:"mod_time"` // Время последнего изменения; example: 2023-01-01T12:00:00Z
}

// TypeRequest представляет запрос для создания/обновления типа
// swagger:model typeRequest
type TypeRequest struct {
	Name string `json:"name"` // Название типа; required: true; example: Йога
}

// TypeResponse представляет ответ с данными типа
// swagger:model typeResponse
type TypeResponse struct {
	ID          int64  `json:"id"`                   // ID типа; example: 1
	Name        string `json:"name"`                 // Название типа; example: Йога
	DateCreated string `json:"created_at,omitempty"` // Дата создания; example: 02.01.2006
}

// CategoryRequest представляет запрос для создания/обновления категории
// swagger:model categoryRequest
type CategoryRequest struct {
	Name    string  `json:"name"`     // Название категории; required: true; example: Утренние практики
	ImgURL  string  `json:"img_url"`  // URL изображения категории; example: https://example.com/category.jpg
	TypeIDs []int64 `json:"type_ids"` // Список ID типов; required: true; example: [1, 2]
}

// CategoryResponse представляет ответ с данными категории
// swagger:model categoryResponse
type CategoryResponse struct {
	ID          int64          `json:"id"`                     // ID категории; example: 1
	Name        string         `json:"name"`                   // Название категории; example: Утренние практики
	ImgURL      string         `json:"img_url,omitempty"`      // URL изображения категории; example: https://example.com/category.jpg
	Types       []TypeResponse `json:"types,omitempty"`        // Список типов категории
	DateCreated string         `json:"date_created,omitempty"` // Дата создания; example: 02.01.2006
}

// UserRequest представляет запрос для создания/обновления пользователя
// swagger:model userRequest
type UserRequest struct {
	Username      string `json:"username"`        // Имя пользователя; required: true; example: admin
	ContentTypeID int64  `json:"content_type_id"` // ID типа контента; example: 1
	Admin         bool   `json:"admin"`           // Флаг администратора; example: true
	Password      string `json:"password"`        // Пароль пользователя; required: true; example: hash(password123)
}

// UserResponse представляет ответ с данными пользователя
// swagger:model userResponse
type UserResponse struct {
	ID            int64  `json:"id"`                // ID пользователя; example: 1
	Username      string `json:"username"`          // Имя пользователя; example: admin
	ContentTypeID string `json:"content_type_id"`   // ID типа контента; example: 1
	ContentType   string `json:"content_type_name"` // Название типа контента; example: Йога
	Admin         bool   `json:"admin"`             // Флаг администратора; example: true
	DateCreated   string `json:"date_created"`      // Дата создания; example: 02.01.2006
}

// SignInRequest представляет запрос на аутентификацию
// swagger:parameters signInRequest
type SignInRequest struct {
	Login    string `json:"login"`    // Логин администратора; required: true; example: admin
	Password string `json:"password"` // Пароль администратора; required: true; example: hash(password123)
}

// SignInResponse представляет ответ на успешную аутентификацию
// swagger:model signInResponse
type SignInResponse struct {
	Message string `json:"message,omitempty"` // Сообщение об успехе; example: Authentication successful
	Error   string `json:"error,omitempty"`   // Сообщение об ошибке; example: Invalid credentials
}

// ErrorResponse представляет стандартный ответ об ошибке
// swagger:model errorResponse
type ErrorResponse struct {
	Error   string `json:"error"`             // Текст ошибки; example: Invalid request
	Details string `json:"details,omitempty"` // Детали ошибки (опционально); example: Login field is required
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
