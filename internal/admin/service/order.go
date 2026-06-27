// Package service provides business logic for the Admin domain.
package service

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
	productmodel "github.com/travel-booking/server/internal/product/model"
	usermodel "github.com/travel-booking/server/internal/user/model"
)

// AdminOrderService provides business logic for admin order management.
type AdminOrderService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAdminOrderService creates a new AdminOrderService.
func NewAdminOrderService(db *gorm.DB, logger *zap.Logger) *AdminOrderService {
	return &AdminOrderService{db: db, logger: logger}
}

// --- Request/Response DTOs ---

// AdminOrderListRequest holds query parameters for admin order listing.
type AdminOrderListRequest struct {
	OrderNo     string `form:"order_no"`
	UserPhone   string `form:"user_phone"`
	Status      string `form:"status"`
	DateFrom    string `form:"date_from"`
	DateTo      string `form:"date_to"`
	ProductType string `form:"product_type"`
	SupplierID  *int64 `form:"supplier_id"`
	Page        int    `form:"page" binding:"min=1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100"`
}

// AdminOrderResponse is the admin order list item.
type AdminOrderResponse struct {
	ID             int64   `json:"id"`
	OrderNo        string  `json:"order_no"`
	UserPhone      string  `json:"user_phone"`
	ProductName    string  `json:"product_name"`
	DepartureDate  string  `json:"departure_date,omitempty"`
	OrderStatus    string  `json:"order_status"`
	PayableAmount  int64   `json:"payable_amount"`
	AdultCount     int     `json:"adult_count"`
	ChildCount     int     `json:"child_count"`
	InfantCount    int     `json:"infant_count"`
	Channel        string  `json:"channel"`
	ContactName    string  `json:"contact_name"`
	CreatedAt      string  `json:"created_at"`
}

// AdminOrderDetailResponse is the full admin order detail.
type AdminOrderDetailResponse struct {
	AdminOrderResponse
	TotalAmount           int64                    `json:"total_amount"`
	DiscountAmount        int64                    `json:"discount_amount"`
	SingleSupplementAmount int64                   `json:"single_supplement_amount"`
	AddonAmount           int64                    `json:"addon_amount"`
	ContactPhone          string                   `json:"contact_phone"`
	Remark                string                   `json:"remark,omitempty"`
	PaymentStatus         string                   `json:"payment_status"`
	CancelReason          string                   `json:"cancel_reason,omitempty"`
	PaidAt                *string                  `json:"paid_at,omitempty"`
	CancelledAt           *string                  `json:"cancelled_at,omitempty"`
	CompletedAt           *string                  `json:"completed_at,omitempty"`
	Travellers            []OrderTravellerResponse `json:"travellers"`
	Payments              []OrderPaymentResponse   `json:"payments"`
	Refunds               []OrderRefundResponse    `json:"refunds"`
	StatusLogs            []OrderStatusLogResponse `json:"status_logs"`
}

// OrderTravellerResponse is a traveller in the order detail.
type OrderTravellerResponse struct {
	ID          int64  `json:"id"`
	Phone       string `json:"phone,omitempty"`
	Gender      string `json:"gender"`
	IsChild     bool   `json:"is_child"`
	IsInfant    bool   `json:"is_infant"`
	BirthDate   *string `json:"birth_date,omitempty"`
}

