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
	// ID видео
	// example: 1
	ID int64 `json:"id"`

	// URL видеофайла
	// example: https://example.com/video.mp4
	URL string `json:"url"`

	// Название видео
	// example: Утренняя йога
	Name string `json:"name"`

	// Описание видео
	// example: 30-минутный комплекс утренних упражнений
	Description string `json:"description"`

	// URL превью изображения
	// example: https://example.com/preview.jpg
	ImgURL string `json:"img_url"`

	// Список категорий видео
	Categories []CategoryResponse `json:"categories"`

	// Дата создания
	// example: 02.01.2006
	DateCreated string `json:"created_at"`
}

// VideoRequest представляет запрос для создания/обновления видео
// swagger:model videoRequest
type VideoRequest struct {
	// URL видеофайла
	// required: true
	// example: https://example.com/video.mp4
	URL string `json:"url"`

	// Название видео
	// required: true
	// example: Утренняя йога
	Name string `json:"name"`

	// Описание видео
	// example: 30-минутный комплекс утренних упражнений
	Description string `json:"description"`

	// URL превью изображения
	// example: https://example.com/preview.jpg
	ImgURL string `json:"img_url"`

	// Список ID категорий
	// example: [1, 2, 3]
	CategoryIDs []int64 `json:"category_ids"`
}

// FileInfo представляет информацию о файле
// swagger:model fileInfo
type FileInfo struct {
	// Имя файла
	// example: video.mp4
	Name string `json:"name"`
	// Размер файла в байтах
	// example: 1024000
	Size int64 `json:"size"`
	// Время последнего изменения
	// example: 2023-01-01T12:00:00Z
	ModTime time.Time `json:"mod_time"`
}

// TypeRequest представляет запрос для создания/обновления типа
// swagger:model typeRequest
type TypeRequest struct {
	// Название типа
	// required: true
	// example: Йога
	Name string `json:"name"`
}

// TypeResponse представляет ответ с данными типа
// swagger:model typeResponse
type TypeResponse struct {
	// ID типа
	// example: 1
	ID float64 `json:"id"`

	// Название типа
	// example: Йога
	Name string `json:"name"`

	// Дата создания
	// example: 02.01.2006
	DateCreated string `json:"created_at,omitempty"`
}

// CategoryRequest представляет запрос для создания/обновления категории
// swagger:model categoryRequest
type CategoryRequest struct {
	// Название категории
	// required: true
	// example: Утренние практики
	Name string `json:"name"`

	// URL изображения категории
	// example: https://example.com/category.jpg
	ImgURL string `json:"img_url"`

	// Список ID типов, к которым относится категория
	// required: true
	// example: [1, 2]
	TypeIDs []int64 `json:"type_ids"`
}

// CategoryResponse представляет ответ с данными категории
// swagger:model categoryResponse
type CategoryResponse struct {
	// ID категории
	// example: 1
	ID float64 `json:"id"`

	// Название категории
	// example: Утренние практики
	Name string `json:"name"`

	// URL изображения категории
	// example: https://example.com/category.jpg
	ImgURL string `json:"img_url,omitempty"`

	// Список типов категории
	Types []TypeResponse `json:"types,omitempty"`

	// Дата создания
	// example: 02.01.2006
	DateCreated string `json:"date_created,omitempty"`
}

// UserRequest представляет запрос для создания/обновления пользователя
// swagger:model userRequest
type UserRequest struct {
	// Имя пользователя
	// required: true
	// example: admin
	Username string `json:"username"`

	// ID типа контента
	// example: 1
	ContentTypeID string `json:"content_type_id"`

	// Флаг администратора
	// example: true
	Admin bool `json:"admin"`

	// Пароль пользователя
	// required: true
	// example: hash(password123)
	Password string `json:"password"`
}

// UserResponse представляет ответ с данными пользователя
// swagger:model userResponse
type UserResponse struct {
	// ID пользователя
	// example: 1
	ID string `json:"id"`

	// Имя пользователя
	// example: admin
	Username string `json:"username"`

	// ID типа контента
	// example: 1
	ContentTypeID string `json:"content_type_id"`

	// Название типа контента
	// example: Йога
	ContentType string `json:"content_type_name"`

	// Флаг администратора
	// example: true
	Admin bool `json:"admin"`

	// Дата создания
	// example: 02.01.2006
	DateCreated string `json:"date_created"`
}

// SignInRequest представляет запрос на аутентификацию
// swagger:parameters signInRequest
type SignInRequest struct {
	// Логин администратора
	// required: true,
	// example: admin
	Login string `json:"login"`
	// Пароль администратора
	// required: true,
	// example: hash(password123)
	Password string `json:"password"`
}

// SignInResponse представляет ответ на успешную аутентификацию
// swagger:model signInResponse
type SignInResponse struct {
	// Сообщение об успехе
	// example: Authentication successful
	Message string `json:"message,omitempty"`

	// Сообщение об ошибке
	// example: Invalid credentials
	Error string `json:"error,omitempty"`
}

// ErrorResponse представляет стандартный ответ об ошибке
// swagger:model errorResponse
type ErrorResponse struct {
	// Текст ошибки
	// example: Invalid request
	Error string `json:"error"`

	// Детали ошибки (опционально)
	// example: Login field is required
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
