// Package service provides business logic for the Order domain.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/common/encrypt"
	ordermodel "github.com/travel-booking/server/internal/order/model"
	orderrepo "github.com/travel-booking/server/internal/order/repository"
	productmodel "github.com/travel-booking/server/internal/product/model"
	productrepo "github.com/travel-booking/server/internal/product/repository"
	productsvc "github.com/travel-booking/server/internal/product/service"
	userrepo "github.com/travel-booking/server/internal/user/repository"
)

// Order errors.
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderNotCancellable = errors.New("order cannot be cancelled")
	ErrNotRealNameVerified = errors.New("user must complete real-name verification before booking")
	ErrInvalidIDCard       = errors.New("invalid ID card format")
)

// OrderService provides business logic for order creation, cancellation, and queries.
type OrderService struct {
	orderRepo    *orderrepo.OrderRepository
	productRepo  *productrepo.ProductRepository
	userRepo     *userrepo.UserRepository
	inventorySvc *productsvc.InventoryService
	encryptor    *encrypt.Encryptor
	logger       *zap.Logger
}

// NewOrderService creates a new OrderService.
func NewOrderService(
	orderRepo *orderrepo.OrderRepository,
	productRepo *productrepo.ProductRepository,
	userRepo *userrepo.UserRepository,
	inventorySvc *productsvc.InventoryService,
	encryptor *encrypt.Encryptor,
	logger *zap.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		productRepo:  productRepo,
		userRepo:     userRepo,
		inventorySvc: inventorySvc,
		encryptor:    encryptor,
		logger:       logger,
	}
}

// CreateOrderRequest is the request body for creating an order.
type CreateOrderRequest struct {
	ProductID   int64            `json:"product_id" binding:"required"`
	DepartureID int64            `json:"departure_id" binding:"required"`
	AdultCount  int              `json:"adult_count" binding:"required,min=1"`
	ChildCount  int              `json:"child_count"`
	InfantCount int              `json:"infant_count"`
	Travellers  []TravellerInput `json:"travellers" binding:"required,min=1"`
	Addons      []AddonInput     `json:"addons"`
	ContactName string           `json:"contact_name" binding:"required"`
	ContactPhone string          `json:"contact_phone" binding:"required"`
	Remark      string           `json:"remark"`
}

// AddonInput represents an optional addon service.
type AddonInput struct {
	AddonID  int64 `json:"addon_id"`
	Quantity int   `json:"quantity"`
}

// CreateOrderResponse is the response for a created order.
type CreateOrderResponse struct {
	OrderID              int64          `json:"order_id"`
	OrderNo              string         `json:"order_no"`
	TotalAmount          int64          `json:"total_amount"`
	PayableAmount        int64          `json:"payable_amount"`
	SingleSupplementAmount int64        `json:"single_supplement_amount"`
	FeeBreakdown         *PriceBreakdown `json:"fee_breakdown"`
	ExpireAt             time.Time      `json:"expire_at"`
}

