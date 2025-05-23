package response

import (
	"encoding/json"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

// VideoResponse представляет собой структуру возврата видео
type VideoResponse struct {
	ID          float64 `json:"id"`          //ID из БД
	URL         string  `json:"url"`         //URL адрес до файла
	Name        string  `json:"name"`        //Название видео
	Description string  `json:"description"` //Описание видео
	Category    string  `json:"category"`    //Название категории
}

// CategoryResponse представляет категорию с ID из БД
type CategoryResponse struct {
	ID   float64 `json:"id"`   // ID из БД
	Name string  `json:"name"` // Название категории
}

type AccountResponse struct {
	TypeID   float64 `json:"type_id"`
	TypeName string  `json:"type_name"`
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
