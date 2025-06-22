package v1

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
	"regexp"
	"strings"
)

// @Summary Submit feedback
// @Description Saves user feedback to the system. Requires: message, at least one contact method (email or telegram), valid email format if provided, valid telegram handle (@ + 5-32 chars) if provided
// @Tags API v1
// @Accept json
// @Produce json
// @Param feedback body response.FeedbackResponse true "Feedback data to submit. Required fields: message. At least one of: email (valid format) or telegram (must start with @, 5-32 chars)"
// @Success 200 {string} string "Feedback successfully saved"
// @Failure 400 {object} response.ErrorResponse "Invalid request: possible reasons - missing required fields, invalid email/telegram format, no contact method provided"
// @Failure 500 {object} response.ErrorResponse "Server Error"
// @Router /feedback [post]
func (h *Handler) feedback(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.api.feedback"

	logger := h.logger.With(
		"op", op,
		"request_id", middleware.GetReqID(r.Context()),
	)

	var req response.FeedbackResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request body", sl.Err(err))
		admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Message == "" {
		logger.Warn("message is required")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Message is required")
		return
	}

	if req.Email != "" {
		if !isValidEmail(req.Email) {
			logger.Warn("invalid email format", "email", req.Email)
			admResponse.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
			return
		}
	}

	if req.Telegram != "" {
		if !isValidTelegram(req.Telegram) {
			logger.Warn("invalid telegram format", "telegram", req.Telegram)
			admResponse.RespondWithError(w, http.StatusBadRequest, "Telegram must start with @ and contain 5-32 characters")
			return
		}
	}

	if req.Email == "" && req.Telegram == "" {
		logger.Warn("no contact method provided")
		admResponse.RespondWithError(w, http.StatusBadRequest, "Either email or telegram must be provided")
		return
	}

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err := h.storage.Feedback(ctx, &req)
	if err != nil {
		logger.Error("feedback save error", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed save feedback")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, "Feedback successfully saved")
}

// Вспомогательные функции валидации
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidTelegram(telegram string) bool {
	if !strings.HasPrefix(telegram, "@") {
		return false
	}
	if len(telegram) < 6 || len(telegram) > 33 { // @ + 5-32 символа
		return false
	}
	telegramRegex := regexp.MustCompile(`^@[a-zA-Z0-9_]+$`)
	return telegramRegex.MatchString(telegram)
}