// CreateOrder creates a new booking order.
// Flow: validate user → validate product/departure → lock stock → calculate price → create order.
func (s *OrderService) CreateOrder(userID int64, req CreateOrderRequest, channel string) (*CreateOrderResponse, error) {
	// 1. Verify user real-name status
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user.RealNameStatus != "verified" {
		return nil, ErrNotRealNameVerified
	}

	// 2. Fetch departure with product info
	dep, err := s.getDepartureWithProduct(req.ProductID, req.DepartureID)
	if err != nil {
		return nil, err
	}

	// 3. Validate departure status and cutoff
	if dep.Status != productmodel.DepartureStatusOpen {
		return nil, productsvc.ErrDepartureNotOpen
	}
	cutoffDate := dep.DepartureDate.AddDate(0, 0, -dep.CutoffDays)
	if time.Now().After(cutoffDate) {
		return nil, productsvc.ErrBookingCutoff
	}

	// 4. Validate travellers
	validationErrs := ValidateTravellers(req.Travellers, req.AdultCount, req.ChildCount, req.InfantCount, dep.DepartureDate)
	if len(validationErrs) > 0 {
		return nil, fmt.Errorf("traveller validation failed: %s", validationErrs[0].Error())
	}

	// 5. Lock inventory
	totalSeats := req.AdultCount + req.ChildCount + req.InfantCount
	ctx := context.Background()
	if err := s.inventorySvc.LockStock(ctx, req.DepartureID, totalSeats); err != nil {
		return nil, err
	}

	// 6. Calculate price
	addonAmount := int64(0) // MVP: addons calculated separately
	priceBreakdown := CalculatePrice(
		int64(dep.AdultPrice),
		int64(dep.ChildPrice),
		int64(dep.InfantPrice),
		int64(dep.SingleSupplement),
		req.AdultCount,
		req.ChildCount,
		req.InfantCount,
		addonAmount,
	)

	// 7. Generate order number
	orderNo := generateOrderNo()

	// 8. Build order model
	order := &ordermodel.MainOrder{
		OrderNo:                orderNo,
		UserID:                 userID,
		ProductID:              req.ProductID,
		DepartureID:            req.DepartureID,
		OrderStatus:            ordermodel.OrderStatusPendingPay,
		PaymentStatus:          ordermodel.PaymentStatusUnpaid,
		TotalAmount:            priceBreakdown.TotalAmount,
		DiscountAmount:         priceBreakdown.DiscountAmount,
		PayableAmount:          priceBreakdown.PayableAmount,
		AdultCount:             req.AdultCount,
		ChildCount:             req.ChildCount,
		InfantCount:            req.InfantCount,
		SingleSupplementAmount: priceBreakdown.SupplementSubtotal,
		AddonAmount:            addonAmount,
		ContactName:            req.ContactName,
		ContactPhone:           req.ContactPhone,
		Channel:                channel,
		Remark:                 req.Remark,
	}

	// 9. Build travellers (encrypt sensitive fields)
	travellers := make([]ordermodel.OrderTraveller, len(req.Travellers))
	for i, t := range req.Travellers {
		encryptedName, err := s.encryptor.Encrypt(t.RealName)
		if err != nil {
			return nil, fmt.Errorf("encrypt traveller name: %w", err)
		}
		encryptedIDCard, err := s.encryptor.Encrypt(t.IDCardNo)
		if err != nil {
			return nil, fmt.Errorf("encrypt traveller id_card: %w", err)
		}

		travellers[i] = ordermodel.OrderTraveller{
			RealName: encryptedName,
			IDCardNo: encryptedIDCard,
			Phone:    t.Phone,
			Gender:   t.Gender,
			IsChild:  t.IsChild,
			IsInfant: t.IsInfant,
		}

		// Parse birth date
		if t.BirthDate != "" {
			bd, err := time.Parse("2006-01-02", t.BirthDate)
			if err == nil {
				travellers[i].BirthDate = &bd
			}
		}

		// Set linked adult ID (will be resolved after creation)
		if t.LinkedAdultTravellerIndex != nil {
			linkedID := int64(*t.LinkedAdultTravellerIndex + 1) // placeholder, resolved below
			travellers[i].LinkedAdultID = &linkedID
		}
	}

	// 10. Create initial status log
	statusLog := &ordermodel.OrderStatusLog{
		FromStatus:   "",
		ToStatus:     ordermodel.OrderStatusPendingPay,
		OperatorType: "user",
		Reason:       "order created",
	}

	// 11. Persist order
	if err := s.orderRepo.Create(order, travellers, statusLog); err != nil {
		// Release stock on failure
		s.inventorySvc.ReleaseStock(ctx, req.DepartureID, totalSeats)
		return nil, fmt.Errorf("create order: %w", err)
	}

	// 12. Resolve linked adult IDs (now that we have real IDs)
	s.resolveLinkedAdults(order.ID, travellers, req.Travellers)

	// 13. Calculate expiry time
	expireAt := order.CreatedAt.Add(30 * time.Minute)

	s.logger.Info("order created",
		zap.Int64("order_id", order.ID),
		zap.String("order_no", orderNo),
		zap.Int64("user_id", userID),
		zap.Int64("payable_amount", priceBreakdown.PayableAmount),
	)

	return &CreateOrderResponse{
		OrderID:                order.ID,
		OrderNo:                orderNo,
		TotalAmount:            priceBreakdown.TotalAmount,
		PayableAmount:          priceBreakdown.PayableAmount,
		SingleSupplementAmount: priceBreakdown.SupplementSubtotal,
		FeeBreakdown:           priceBreakdown,
		ExpireAt:               expireAt,
	}, nil
}

