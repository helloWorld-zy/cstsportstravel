package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/middleware"
)

// WorkspaceProductHandler handles supplier workspace product management requests.
type WorkspaceProductHandler struct {
	logger *zap.Logger
}

// NewWorkspaceProductHandler creates a new WorkspaceProductHandler.
func NewWorkspaceProductHandler(logger *zap.Logger) *WorkspaceProductHandler {
	return &WorkspaceProductHandler{logger: logger}
}

// ListProducts handles GET /api/v2/supplier/products.
func (h *WorkspaceProductHandler) ListProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	supplierID := getSupplierID(c)
	status := c.DefaultQuery("status", "all")
	keyword := c.DefaultQuery("keyword", "")
	page := parseIntDefault(c, "page", 1)
	pageSize := parseIntDefault(c, "pageSize", 20)

	_ = tenantID
	_ = supplierID
	_ = status
	_ = keyword
	_ = page
	_ = pageSize

	// TODO: query products filtered by supplier_id for data isolation
	response.OK(c, gin.H{
		"items": []interface{}{},
		"total": 0,
	})
}

// CreateProductRequest represents a product creation request.
type CreateProductRequest struct {
	ProductName       string `json:"productName" binding:"required"`
	ProductType       string `json:"productType" binding:"required"`
	DestinationID     int64  `json:"destinationId" binding:"required"`
	Days              int    `json:"days" binding:"required"`
	Nights            int    `json:"nights"`
	DepartureCityIDs  []int  `json:"departureCityIds"`
	MinGroupSize      int    `json:"minGroupSize"`
	MaxGroupSize      int    `json:"maxGroupSize"`
	ProductLevel      string `json:"productLevel"`
	FeeIncluded       string `json:"feeIncluded"`
	FeeExcluded       string `json:"feeExcluded"`
	BookingNotice     string `json:"bookingNotice"`
}

// CreateProduct handles POST /api/v2/supplier/products.
func (h *WorkspaceProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)
	supplierID := getSupplierID(c)

	_ = tenantID
	_ = supplierID

	// TODO: create product with supplier_id association
	// TODO: set status to "pending_review"
	response.OK(c, gin.H{
		"message": "product created, pending review",
	})
}

// UpdateProduct handles PUT /api/v2/supplier/products/:id.
func (h *WorkspaceProductHandler) UpdateProduct(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	supplierID := getSupplierID(c)
	_ = supplierID
	_ = id

	// TODO: verify product belongs to supplier (data isolation)
	// TODO: if key fields changed, set status to "change_pending_review"
	response.OK(c, gin.H{
		"message": "product updated",
	})
}

// ToggleProductStatus handles POST /api/v2/supplier/products/:id/toggle.
func (h *WorkspaceProductHandler) ToggleProductStatus(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "invalid product id")
		return
	}

	supplierID := getSupplierID(c)
	_ = supplierID
	_ = id

	// TODO: toggle product active/inactive status
	response.OK(c, gin.H{
		"message": "product status toggled",
	})
}
