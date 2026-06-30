package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/middleware"
)

// WorkspaceOrderHandler handles supplier workspace order management requests.
type WorkspaceOrderHandler struct {
	logger *zap.Logger
}

// NewWorkspaceOrderHandler creates a new WorkspaceOrderHandler.
func NewWorkspaceOrderHandler(logger *zap.Logger) *WorkspaceOrderHandler {
	return &WorkspaceOrderHandler{logger: logger}
}

// ListOrders handles GET /api/v2/supplier/orders.
// Only returns orders belonging to the authenticated supplier (data isolation).
func (h *WorkspaceOrderHandler) ListOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	supplierID := getSupplierID(c)
	status := c.DefaultQuery("status", "all")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	page := parseIntDefault(c, "page", 1)
	pageSize := parseIntDefault(c, "pageSize", 20)

	_ = tenantID
	_ = supplierID
	_ = status
	_ = startDate
	_ = endDate
	_ = page
	_ = pageSize

	// TODO: query orders filtered by supplier_id for data isolation
	response.OK(c, gin.H{
		"items": []interface{}{},
		"total": 0,
	})
}

// OrderActionRequest represents an order confirm/reject request.
type OrderActionRequest struct {
	Action string `json:"action" binding:"required,oneof=confirm reject"`
	Reason string `json:"reason"`
}

// ConfirmOrder handles POST /api/v2/supplier/orders/:id/confirm.
func (h *WorkspaceOrderHandler) ConfirmOrder(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	var req OrderActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	supplierID := getSupplierID(c)
	_ = supplierID
	_ = id

	// TODO: verify order belongs to supplier (data isolation)
	// TODO: confirm or reject order
	response.OK(c, gin.H{
		"message": "order " + req.Action + "ed",
	})
}

// GetOrderDetail handles GET /api/v2/supplier/orders/:id.
func (h *WorkspaceOrderHandler) GetOrderDetail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid order id")
		return
	}

	supplierID := getSupplierID(c)
	_ = supplierID
	_ = id

	// TODO: verify order belongs to supplier (data isolation)
	// TODO: return order detail
	response.OK(c, gin.H{
		"orderId": id,
	})
}