// resolveLinkedAdults updates linked_adult_id for children/infants after order creation.
func (s *OrderService) resolveLinkedAdults(orderID int64, travellers []ordermodel.OrderTraveller, inputs []TravellerInput) {
	// Build a map of adult index -> traveller ID
	adultIDs := make(map[int]int64)
	for i, t := range travellers {
		if !t.IsChild && !t.IsInfant {
			adultIDs[i] = t.ID
		}
	}

	// Update linked adults
	for i, input := range inputs {
		if input.LinkedAdultTravellerIndex != nil {
			adultIdx := *input.LinkedAdultTravellerIndex
			if adultID, ok := adultIDs[adultIdx]; ok && i < len(travellers) {
				travellers[i].LinkedAdultID = &adultID
				s.orderRepo.CreateStatusLog(&ordermodel.OrderStatusLog{
					OrderID:      orderID,
					FromStatus:   "",
					ToStatus:     "",
					OperatorType: "system",
					Reason:       fmt.Sprintf("linked traveller %d to adult %d", travellers[i].ID, adultID),
				})
			}
		}
	}
}

// GetOrderList returns a paginated order list for a user.
func (s *OrderService) GetOrderList(userID int64, status string, page, pageSize int) ([]OrderListItem, int64, error) {
	filter := orderrepo.OrderFilter{
		UserID:  userID,
		Status:  status,
		Page:    page,
		PerPage: pageSize,
	}

	orders, total, err := s.orderRepo.FindByUserID(filter)
	if err != nil {
		return nil, 0, err
	}

	items := make([]OrderListItem, len(orders))
	for i, o := range orders {
		items[i] = s.toOrderListItem(&o)
	}

	return items, total, nil
}

// GetOrderDetail returns the full order detail for a user.
func (s *OrderService) GetOrderDetail(userID, orderID int64) (*OrderDetailResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	// Verify ownership
	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	return s.toOrderDetailResponse(order), nil
}

// CancelOrder cancels a pending_pay order and releases inventory.
func (s *OrderService) CancelOrder(userID, orderID int64, reason string) error {
	order, err := s.orderRepo.FindByIDBasic(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}

	// Verify ownership
	if order.UserID != userID {
		return ErrOrderNotFound
	}

	// Only pending_pay orders can be cancelled
	if order.OrderStatus != ordermodel.OrderStatusPendingPay {
		return ErrOrderNotCancellable
	}

	// Update status
	if err := s.orderRepo.UpdateStatus(orderID,
		ordermodel.OrderStatusPendingPay,
		ordermodel.OrderStatusCancelled,
		"user", &userID, reason,
	); err != nil {
		return err
	}

	// Release inventory
	totalSeats := order.AdultCount + order.ChildCount + order.InfantCount
	ctx := context.Background()
	if err := s.inventorySvc.ReleaseStock(ctx, order.DepartureID, totalSeats); err != nil {
		s.logger.Error("failed to release stock on cancel",
			zap.Int64("order_id", orderID),
			zap.Error(err),
		)
	}

	s.logger.Info("order cancelled",
		zap.Int64("order_id", orderID),
		zap.String("reason", reason),
	)

	return nil
}

