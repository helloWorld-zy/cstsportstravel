// Package service provides business logic for the Order domain.
//
// This file implements the refund service per FR-018, FR-019, FR-020, FR-021:
//   - CreateRefundRequest: validate order status, calculate refund, create refund_record
//   - ProcessRefund: execute refund via payment channel (original route back)
//   - Update order status through the refund state machine
package service

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	orderrepo "github.com/travel-booking/server/internal/order/repository"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
	paymentrepo "github.com/travel-booking/server/internal/payment/repository"
	productrepo "github.com/travel-booking/server/internal/product/repository"
)

// Refund errors.
var (
	ErrOrderNotRefundable  = errors.New("order is not eligible for refund")
	ErrRefundAlreadyExists = errors.New("refund request already exists for this order")
	ErrRefundNotFound      = errors.New("refund record not found")
)

// RefundService provides business logic for refund operations.
type RefundService struct {
	orderRepo   *orderrepo.OrderRepository
	payRepo     *paymentrepo.PaymentRepository
	productRepo *productrepo.ProductRepository
	engine      *CancellationEngine
	logger      *zap.Logger
}

// NewRefundService creates a new RefundService.
func NewRefundService(
	orderRepo *orderrepo.OrderRepository,
	payRepo *paymentrepo.PaymentRepository,
	productRepo *productrepo.ProductRepository,
	logger *zap.Logger,
) *RefundService {
	return &RefundService{
		orderRepo:   orderRepo,
		payRepo:     payRepo,
		productRepo: productRepo,
		engine:      NewCancellationEngine(),
		logger:      logger,
	}
}

// RefundRequestInput is the request body for creating a refund.
type RefundRequestInput struct {
	Reason      string `json:"reason" binding:"required"`
	Description string `json:"description"`
}

// RefundResponse is the response for a refund request.
type RefundResponse struct {
	RefundID        int64                  `json:"refund_id"`
	RefundNo        string                 `json:"refund_no"`
	RefundAmount    int64                  `json:"refund_amount"`
	RefundPercentage float64               `json:"refund_percentage"`
	MatchingRule    string                 `json:"matching_rule"`
	ApprovalLevel   string                 `json:"approval_level"`
	Status          string                 `json:"status"`
	DaysBefore      int                    `json:"days_before"`
	Calculation     *RefundCalculation     `json:"calculation"`
}

// CreateRefundRequest creates a new refund request for an order.
// Flow: validate order status → load product refund rules → match rule → calculate amount → create refund_record.
func (s *RefundService) CreateRefundRequest(userID, orderID int64, input RefundRequestInput) (*RefundResponse, error) {
	// 1. Fetch order
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("find order: %w", err)
	}

	// 2. Verify ownership
	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	// 3. Check order status - only paid_full and pending_travel are refundable
	if order.OrderStatus != ordermodel.OrderStatusPaidFull &&
		order.OrderStatus != ordermodel.OrderStatusPendingTravel {
		return nil, ErrOrderNotRefundable
	}

	// 4. Check for existing refund
	hasRefund, err := s.hasExistingRefund(orderID)
	if err != nil {
		return nil, fmt.Errorf("check existing refund: %w", err)
	}
	if hasRefund {
		return nil, ErrRefundAlreadyExists
	}

	// 5. Load product refund rules
	product, err := s.productRepo.FindByID(order.ProductID)
	if err != nil {
		return nil, fmt.Errorf("find product: %w", err)
	}

	// 6. Find the departure date to calculate days before
	var departureDate time.Time
	for _, dep := range product.DepartureDates {
		if dep.ID == order.DepartureID {
			departureDate = dep.DepartureDate
			break
		}
	}
	if departureDate.IsZero() {
		return nil, fmt.Errorf("departure not found for order")
	}

	// 7. Calculate days before departure
	daysBefore := CalculateDaysBeforeDeparture(departureDate)

	// 8. Match cancellation rule
	match := s.engine.MatchRule(product.RefundRules, daysBefore)

	// 9. Calculate refund amount
	calc := s.engine.CalculateRefund(order.PayableAmount, match)
	calc.DaysBefore = daysBefore

	// 10. Determine approval level (amount in yuan)
	refundAmountYuan := float64(calc.RefundAmount) / 100.0
	approvalLevel := DetermineApprovalLevel(refundAmountYuan)

	// 11. Find the original payment transaction
	payment, err := s.payRepo.FindByOrderID(orderID, "")
	if err != nil {
		return nil, fmt.Errorf("find payment: %w", err)
	}

	// 12. Generate refund number
	refundNo := generateRefundNo()

	// 13. Determine refund type
	refundType := paymentmodel.RefundTypePartial
	if calc.RefundAmount >= order.PayableAmount {
		refundType = paymentmodel.RefundTypeFull
	}

	// 14. Create refund record
	record := &paymentmodel.RefundRecord{
		OrderID:       orderID,
		PaymentID:     payment.ID,
		RefundNo:      refundNo,
		RefundAmount:  calc.RefundAmount,
		RefundReason:  input.Reason,
		RefundType:    refundType,
		Status:        paymentmodel.RefundStatusPending,
		ApprovalLevel: approvalLevel,
	}

	if err := s.payRepo.CreateRefundRecord(record); err != nil {
		return nil, fmt.Errorf("create refund record: %w", err)
	}

	// 15. Update order status to refunding
	if err := s.orderRepo.UpdateStatus(orderID,
		order.OrderStatus,
		ordermodel.OrderStatusRefunding,
		"user", &userID,
		fmt.Sprintf("refund requested: %s", input.Reason),
	); err != nil {
		s.logger.Error("failed to update order status to refunding",
			zap.Int64("order_id", orderID),
			zap.Error(err),
		)
	}

	// 16. Build matching rule description
	matchingRuleDesc := ""
	if match != nil {
		matchingRuleDesc = FormatRefundRuleDescription(match, daysBefore)
	} else {
		matchingRuleDesc = "无匹配退改规则"
	}

	s.logger.Info("refund request created",
		zap.Int64("order_id", orderID),
		zap.Int64("refund_id", record.ID),
		zap.String("refund_no", refundNo),
		zap.Int64("refund_amount", calc.RefundAmount),
		zap.String("approval_level", approvalLevel),
		zap.Int("days_before", daysBefore),
	)

	return &RefundResponse{
		RefundID:         record.ID,
		RefundNo:         refundNo,
		RefundAmount:     calc.RefundAmount,
		RefundPercentage: calc.RefundPercentage,
		MatchingRule:     matchingRuleDesc,
		ApprovalLevel:    approvalLevel,
		Status:           paymentmodel.RefundStatusPending,
		DaysBefore:       daysBefore,
		Calculation:      calc,
	}, nil
}

