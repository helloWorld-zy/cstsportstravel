package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/user/service"
)

// TravellerHandler handles HTTP requests for frequent travellers.
type TravellerHandler struct {
	travellerService *service.TravellerService
	logger           *zap.Logger
}

// NewTravellerHandler creates a new TravellerHandler.
func NewTravellerHandler(travellerService *service.TravellerService, logger *zap.Logger) *TravellerHandler {
	return &TravellerHandler{
		travellerService: travellerService,
		logger:           logger,
	}
}

// ListTravellers handles GET /api/v1/users/me/travellers.
func (h *TravellerHandler) ListTravellers(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	travellers, err := h.travellerService.List(userID)
	if err != nil {
		h.logger.Error("failed to list travellers", zap.Int64("user_id", userID), zap.Error(err))
		response.ServerError(c, "failed to list travellers")
		return
	}

	response.OK(c, travellers)
}

// CreateTraveller handles POST /api/v1/users/me/travellers.
func (h *TravellerHandler) CreateTraveller(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	var req service.CreateTravellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: real_name and id_card_no are required")
		return
	}

	traveller, err := h.travellerService.Create(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidIDCard):
			response.BadRequest(c, err.Error())
		case errors.Is(err, service.ErrMaxTravellersReached):
			response.BadRequest(c, err.Error())
		default:
			h.logger.Error("failed to create traveller", zap.Int64("user_id", userID), zap.Error(err))
			response.ServerError(c, "failed to create traveller")
		}
		return
	}

	response.OK(c, traveller)
}

// UpdateTraveller handles PUT /api/v1/users/me/travellers/:id.
func (h *TravellerHandler) UpdateTraveller(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	travellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid traveller id")
		return
	}

	var req service.UpdateTravellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	traveller, err := h.travellerService.Update(userID, travellerID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTravellerNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, service.ErrInvalidIDCard):
			response.BadRequest(c, err.Error())
		default:
			h.logger.Error("failed to update traveller", zap.Int64("user_id", userID), zap.Error(err))
			response.ServerError(c, "failed to update traveller")
		}
		return
	}

	response.OK(c, traveller)
}

// DeleteTraveller handles DELETE /api/v1/users/me/travellers/:id.
func (h *TravellerHandler) DeleteTraveller(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "authentication required")
		return
	}

	travellerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid traveller id")
		return
	}

	if err := h.travellerService.Delete(userID, travellerID); err != nil {
		if errors.Is(err, service.ErrTravellerNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		h.logger.Error("failed to delete traveller", zap.Int64("user_id", userID), zap.Error(err))
		response.ServerError(c, "failed to delete traveller")
		return
	}

	response.OKMessage(c, "traveller deleted")
}