// --- Helper functions ---

// getDepartureWithProduct fetches a departure and verifies it belongs to the product.
func (s *OrderService) getDepartureWithProduct(productID, departureID int64) (*productmodel.DepartureDate, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product not found")
		}
		return nil, err
	}

	for _, dep := range product.DepartureDates {
		if dep.ID == departureID {
			return &dep, nil
		}
	}

	return nil, fmt.Errorf("departure not found for product")
}

// generateOrderNo generates an order number: ORD-YYYYMMDD-HHMMSS-XXXX.
func generateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("ORD-%s-%s-%04d",
		now.Format("20060102"),
		now.Format("150405"),
		now.UnixNano()%10000,
	)
}

// --- Response DTOs ---

// OrderListItem is a summary of an order for list view.
type OrderListItem struct {
	ID               int64   `json:"id"`
	OrderNo          string  `json:"order_no"`
	OrderStatus      string  `json:"order_status"`
	ProductID        int64   `json:"product_id"`
	ProductName      string  `json:"product_name"`
	CoverImage       string  `json:"cover_image"`
	Days             int     `json:"days"`
	AdultCount       int     `json:"adult_count"`
	ChildCount       int     `json:"child_count"`
	InfantCount      int     `json:"infant_count"`
	PayableAmount    int64   `json:"payable_amount"`
	CreatedAt        string  `json:"created_at"`
}

// OrderDetailResponse is the full order detail.
type OrderDetailResponse struct {
	ID                     int64                    `json:"id"`
	OrderNo                string                   `json:"order_no"`
	OrderStatus            string                   `json:"order_status"`
	PaymentStatus          string                   `json:"payment_status"`
	ProductID              int64                    `json:"product_id"`
	ProductName            string                   `json:"product_name"`
	CoverImage             string                   `json:"cover_image"`
	Days                   int                      `json:"days"`
	DepartureDate          string                   `json:"departure_date"`
	ReturnDate             string                   `json:"return_date"`
	AdultCount             int                      `json:"adult_count"`
	ChildCount             int                      `json:"child_count"`
	InfantCount            int                      `json:"infant_count"`
	TotalAmount            int64                    `json:"total_amount"`
	DiscountAmount         int64                    `json:"discount_amount"`
	PayableAmount          int64                    `json:"payable_amount"`
	SingleSupplementAmount int64                    `json:"single_supplement_amount"`
	AddonAmount            int64                    `json:"addon_amount"`
	ContactName            string                   `json:"contact_name"`
	ContactPhone           string                   `json:"contact_phone"`
	Travellers             []OrderTravellerResponse `json:"travellers"`
	StatusLogs             []OrderStatusLogResponse `json:"status_logs"`
	CreatedAt              string                   `json:"created_at"`
	PaidAt                 string                   `json:"paid_at,omitempty"`
	ExpireAt               string                   `json:"expire_at"`
	CancelReason           string                   `json:"cancel_reason,omitempty"`
}

// OrderTravellerResponse is the traveller info in order detail (with masked sensitive fields).
type OrderTravellerResponse struct {
	ID            int64  `json:"id"`
	RealName      string `json:"real_name"`
	IDCardNo      string `json:"id_card_no"`
	Phone         string `json:"phone"`
	BirthDate     string `json:"birth_date,omitempty"`
	Gender        string `json:"gender"`
	IsChild       bool   `json:"is_child"`
	IsInfant      bool   `json:"is_infant"`
	LinkedAdultID *int64 `json:"linked_adult_id,omitempty"`
}

// OrderStatusLogResponse is the status log entry.
type OrderStatusLogResponse struct {
	FromStatus   string `json:"from_status"`
	ToStatus     string `json:"to_status"`
	OperatorType string `json:"operator_type"`
	Reason       string `json:"reason"`
	CreatedAt    string `json:"created_at"`
}

