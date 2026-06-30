// Package handler provides HTTP handlers for the Product domain.
package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/product/service"
)

// PreTripHandler handles HTTP requests for pre-trip services.
type PreTripHandler struct {
	outboundSvc *service.OutboundProductService
	logger      *zap.Logger
}

// NewPreTripHandler creates a new PreTripHandler.
func NewPreTripHandler(outboundSvc *service.OutboundProductService, logger *zap.Logger) *PreTripHandler {
	return &PreTripHandler{
		outboundSvc: outboundSvc,
		logger:      logger,
	}
}

// GetPreTripInfo handles GET /api/v2/visa/countries/:countryId/pretrip.
// Returns pre-trip information: entry policy, cash regulation, entry card guide, customs guide, emergency contacts.
func (h *PreTripHandler) GetPreTripInfo(c *gin.Context) {
	countryID, err := parseID(c, "countryId")
	if err != nil {
		response.BadRequest(c, "invalid country id")
		return
	}

	info, err := h.outboundSvc.GetPreTripInfo(countryID)
	if err != nil {
		h.logger.Error("failed to get pretrip info", zap.Int64("country_id", countryID), zap.Error(err))
		response.NotFound(c, "country not found")
		return
	}

	response.OK(c, info)
}
