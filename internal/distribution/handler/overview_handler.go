package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/repository"
)

// OverviewHandler handles distributor overview operations.
type OverviewHandler struct {
	distributorRepo *repository.DistributorRepository
	commissionRepo  *repository.CommissionRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewOverviewHandler creates a new OverviewHandler.
func NewOverviewHandler(
	distributorRepo *repository.DistributorRepository,
	commissionRepo *repository.CommissionRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *OverviewHandler {
	return &OverviewHandler{
		distributorRepo: distributorRepo,
		commissionRepo:  commissionRepo,
		db:              db,
		logger:          logger,
	}
}

// GetOverview handles GET /api/v2/distributor/overview
// PRD §8.5.1: 首页为数据概览看板
func (h *OverviewHandler) GetOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

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

	// Get commission sums by status
	_, err = h.commissionRepo.SumByStatus(distributor.ID)
	if err != nil {
		h.logger.Error("failed to get commission sums", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"total_commission":    distributor.TotalCommission,
			"withdrawable_amount": distributor.WithdrawableAmount,
			"frozen_amount":       distributor.FrozenAmount,
			"today_orders":        0, // TODO: Calculate from orders
			"today_commission":    0, // TODO: Calculate from commissions
			"announcements":       []gin.H{}, // TODO: Fetch from announcement service
		},
	})
}
