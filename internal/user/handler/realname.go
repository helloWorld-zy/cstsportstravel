package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/user/service"
)

// RealNameHandler handles HTTP requests for real-name verification.
type RealNameHandler struct {
	realNameService *service.RealNameService
	logger          *zap.Logger
}

// NewRealNameHandler creates a new RealNameHandler.
func NewRealNameHandler(realNameService *service.RealNameService, logger *zap.Logger) *RealNameHandler {
	return &RealNameHandler{
		realNameService: realNameService,
		logger:          logger,
	}
}

// SubmitVerification handles POST /api/v1/users/me/real-name.
func (h *RealNameHandler) SubmitVerification(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req service.SubmitVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "real_name and id_card_no are required")
		return
	}

	result, err := h.realNameService.SubmitVerification(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidIDCard):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrDailyLimitExceeded):
			response.Fail(c, http.StatusTooManyRequests, response.CodeTooManyReq, err.Error())
		default:
			h.logger.Error("real-name verification failed", zap.Int64("user_id", userID), zap.Error(err))
			response.ServerError(c, "verification failed")
		}
		return
	}

	response.OK(c, result)
}
