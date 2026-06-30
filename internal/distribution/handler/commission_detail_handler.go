package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/repository"
)

// CommissionDetailHandler handles commission detail operations.
type CommissionDetailHandler struct {
	distributorRepo *repository.DistributorRepository
	commissionRepo  *repository.CommissionRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewCommissionDetailHandler creates a new CommissionDetailHandler.
func NewCommissionDetailHandler(
	distributorRepo *repository.DistributorRepository,
	commissionRepo *repository.CommissionRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *CommissionDetailHandler {
	return &CommissionDetailHandler{
		distributorRepo: distributorRepo,
		commissionRepo:  commissionRepo,
		db:              db,
		logger:          logger,
	}
}

// ListCommissions handles GET /api/v2/distributor/commissions
// PRD §8.5.4: 佣金明细列表
func (h *CommissionDetailHandler) ListCommissions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	status := c.DefaultQuery("status", "")
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

	details, total, err := h.commissionRepo.FindByDistributorID(distributor.ID, status, pageNum, pageSizeNum)
	if err != nil {
		h.logger.Error("failed to list commissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Calculate summary
	sums, _ := h.commissionRepo.SumByStatus(distributor.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items": details,
			"summary": gin.H{
				"total_amount":        distributor.TotalCommission,
				"frozen_amount":       sums["frozen"],
				"withdrawable_amount": sums["withdrawable"],
			},
			"total":     total,
			"page":      pageNum,
			"page_size": pageSizeNum,
		},
	})
}