// toOrderListItem converts a MainOrder to an OrderListItem.
func (s *OrderService) toOrderListItem(order *ordermodel.MainOrder) OrderListItem {
	return OrderListItem{
		ID:            order.ID,
		OrderNo:       order.OrderNo,
		OrderStatus:   order.OrderStatus,
		ProductID:     order.ProductID,
		AdultCount:    order.AdultCount,
		ChildCount:    order.ChildCount,
		InfantCount:   order.InfantCount,
		PayableAmount: order.PayableAmount,
		CreatedAt:     order.CreatedAt.Format(time.RFC3339),
	}
}

// toOrderDetailResponse converts a MainOrder to an OrderDetailResponse.
func (s *OrderService) toOrderDetailResponse(order *ordermodel.MainOrder) *OrderDetailResponse {
	resp := &OrderDetailResponse{
		ID:                     order.ID,
		OrderNo:                order.OrderNo,
		OrderStatus:            order.OrderStatus,
		PaymentStatus:          order.PaymentStatus,
		ProductID:              order.ProductID,
		AdultCount:             order.AdultCount,
		ChildCount:             order.ChildCount,
		InfantCount:            order.InfantCount,
		TotalAmount:            order.TotalAmount,
		DiscountAmount:         order.DiscountAmount,
		PayableAmount:          order.PayableAmount,
		SingleSupplementAmount: order.SingleSupplementAmount,
		AddonAmount:            order.AddonAmount,
		ContactName:            order.ContactName,
		ContactPhone:           order.ContactPhone,
		CreatedAt:              order.CreatedAt.Format(time.RFC3339),
		ExpireAt:               order.CreatedAt.Add(30 * time.Minute).Format(time.RFC3339),
		CancelReason:           order.CancelReason,
	}

	if order.PaidAt != nil {
		resp.PaidAt = order.PaidAt.Format(time.RFC3339)
	}

	// Convert travellers (decrypt and mask)
	resp.Travellers = make([]OrderTravellerResponse, len(order.Travellers))
	for i, t := range order.Travellers {
		tr := OrderTravellerResponse{
			ID:            t.ID,
			Phone:         t.Phone,
			Gender:        t.Gender,
			IsChild:       t.IsChild,
			IsInfant:      t.IsInfant,
			LinkedAdultID: t.LinkedAdultID,
		}

		// Decrypt and mask sensitive fields
		if s.encryptor != nil {
			if name, err := s.encryptor.Decrypt(t.RealName); err == nil {
				tr.RealName = maskName(name)
			}
			if idCard, err := s.encryptor.Decrypt(t.IDCardNo); err == nil {
				tr.IDCardNo = maskIDCard(idCard)
			}
		}

		if t.BirthDate != nil {
			tr.BirthDate = t.BirthDate.Format("2006-01-02")
		}

		resp.Travellers[i] = tr
	}

	// Convert status logs
	resp.StatusLogs = make([]OrderStatusLogResponse, len(order.StatusLogs))
	for i, l := range order.StatusLogs {
		resp.StatusLogs[i] = OrderStatusLogResponse{
			FromStatus:   l.FromStatus,
			ToStatus:     l.ToStatus,
			OperatorType: l.OperatorType,
			Reason:       l.Reason,
			CreatedAt:    l.CreatedAt.Format(time.RFC3339),
		}
	}

	return resp
}

// maskName masks a real name, showing only the surname.
func maskName(name string) string {
	runes := []rune(name)
	if len(runes) <= 1 {
		return name
	}
	return string(runes[0]) + "**"
}

// maskIDCard masks an ID card number.
func maskIDCard(idCard string) string {
	if len(idCard) < 14 {
		return "****"
	}
	return idCard[:6] + "********" + idCard[14:]
}
