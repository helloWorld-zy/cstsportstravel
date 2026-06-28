// Package handler provides HTTP handlers for the Product domain.
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/product/repository"
)

// DestinationHandler handles HTTP requests for destination data.
type DestinationHandler struct {
	destRepo *repository.DestinationRepository
	logger   *zap.Logger
}

// NewDestinationHandler creates a new DestinationHandler.
func NewDestinationHandler(destRepo *repository.DestinationRepository, logger *zap.Logger) *DestinationHandler {
	return &DestinationHandler{
		destRepo: destRepo,
		logger:   logger,
	}
}

// PopularDestinationResponse is a popular destination with stats.
type PopularDestinationResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	CoverImage   string `json:"cover_image,omitempty"`
	ProductCount int    `json:"product_count"`
	MinPrice     int    `json:"min_price"` // display yuan
}

// ListPopularDestinations handles GET /api/v1/destinations/popular.
func (h *DestinationHandler) ListPopularDestinations(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	destinations, err := h.destRepo.FindPopularWithStats(limit)
	if err != nil {
		h.logger.Error("failed to get popular destinations", zap.Error(err))
		response.ServerError(c, "failed to get popular destinations")
		return
	}

	results := make([]PopularDestinationResponse, len(destinations))
	for i, d := range destinations {
		minPriceYuan := d.MinPrice / 100 // cents to yuan
		results[i] = PopularDestinationResponse{
			ID:           d.ID,
			Name:         d.Name,
			CoverImage:   d.CoverImage,
			ProductCount: d.ProductCount,
			MinPrice:     minPriceYuan,
		}
	}

	response.OK(c, results)
}