// OrderPaymentResponse is a payment record in the order detail.
type OrderPaymentResponse struct {
	ID             int64   `json:"id"`
	PaymentNo      string  `json:"payment_no"`
	Channel        string  `json:"channel"`
	Method         string  `json:"method"`
	Amount         int64   `json:"amount"`
	Status         string  `json:"status"`
	ChannelTradeNo string  `json:"channel_trade_no,omitempty"`
	PaidAt         *string `json:"paid_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// OrderRefundResponse is a refund record in the order detail.
type OrderRefundResponse struct {
	ID              int64   `json:"id"`
	RefundNo        string  `json:"refund_no"`
	RefundAmount    int64   `json:"refund_amount"`
	RefundReason    string  `json:"refund_reason"`
	RefundType      string  `json:"refund_type"`
	Status          string  `json:"status"`
	ApprovalLevel   string  `json:"approval_level"`
	ApprovedBy      *int64  `json:"approved_by,omitempty"`
	ApprovedAt      *string `json:"approved_at,omitempty"`
	CompletedAt     *string `json:"completed_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// OrderStatusLogResponse is a status log entry.
type OrderStatusLogResponse struct {
	ID           int64  `json:"id"`
	FromStatus   string `json:"from_status"`
	ToStatus     string `json:"to_status"`
	OperatorType string `json:"operator_type"`
	OperatorID   *int64 `json:"operator_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// PaginatedAdminOrdersResponse is the paginated order list.
type PaginatedAdminOrdersResponse struct {
	Items    []AdminOrderResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// --- Service Methods ---

// ListOrders returns a paginated admin order list with multi-dimension filters.
func (s *AdminOrderService) ListOrders(req AdminOrderListRequest) (*PaginatedAdminOrdersResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	query := s.db.Model(&ordermodel.MainOrder{})

	// Apply filters
	if req.OrderNo != "" {
		query = query.Where("order_no = ?", req.OrderNo)
	}
	if req.UserPhone != "" {
		// Search by contact_phone (plain text) since user phone is encrypted
		query = query.Where("contact_phone LIKE ?", "%"+req.UserPhone+"%")
	}
	if req.Status != "" && req.Status != "all" {
		query = query.Where("order_status = ?", req.Status)
	}
	if req.DateFrom != "" {
		query = query.Where("created_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		// Include the entire end day
		query = query.Where("created_at < ?", req.DateTo+" 23:59:59")
	}
	if req.ProductType != "" {
		query = query.Where("product_id IN (SELECT id FROM product WHERE product_type = ?)", req.ProductType)
	}
	if req.SupplierID != nil {
		query = query.Where("product_id IN (SELECT id FROM product WHERE supplier_id = ?)", *req.SupplierID)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	// Paginate
	offset := (req.Page - 1) * req.PageSize
	var orders []ordermodel.MainOrder
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("find orders: %w", err)
	}

	// Collect product IDs for batch lookup
	productIDs := make(map[int64]bool)
	for _, o := range orders {
		productIDs[o.ProductID] = true
	}

	// Batch load products
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
	items := make([]AdminOrderResponse, len(orders))
	for i, o := range orders {
		prodName := ""
		if p, ok := productsMap[o.ProductID]; ok {
			prodName = p.ProductName
		}

		items[i] = AdminOrderResponse{
			ID:            o.ID,
			OrderNo:       o.OrderNo,
			UserPhone:     maskPhone(o.ContactPhone),
			ProductName:   prodName,
			OrderStatus:   o.OrderStatus,
			PayableAmount: o.PayableAmount,
			AdultCount:    o.AdultCount,
			ChildCount:    o.ChildCount,
			InfantCount:   o.InfantCount,
			Channel:       o.Channel,
			ContactName:   o.ContactName,
			CreatedAt:     o.CreatedAt.Format(time.RFC3339),
		}
	}

	return &PaginatedAdminOrdersResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetOrderDetail returns the full order detail with all relations.
func (s *AdminOrderService) GetOrderDetail(orderID int64) (*AdminOrderDetailResponse, error) {
	var order ordermodel.MainOrder
	err := s.db.
		Preload("Travellers", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("StatusLogs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&order, orderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound // reuse error
		}
		return nil, fmt.Errorf("find order: %w", err)
	}

	// Load product name
	var product productmodel.Product
	s.db.Select("product_name").First(&product, order.ProductID)

	// Load payment transactions
	var payments []paymentmodel.PaymentTransaction
	s.db.Where("order_id = ?", orderID).Order("created_at ASC").Find(&payments)

	// Load refund records
	var refunds []paymentmodel.RefundRecord
	s.db.Where("order_id = ?", orderID).Order("created_at ASC").Find(&refunds)

	// Load user phone (for masking)
	var user usermodel.UserAccount
	s.db.Select("phone").First(&user, order.UserID)

	// Build travellers
	travellers := make([]OrderTravellerResponse, len(order.Travellers))
	for i, t := range order.Travellers {
		var bd *string
		if t.BirthDate != nil {
			s := t.BirthDate.Format("2006-01-02")
			bd = &s
		}
		travellers[i] = OrderTravellerResponse{
			ID:        t.ID,
			Phone:     t.Phone,
			Gender:    t.Gender,
			IsChild:   t.IsChild,
			IsInfant:  t.IsInfant,
			BirthDate: bd,
		}
	}

	// Build payments
	paymentResps := make([]OrderPaymentResponse, len(payments))
	for i, p := range payments {
		var paidAt *string
		if p.PaidAt != nil {
			s := p.PaidAt.Format(time.RFC3339)
			paidAt = &s
		}
		paymentResps[i] = OrderPaymentResponse{
			ID:             p.ID,
			PaymentNo:      p.PaymentNo,
			Channel:        p.Channel,
			Method:         p.Method,
			Amount:         p.Amount,
			Status:         p.Status,
			ChannelTradeNo: p.ChannelTradeNo,
			PaidAt:         paidAt,
			CreatedAt:      p.CreatedAt.Format(time.RFC3339),
		}
	}

	// Build refunds
	refundResps := make([]OrderRefundResponse, len(refunds))
	for i, r := range refunds {
		var approvedAt, completedAt *string
		if r.ApprovedAt != nil {
			s := r.ApprovedAt.Format(time.RFC3339)
			approvedAt = &s
		}
		if r.CompletedAt != nil {
			s := r.CompletedAt.Format(time.RFC3339)
			completedAt = &s
		}
		refundResps[i] = OrderRefundResponse{
			ID:           r.ID,
			RefundNo:     r.RefundNo,
			RefundAmount: r.RefundAmount,
			RefundReason: r.RefundReason,
			RefundType:   r.RefundType,
			Status:       r.Status,
			ApprovalLevel: r.ApprovalLevel,
			ApprovedBy:   r.ApprovedBy,
			ApprovedAt:   approvedAt,
			CompletedAt:  completedAt,
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		}
	}

	// Build status logs
	logResps := make([]OrderStatusLogResponse, len(order.StatusLogs))
	for i, l := range order.StatusLogs {
		logResps[i] = OrderStatusLogResponse{
			ID:           l.ID,
			FromStatus:   l.FromStatus,
			ToStatus:     l.ToStatus,
			OperatorType: l.OperatorType,
			OperatorID:   l.OperatorID,
			Reason:       l.Reason,
			CreatedAt:    l.CreatedAt.Format(time.RFC3339),
		}
	}

	// Build timestamps
	var paidAt, cancelledAt, completedAt *string
	if order.PaidAt != nil {
		s := order.PaidAt.Format(time.RFC3339)
		paidAt = &s
	}
	if order.CancelledAt != nil {
		s := order.CancelledAt.Format(time.RFC3339)
		cancelledAt = &s
	}
	if order.CompletedAt != nil {
		s := order.CompletedAt.Format(time.RFC3339)
		completedAt = &s
	}

	return &AdminOrderDetailResponse{
		AdminOrderResponse: AdminOrderResponse{
			ID:            order.ID,
			OrderNo:       order.OrderNo,
			UserPhone:     maskPhone(user.Phone),
			ProductName:   product.ProductName,
			OrderStatus:   order.OrderStatus,
			PayableAmount: order.PayableAmount,
			AdultCount:    order.AdultCount,
			ChildCount:    order.ChildCount,
			InfantCount:   order.InfantCount,
			Channel:       order.Channel,
			ContactName:   order.ContactName,
			CreatedAt:     order.CreatedAt.Format(time.RFC3339),
		},
		TotalAmount:            order.TotalAmount,
		DiscountAmount:         order.DiscountAmount,
		SingleSupplementAmount: order.SingleSupplementAmount,
		AddonAmount:            order.AddonAmount,
		ContactPhone:           maskPhone(order.ContactPhone),
		Remark:                 order.Remark,
		PaymentStatus:          order.PaymentStatus,
		CancelReason:           order.CancelReason,
		PaidAt:                 paidAt,
		CancelledAt:            cancelledAt,
		CompletedAt:            completedAt,
		Travellers:             travellers,
		Payments:               paymentResps,
		Refunds:                refundResps,
		StatusLogs:             logResps,
	}, nil
}

// maskPhone masks a phone number: 138****8000.
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