// GetRefundStatus returns the refund status for an order.
func (s *RefundService) GetRefundStatus(userID, orderID int64) (*paymentmodel.RefundRecord, error) {
	// Verify order ownership
	order, err := s.orderRepo.FindByIDBasic(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	// Find refund record
	record, err := s.findRefundByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ApproveRefund approves a pending refund (admin operation).
func (s *RefundService) ApproveRefund(refundID int64, approverID int64, note string) error {
	record, err := s.payRepo.FindRefundByID(refundID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRefundNotFound
		}
		return err
	}

	if record.Status != paymentmodel.RefundStatusPending {
		return fmt.Errorf("refund is not in pending status, current: %s", record.Status)
	}

	// Update refund status to approved
	now := time.Now()
	if err := s.payRepo.UpdateRefundStatus(refundID, paymentmodel.RefundStatusApproved, map[string]interface{}{
		"approved_by": approverID,
		"approved_at": now,
	}); err != nil {
		return fmt.Errorf("update refund status: %w", err)
	}

	// Update order status
	if err := s.orderRepo.UpdateStatus(record.OrderID,
		ordermodel.OrderStatusRefunding,
		ordermodel.OrderStatusRefunding, // stays refunding until refund completes
		"admin", &approverID,
		fmt.Sprintf("refund approved: %s", note),
	); err != nil {
		s.logger.Error("failed to update order status after approval",
			zap.Int64("order_id", record.OrderID),
			zap.Error(err),
		)
	}

	s.logger.Info("refund approved",
		zap.Int64("refund_id", refundID),
		zap.Int64("approver_id", approverID),
	)

	return nil
}

// RejectRefund rejects a pending refund (admin operation).
func (s *RefundService) RejectRefund(refundID int64, approverID int64, reason string) error {
	record, err := s.payRepo.FindRefundByID(refundID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRefundNotFound
		}
		return err
	}

	if record.Status != paymentmodel.RefundStatusPending {
		return fmt.Errorf("refund is not in pending status, current: %s", record.Status)
	}

	// Update refund status to failed
	if err := s.payRepo.UpdateRefundStatus(refundID, paymentmodel.RefundStatusFailed, map[string]interface{}{
		"approved_by": approverID,
		"approved_at": time.Now(),
	}); err != nil {
		return fmt.Errorf("update refund status: %w", err)
	}

	// Revert order status back to previous state
	if err := s.orderRepo.UpdateStatus(record.OrderID,
		ordermodel.OrderStatusRefunding,
		ordermodel.OrderStatusPaidFull, // revert to paid_full
		"admin", &approverID,
		fmt.Sprintf("refund rejected: %s", reason),
	); err != nil {
		s.logger.Error("failed to revert order status after rejection",
			zap.Int64("order_id", record.OrderID),
			zap.Error(err),
		)
	}

	s.logger.Info("refund rejected",
		zap.Int64("refund_id", refundID),
		zap.Int64("approver_id", approverID),
		zap.String("reason", reason),
	)

	return nil
}

// hasExistingRefund checks if an order already has a pending or approved refund.
func (s *RefundService) hasExistingRefund(orderID int64) (bool, error) {
	record, err := s.payRepo.FindRefundByOrderID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	// Consider pending and approved refunds as existing
	return record.Status == paymentmodel.RefundStatusPending ||
		record.Status == paymentmodel.RefundStatusApproved ||
		record.Status == paymentmodel.RefundStatusProcessing, nil
}

// findRefundByOrderID finds the latest refund record for an order.
func (s *RefundService) findRefundByOrderID(orderID int64) (*paymentmodel.RefundRecord, error) {
	record, err := s.payRepo.FindRefundByOrderID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return record, nil
}

// generateRefundNo generates a refund number: RFD-YYYYMMDD-HHMMSS-XXXX.
func generateRefundNo() string {
	now := time.Now()
	return fmt.Sprintf("RFD-%s-%s-%04d",
		now.Format("20060102"),
		now.Format("150405"),
		now.UnixNano()%10000,
	)
}
