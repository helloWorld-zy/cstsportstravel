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

// WithdrawalHandler handles distributor withdrawal operations.
type WithdrawalHandler struct {
	distributorRepo  *repository.DistributorRepository
	withdrawalRepo   *repository.WithdrawalRepository
	commissionRepo   *repository.CommissionRepository
	db               *gorm.DB
	logger           *zap.Logger
	minWithdrawal    float64
}

// NewWithdrawalHandler creates a new WithdrawalHandler.
func NewWithdrawalHandler(
	distributorRepo *repository.DistributorRepository,
	withdrawalRepo *repository.WithdrawalRepository,
	commissionRepo *repository.CommissionRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *WithdrawalHandler {
	return &WithdrawalHandler{
		distributorRepo: distributorRepo,
		withdrawalRepo:  withdrawalRepo,
		commissionRepo:  commissionRepo,
		db:              db,
		logger:          logger,
		minWithdrawal:   100, // 最低提现门槛100元
	}
}

// CreateWithdrawalRequest represents the request body for creating a withdrawal.
type CreateWithdrawalRequest struct {
	Amount float64 `json:"amount" binding:"required,min=100"`
}

// CreateWithdrawal handles POST /api/v2/distributor/withdrawals
// PRD §8.5.4: 最低提现门槛100元
func (h *WithdrawalHandler) CreateWithdrawal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	var req CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

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

	// Check distributor status
	if !distributor.IsActive() {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "Distributor is not active"})
		return
	}

	// Check minimum withdrawal amount
	if req.Amount < h.minWithdrawal {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("Minimum withdrawal amount is %.2f", h.minWithdrawal),
		})
		return
	}

	// Check withdrawable balance
	if req.Amount > distributor.WithdrawableAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Insufficient withdrawable balance",
		})
		return
	}

	// Check bank info
	if distributor.BankName == "" || distributor.BankAccountNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Bank information is incomplete"})
		return
	}

	// Generate withdrawal number
	withdrawalNo, err := h.withdrawalRepo.GenerateWithdrawalNo()
	if err != nil {
		h.logger.Error("failed to generate withdrawal number", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Create withdrawal record
	withdrawal := &domain.WithdrawalRecord{
		TenantID:          distributor.TenantID,
		WithdrawalNo:      withdrawalNo,
		DistributorID:     distributor.ID,
		Amount:            req.Amount,
		BankName:          distributor.BankName,
		BankAccountName:   distributor.BankAccountName,
		BankAccountNumber: distributor.BankAccountNumber,
		Status:            domain.WithdrawalStatusPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Create withdrawal record
		if err := tx.Create(withdrawal).Error; err != nil {
			return err
		}

		// Deduct from withdrawable balance
		if err := tx.Model(&domain.Distributor{}).
			Where("id = ?", distributor.ID).
			Updates(map[string]interface{}{
				"withdrawable_amount": gorm.Expr("withdrawable_amount - ?", req.Amount),
				"updated_at":          time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		h.logger.Error("failed to create withdrawal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Withdrawal request submitted",
		"data": gin.H{
			"withdrawal_no": withdrawal.WithdrawalNo,
			"amount":        withdrawal.Amount,
			"status":        withdrawal.Status,
		},
	})
}

// ListWithdrawals handles GET /api/v2/distributor/withdrawals
func (h *WithdrawalHandler) ListWithdrawals(c *gin.Context) {
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

	records, total, err := h.withdrawalRepo.FindByDistributorID(distributor.ID, status, pageNum, pageSizeNum)
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
