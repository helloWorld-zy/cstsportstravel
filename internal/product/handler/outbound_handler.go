// Package handler provides HTTP handlers for the Product domain.
package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/product/service"
)

// OutboundHandler handles HTTP requests for outbound travel products.
type OutboundHandler struct {
	outboundSvc *service.OutboundProductService
	logger      *zap.Logger
}

// NewOutboundHandler creates a new OutboundHandler.
func NewOutboundHandler(outboundSvc *service.OutboundProductService, logger *zap.Logger) *OutboundHandler {
	return &OutboundHandler{
		outboundSvc: outboundSvc,
		logger:      logger,
	}
}

// ListOutboundProducts handles GET /api/v2/products/outbound.
// Supports filters: continent, country_id, visa_type, origin_city, days_min, days_max, keyword, sort, page, page_size.
func (h *OutboundHandler) ListOutboundProducts(c *gin.Context) {
	var req service.ListOutboundProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "invalid query parameters")
		return
	}

	result, err := h.outboundSvc.ListOutboundProducts(req)
	if err != nil {
		h.logger.Error("failed to list outbound products", zap.Error(err))
		response.ServerError(c, "failed to list outbound products")
		return
	}

	response.OK(c, result)
}

// GetContinentTree handles GET /api/v2/products/outbound/continents.
// Returns continent→country hierarchy for filter UI.
func (h *OutboundHandler) GetContinentTree(c *gin.Context) {
	tree, err := h.outboundSvc.GetContinentTree()
	if err != nil {
		h.logger.Error("failed to get continent tree", zap.Error(err))
		response.ServerError(c, "failed to get continent tree")
		return
	}

	response.OK(c, tree)
}

// GetOutboundProductDetail handles GET /api/v2/products/outbound/:id.
// Returns product detail with visa info card, flight info, material preview.
func (h *OutboundHandler) GetOutboundProductDetail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	detail, err := h.outboundSvc.GetOutboundProductDetail(id)
	if err != nil {
		h.logger.Error("failed to get outbound product detail", zap.Int64("id", id), zap.Error(err))
		response.NotFound(c, "product not found")
		return
	}

	response.OK(c, detail)
}

// GetCountryVisaInfo handles GET /api/v2/visa/countries/:countryId/info.
// Returns visa information for a destination country.
func (h *OutboundHandler) GetCountryVisaInfo(c *gin.Context) {
	countryID, err := parseID(c, "countryId")
	if err != nil {
		response.BadRequest(c, "invalid country id")
		return
	}

	pretrip, err := h.outboundSvc.GetPreTripInfo(countryID)
	if err != nil {
		h.logger.Error("failed to get country visa info", zap.Int64("country_id", countryID), zap.Error(err))
		response.NotFound(c, "country not found")
		return
	}

	response.OK(c, pretrip)
}
