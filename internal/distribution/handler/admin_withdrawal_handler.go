package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// AdminWithdrawalHandler handles admin withdrawal management operations.
type AdminWithdrawalHandler struct {
	distributorRepo *repository.DistributorRepository
	withdrawalRepo  *repository.WithdrawalRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewAdminWithdrawalHandler creates a new AdminWithdrawalHandler.
func NewAdminWithdrawalHandler(
	distributorRepo *repository.DistributorRepository,
	withdrawalRepo *repository.WithdrawalRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *AdminWithdrawalHandler {
	return &AdminWithdrawalHandler{
		distributorRepo: distributorRepo,
		withdrawalRepo:  withdrawalRepo,
		db:              db,
		logger:          logger,
	}
}

// ListWithdrawals handles GET /api/v2/admin/distribution/withdrawals
// PRD §8.6.2: 提现审核列表
func (h *AdminWithdrawalHandler) ListWithdrawals(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	pageNum := 1
	pageSizeNum := 20
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &pageSizeNum)

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	records, total, err := h.withdrawalRepo.ListAll(tid, status, pageNum, pageSizeNum)
	if err != nil {
		h.logger.Error("failed to list withdrawals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items":     records,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSizeNum,
		},
	})
}

// ProcessWithdrawalRequest represents the request body for processing a withdrawal.
type ProcessWithdrawalRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason"`
}

// ProcessWithdrawal handles POST /api/v2/admin/distribution/withdrawals/:id/process
func (h *AdminWithdrawalHandler) ProcessWithdrawal(c *gin.Context) {
	id := c.GetInt64("id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid withdrawal ID"})
		return
	}

	var req ProcessWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	adminID, _ := c.Get("user_id")
	aid, _ := adminID.(int64)

	withdrawal, err := h.withdrawalRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Withdrawal not found"})
			return
		}
		h.logger.Error("failed to find withdrawal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	if !withdrawal.IsPending() {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Withdrawal is not pending"})
		return
	}

	var targetStatus string
	switch req.Action {
	case "approve":
		targetStatus = domain.WithdrawalStatusApproved
	case "reject":
		if req.Reason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Rejection reason is required"})
			return
		}
		targetStatus = domain.WithdrawalStatusRejected
	}

	if err := h.withdrawalRepo.UpdateStatus(id, targetStatus, &aid, req.Reason); err != nil {
		h.logger.Error("failed to process withdrawal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// If rejected, return amount to distributor's withdrawable balance
	if req.Action == "reject" {
		if err := h.db.Model(&domain.Distributor{}).
			Where("id = ?", withdrawal.DistributorID).
			Updates(map[string]interface{}{
				"withdrawable_amount": gorm.Expr("withdrawable_amount + ?", withdrawal.Amount),
				"updated_at":          time.Now(),
			}).Error; err != nil {
			h.logger.Error("failed to return amount to distributor", zap.Error(err))
		}
	}

	// TODO: Send notification to distributor

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Withdrawal processed",
	})
}

// BatchProcessWithdrawalRequest represents the request body for batch processing withdrawals.
type BatchProcessWithdrawalRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// BatchProcessWithdrawals handles POST /api/v2/admin/distribution/withdrawals/batch-process
func (h *AdminWithdrawalHandler) BatchProcessWithdrawals(c *gin.Context) {
	var req BatchProcessWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	adminID, _ := c.Get("user_id")
	aid, _ := adminID.(int64)

	processed := 0
	for _, id := range req.IDs {
		withdrawal, err := h.withdrawalRepo.FindByID(id)
		if err != nil {
			continue
		}

		if !withdrawal.IsPending() {
			continue
		}

		if err := h.withdrawalRepo.UpdateStatus(id, domain.WithdrawalStatusApproved, &aid, ""); err != nil {
			h.logger.Error("failed to process withdrawal in batch", zap.Int64("id", id), zap.Error(err))
			continue
		}

		processed++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Batch processing completed",
		"data": gin.H{
			"processed": processed,
			"total":     len(req.IDs),
		},
	})
}
