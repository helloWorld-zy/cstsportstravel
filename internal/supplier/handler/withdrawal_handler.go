package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/middleware"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// WithdrawalHandler handles supplier withdrawal requests.
type WithdrawalHandler struct {
	supplierRepo *repository.SupplierRepository
	logger       *zap.Logger
}

// NewWithdrawalHandler creates a new WithdrawalHandler.
func NewWithdrawalHandler(supplierRepo *repository.SupplierRepository, logger *zap.Logger) *WithdrawalHandler {
	return &WithdrawalHandler{
		supplierRepo: supplierRepo,
		logger:       logger,
	}
}

// WithdrawalRequest represents a withdrawal application request.
type WithdrawalRequest struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	BankAccountID int64   `json:"bankAccountId"`
}

// ApplyWithdrawal handles POST /api/v2/supplier/withdrawals.
func (h *WithdrawalHandler) ApplyWithdrawal(c *gin.Context) {
	var req WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	supplierID := getSupplierID(c)

	supplier, err := h.supplierRepo.FindByID(tenantID, supplierID)
	if err != nil {
		response.NotFound(c, "supplier not found")
		return
	}

	if !supplier.IsActive() {
		response.BusinessError(c, response.CodeForbidden, "supplier account is not active")
		return
	}

	// TODO: Check available balance
	// TODO: Create withdrawal record
	// TODO: Submit for approval

	response.OK(c, gin.H{
		"message": "withdrawal application submitted",
		"amount":  req.Amount,
	})
}

// ListWithdrawals handles GET /api/v2/supplier/withdrawals.
func (h *WithdrawalHandler) ListWithdrawals(c *gin.Context) {
	// TODO: implement withdrawal list with pagination
	response.OK(c, gin.H{
		"items": []interface{}{},
		"total": 0,
	})
}
