package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/repository"
)

// PerformanceHandler handles performance dashboard operations.
type PerformanceHandler struct {
	distributorRepo *repository.DistributorRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewPerformanceHandler creates a new PerformanceHandler.
func NewPerformanceHandler(
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *PerformanceHandler {
	return &PerformanceHandler{
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
	}
}

// GetPerformance handles GET /api/v2/distributor/performance
// PRD §8.5.5: 业绩看板
func (h *PerformanceHandler) GetPerformance(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	period := c.DefaultQuery("period", "last30days")

	_, err := h.distributorRepo.FindByUserID(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Distributor not found"})
			return
		}
		h.logger.Error("failed to find distributor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// TODO: Calculate actual performance data from database
	// For now, return placeholder data structure
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"period": period,
			"trend": gin.H{
				"visitors":  []gin.H{},
				"orders":    []gin.H{},
				"amounts":   []gin.H{},
				"commissions": []gin.H{},
			},
			"summary": gin.H{
				"today": gin.H{
					"visitors":   0,
					"orders":     0,
					"amount":     0,
					"commission": 0,
				},
				"this_week": gin.H{
					"visitors":   0,
					"orders":     0,
					"amount":     0,
					"commission": 0,
				},
				"this_month": gin.H{
					"visitors":   0,
					"orders":     0,
					"amount":     0,
					"commission": 0,
				},
			},
			"product_ranking": []gin.H{},
			"channel_analysis": []gin.H{},
		},
	})
}
