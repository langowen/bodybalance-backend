package v1

import (
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/langowen/bodybalance-backend/internal/http-server/admin/admResponse"
	"github.com/langowen/bodybalance-backend/internal/http-server/api/v1/response"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"net/http"
)

// @Summary Submit feedback
// @Description Saves user feedback to the system
// @Tags API v1
// @Accept json
// @Produce json
// @Param feedback body response.FeedbackResponse true "Feedback data to submit"
// @Success 200 {string} string "Feedback successfully saved"
// @Failure 400 {object} response.ErrorResponse "Invalid request format"
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

	ctx := logging.ContextWithLogger(r.Context(), logger)

	err := h.storage.Feedback(ctx, &req)
	if err != nil {
		logger.Error("feedback save error", sl.Err(err))
		response.RespondWithError(w, http.StatusInternalServerError, "Server Error", "Failed save feedback")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, "Feedback successfully saved")
}
