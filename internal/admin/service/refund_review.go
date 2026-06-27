// Package service provides business logic for the Admin domain.
package service

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	orderservice "github.com/travel-booking/server/internal/order/service"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
	productmodel "github.com/travel-booking/server/internal/product/model"
)

// Refund review errors.
var (
	ErrRefundNotFound       = errors.New("refund not found")
	ErrRefundNotPending     = errors.New("refund is not in pending status")
	ErrInsufficientApproval = errors.New("insufficient approval authority for this amount")
)

// AdminRefundReviewService provides business logic for admin refund review.
type AdminRefundReviewService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAdminRefundReviewService creates a new AdminRefundReviewService.
func NewAdminRefundReviewService(db *gorm.DB, logger *zap.Logger) *AdminRefundReviewService {
	return &AdminRefundReviewService{db: db, logger: logger}
}

// --- Request/Response DTOs ---

// RefundListRequest holds query parameters for refund listing.
type RefundListRequest struct {
	Status   string `form:"status" binding:"omitempty,oneof=pending approved processing success failed"`
	OrderNo  string `form:"order_no"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}

// RefundListItem is a single refund in the list.
type RefundListItem struct {
	ID              int64   `json:"id"`
	OrderID         int64   `json:"order_id"`
	OrderNo         string  `json:"order_no"`
	UserPhone       string  `json:"user_phone"`
	ProductName     string  `json:"product_name"`
	RefundNo        string  `json:"refund_no"`
	RefundAmount    int64   `json:"refund_amount"`
	RefundReason    string  `json:"refund_reason"`
	RefundType      string  `json:"refund_type"`
	Status          string  `json:"status"`
	ApprovalLevel   string  `json:"approval_level"`
	ApprovalLevelCN string  `json:"approval_level_cn"`
	CreatedAt       string  `json:"created_at"`
}

// RefundDetailResponse is the full refund detail.
type RefundDetailResponse struct {
	RefundListItem
	PayableAmount     int64                      `json:"payable_amount"`
	DaysBeforeDeparture int                      `json:"days_before_departure"`
	MatchingRule      string                     `json:"matching_rule"`
	RefundPercentage  float64                    `json:"refund_percentage"`
	ApprovedBy        *int64                     `json:"approved_by,omitempty"`
	ApprovedAt        *string                    `json:"approved_at,omitempty"`
	CompletedAt       *string                    `json:"completed_at,omitempty"`
	Order             *AdminOrderResponse        `json:"order,omitempty"`
}

// PaginatedRefundListResponse is the paginated refund list.
type PaginatedRefundListResponse struct {
	Items    []RefundListItem `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// ApproveRefundRequest is the request body for approving a refund.
type ApproveRefundRequest struct {
	Note string `json:"note"`
}

// RejectRefundRequest is the request body for rejecting a refund.
type RejectRefundRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// --- Service Methods ---

// ListRefundRequests returns a paginated refund request list.
func (s *AdminRefundReviewService) ListRefundRequests(req RefundListRequest) (*PaginatedRefundListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	query := s.db.Model(&paymentmodel.RefundRecord{})

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.OrderNo != "" {
		query = query.Where("order_id IN (SELECT id FROM main_order WHERE order_no = ?)", req.OrderNo)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count refunds: %w", err)
	}

	// Paginate
	offset := (req.Page - 1) * req.PageSize
	var records []paymentmodel.RefundRecord
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find refunds: %w", err)
	}

	// Batch load orders and products
	orderIDs := make(map[int64]bool)
	for _, r := range records {
		orderIDs[r.OrderID] = true
	}

	ordersMap := make(map[int64]*ordermodel.MainOrder)
	productIDs := make(map[int64]bool)
	if len(orderIDs) > 0 {
		var ids []int64
		for id := range orderIDs {
			ids = append(ids, id)
		}
		var orders []ordermodel.MainOrder
		if err := s.db.Where("id IN ?", ids).Find(&orders).Error; err == nil {
			for i := range orders {
				ordersMap[orders[i].ID] = &orders[i]
				productIDs[orders[i].ProductID] = true
			}
		}
	}

	productsMap := make(map[int64]*productmodel.Product)
	if len(productIDs) > 0 {
		var ids []int64
		for id := range productIDs {
			ids = append(ids, id)
		}
		var products []productmodel.Product
		if err := s.db.Where("id IN ?", ids).Find(&products).Error; err == nil {
			for i := range products {
				productsMap[products[i].ID] = &products[i]
			}
		}
	}

	// Build response
	items := make([]RefundListItem, len(records))
	for i, r := range records {
		orderNo := ""
		prodName := ""
		userPhone := ""
		if o, ok := ordersMap[r.OrderID]; ok {
			orderNo = o.OrderNo
			userPhone = maskPhone(o.ContactPhone)
			if p, ok := productsMap[o.ProductID]; ok {
				prodName = p.ProductName
			}
		}

		items[i] = RefundListItem{
			ID:              r.ID,
			OrderID:         r.OrderID,
			OrderNo:         orderNo,
			UserPhone:       userPhone,
			ProductName:     prodName,
			RefundNo:        r.RefundNo,
			RefundAmount:    r.RefundAmount,
			RefundReason:    r.RefundReason,
			RefundType:      r.RefundType,
			Status:          r.Status,
			ApprovalLevel:   r.ApprovalLevel,
			ApprovalLevelCN: approvalLevelCN(r.ApprovalLevel),
			CreatedAt:       r.CreatedAt.Format(time.RFC3339),
		}
	}

	return &PaginatedRefundListResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetRefundDetail returns the full refund detail.
func (s *AdminRefundReviewService) GetRefundDetail(refundID int64) (*RefundDetailResponse, error) {
	var record paymentmodel.RefundRecord
	if err := s.db.First(&record, refundID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, fmt.Errorf("find refund: %w", err)
	}

	// Load order
	var order ordermodel.MainOrder
	s.db.First(&order, record.OrderID)

	// Load product for name and refund rules
	var product productmodel.Product
	s.db.First(&product, order.ProductID)

	// Load departure to calculate days before
	var departure productmodel.DepartureDate
	s.db.Where("id = ?", order.DepartureID).First(&departure)

	daysBefore := orderservice.CalculateDaysBeforeDeparture(departure.DepartureDate)

	// Match cancellation rule
	engine := orderservice.NewCancellationEngine()
	match := engine.MatchRule(product.RefundRules, daysBefore)
	matchingRuleDesc := "无匹配退改规则"
	var refundPct float64
	if match != nil {
		matchingRuleDesc = orderservice.FormatRefundRuleDescription(match, daysBefore)
		refundPct = match.RefundPercentage
	}

	// Build timestamps
	var approvedAt, completedAt *string
	if record.ApprovedAt != nil {
		s := record.ApprovedAt.Format(time.RFC3339)
		approvedAt = &s
	}
	if record.CompletedAt != nil {
		s := record.CompletedAt.Format(time.RFC3339)
		completedAt = &s
	}

	orderNo := order.OrderNo

	return &RefundDetailResponse{
		RefundListItem: RefundListItem{
			ID:              record.ID,
			OrderID:         record.OrderID,
			OrderNo:         orderNo,
			UserPhone:       maskPhone(order.ContactPhone),
			ProductName:     product.ProductName,
			RefundNo:        record.RefundNo,
			RefundAmount:    record.RefundAmount,
			RefundReason:    record.RefundReason,
			RefundType:      record.RefundType,
			Status:          record.Status,
			ApprovalLevel:   record.ApprovalLevel,
			ApprovalLevelCN: approvalLevelCN(record.ApprovalLevel),
			CreatedAt:       record.CreatedAt.Format(time.RFC3339),
		},
		PayableAmount:       order.PayableAmount,
		DaysBeforeDeparture: daysBefore,
		MatchingRule:        matchingRuleDesc,
		RefundPercentage:    refundPct,
		ApprovedBy:          record.ApprovedBy,
		ApprovedAt:          approvedAt,
		CompletedAt:         completedAt,
	}, nil
}

// ApproveRefund approves a pending refund with tiered authority check.
func (s *AdminRefundReviewService) ApproveRefund(refundID int64, operatorID int64, operatorRoles []string, note string) error {
	record, err := s.findRefundByID(refundID)
	if err != nil {
		return err
	}

	if record.Status != paymentmodel.RefundStatusPending {
		return ErrRefundNotPending
	}

	// Check tiered approval authority
	if !canApprove(record.ApprovalLevel, operatorRoles) {
		return ErrInsufficientApproval
	}

	// Update refund status
	now := time.Now()
	if err := s.db.Model(&paymentmodel.RefundRecord{}).
		Where("id = ?", refundID).
		Updates(map[string]interface{}{
			"status":      paymentmodel.RefundStatusApproved,
			"approved_by": operatorID,
			"approved_at": now,
		}).Error; err != nil {
		return fmt.Errorf("update refund: %w", err)
	}

	s.logger.Info("refund approved by admin",
		zap.Int64("refund_id", refundID),
		zap.Int64("operator_id", operatorID),
		zap.String("approval_level", record.ApprovalLevel),
		zap.String("note", note),
	)

	return nil
}

// RejectRefund rejects a pending refund.
func (s *AdminRefundReviewService) RejectRefund(refundID int64, operatorID int64, reason string) error {
	record, err := s.findRefundByID(refundID)
	if err != nil {
		return err
	}

	if record.Status != paymentmodel.RefundStatusPending {
		return ErrRefundNotPending
	}

	// Update refund status to failed
	now := time.Now()
	if err := s.db.Model(&paymentmodel.RefundRecord{}).
		Where("id = ?", refundID).
		Updates(map[string]interface{}{
			"status":      paymentmodel.RefundStatusFailed,
			"approved_by": operatorID,
			"approved_at": now,
		}).Error; err != nil {
		return fmt.Errorf("update refund: %w", err)
	}

	// Revert order status from refunding back to paid_full
	if err := s.db.Model(&ordermodel.MainOrder{}).
		Where("id = ? AND order_status = ?", record.OrderID, ordermodel.OrderStatusRefunding).
		Updates(map[string]interface{}{
			"order_status":   ordermodel.OrderStatusPaidFull,
			"payment_status": ordermodel.PaymentStatusPaid,
			"updated_at":     now,
		}).Error; err != nil {
		s.logger.Error("failed to revert order status after refund rejection",
			zap.Int64("order_id", record.OrderID),
			zap.Error(err),
		)
	}

	// Create status log
	s.db.Create(&ordermodel.OrderStatusLog{
		OrderID:      record.OrderID,
		FromStatus:   ordermodel.OrderStatusRefunding,
		ToStatus:     ordermodel.OrderStatusPaidFull,
		OperatorType: "admin",
		OperatorID:   &operatorID,
		Reason:       fmt.Sprintf("refund rejected: %s", reason),
	})

	s.logger.Info("refund rejected by admin",
		zap.Int64("refund_id", refundID),
		zap.Int64("operator_id", operatorID),
		zap.String("reason", reason),
	)

	return nil
}

// findRefundByID finds a refund record by ID.
func (s *AdminRefundReviewService) findRefundByID(id int64) (*paymentmodel.RefundRecord, error) {
	var record paymentmodel.RefundRecord
	if err := s.db.First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return &record, nil
}

// canApprove checks if the operator with given roles can approve at the required level.
// Tiered approval per spec: ≤1000 operator, 1000-5000 finance_director, >5000 director.
func canApprove(requiredLevel string, operatorRoles []string) bool {
	// Admin role can approve everything
	for _, r := range operatorRoles {
		if r == "admin" || r == "director" {
			return true
		}
	}

	switch requiredLevel {
	case paymentmodel.ApprovalLevelOperator:
		// Any operator/finance_director/director can approve
		return true
	case paymentmodel.ApprovalLevelFinanceDirector:
		for _, r := range operatorRoles {
			if r == "finance_director" || r == "finance" {
				return true
			}
		}
		return false
	case paymentmodel.ApprovalLevelDirector:
		for _, r := range operatorRoles {
			if r == "director" {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// approvalLevelCN returns the Chinese label for an approval level.
func approvalLevelCN(level string) string {
	switch level {
	case paymentmodel.ApprovalLevelOperator:
		return "运营审批"
	case paymentmodel.ApprovalLevelFinanceDirector:
		return "财务主管审批"
	case paymentmodel.ApprovalLevelDirector:
		return "总监审批"
	default:
		return level
	}
}
