package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/repository"
)

// PromotionStatsHandler handles promotion statistics operations.
type PromotionStatsHandler struct {
	distributorRepo   *repository.DistributorRepository
	promotionLinkRepo *repository.PromotionLinkRepository
	db                *gorm.DB
	logger            *zap.Logger
}

// NewPromotionStatsHandler creates a new PromotionStatsHandler.
func NewPromotionStatsHandler(
	distributorRepo *repository.DistributorRepository,
	promotionLinkRepo *repository.PromotionLinkRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *PromotionStatsHandler {
	return &PromotionStatsHandler{
		distributorRepo:   distributorRepo,
		promotionLinkRepo: promotionLinkRepo,
		db:                db,
		logger:            logger,
	}
}

// GetPromotionStats handles GET /api/v2/distributor/promotion-stats
// PRD §8.5.2: 推广数据统计
func (h *PromotionStatsHandler) GetPromotionStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	period := c.DefaultQuery("period", "last7days")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	pageNum := 1
	pageSizeNum := 20
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &pageSizeNum)

	distributor, err := h.distributorRepo.FindByUserID(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Distributor not found"})
			return
		}
		h.logger.Error("failed to find distributor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Get promotion links
	links, total, err := h.promotionLinkRepo.FindByDistributorID(distributor.ID, pageNum, pageSizeNum)
	if err != nil {
		h.logger.Error("failed to list promotion links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Build stats
	type ProductStat struct {
		ProductID    int64   `json:"product_id"`
		ProductName  string  `json:"product_name"`
		ClickPV      int64   `json:"click_pv"`
		ClickUV      int64   `json:"click_uv"`
		OrderCount   int64   `json:"order_count"`
		OrderAmount  float64 `json:"order_amount"`
		CommissionRate float64 `json:"commission_rate"`
	}

	stats := make([]ProductStat, 0)
	for _, link := range links {
		stats = append(stats, ProductStat{
			ProductID:   link.ProductID,
			ClickPV:     link.ClickPV,
			ClickUV:     link.ClickUV,
			OrderCount:  link.OrderCount,
			OrderAmount: link.OrderAmount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"period": period,
			"items":  stats,
			"total":  total,
		},
	})
}
