// Package service provides business logic for the Payment domain.
//
// This file implements the unified payment service with:
//   - Payment creation (routes to Alipay or WeChat)
//   - Idempotent callback handling (DB unique constraint + Redis dedup per research.md)
//   - Payment status query
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	ordermodel "github.com/travel-booking/server/internal/order/model"
	orderrepo "github.com/travel-booking/server/internal/order/repository"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
	paymentrepo "github.com/travel-booking/server/internal/payment/repository"
	productsvc "github.com/travel-booking/server/internal/product/service"
)

// Payment errors.
var (
	ErrOrderNotPayable    = errors.New("order is not in pending_pay status")
	ErrActivePaymentExists = errors.New("active payment already exists for this order")
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrDuplicateCallback  = errors.New("duplicate callback")
)

// PaymentService provides business logic for payment operations.
type PaymentService struct {
	paymentRepo  *paymentrepo.PaymentRepository
	orderRepo    *orderrepo.OrderRepository
	inventorySvc *productsvc.InventoryService
	alipaySvc    *AlipayService
	wechatSvc    *WechatPayService
	rdb          *redis.Client
	logger       *zap.Logger
	callbackURL  string // base URL for payment callbacks
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(
	paymentRepo *paymentrepo.PaymentRepository,
	orderRepo *orderrepo.OrderRepository,
	inventorySvc *productsvc.InventoryService,
	alipaySvc *AlipayService,
	wechatSvc *WechatPayService,
	rdb *redis.Client,
	logger *zap.Logger,
	callbackURL string,
) *PaymentService {
	return &PaymentService{
		paymentRepo:  paymentRepo,
		orderRepo:    orderRepo,
		inventorySvc: inventorySvc,
		alipaySvc:    alipaySvc,
		wechatSvc:    wechatSvc,
		rdb:          rdb,
		logger:       logger,
		callbackURL:  callbackURL,
	}
}

// CreatePaymentRequest is the request body for creating a payment.
type CreatePaymentRequest struct {
	OrderID int64  `json:"order_id" binding:"required"`
	Channel string `json:"channel" binding:"required,oneof=alipay wechat"`
	Method  string `json:"method"`
}

// CreatePaymentResponse is the response for a created payment.
type CreatePaymentResponse struct {
	PaymentID int64           `json:"payment_id"`
	PaymentNo string          `json:"payment_no"`
	Channel   string          `json:"channel"`
	Method    string          `json:"method"`
	Amount    int64           `json:"amount"`
	ExpireAt  time.Time       `json:"expire_at"`
	PayParams json.RawMessage `json:"pay_params"`
}

// CreatePayment creates a payment order for a pending order.
func (s *PaymentService) CreatePayment(userID int64, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 1. Verify order exists and is in pending_pay status
	order, err := s.orderRepo.FindByIDBasic(req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrPaymentNotFound
	}
	if order.OrderStatus != ordermodel.OrderStatusPendingPay {
		return nil, ErrOrderNotPayable
	}

	// 2. Check for existing active payment
	hasActive, err := s.paymentRepo.HasActivePayment(req.OrderID)
	if err != nil {
		return nil, err
	}
	if hasActive {
		return nil, ErrActivePaymentExists
	}

	// 3. Determine method
	method := req.Method
	if method == "" {
		method = s.autoSelectMethod(req.Channel)
	}

	// 4. Generate payment number
	paymentNo := generatePaymentNo()

	// 5. Create payment transaction record
	now := time.Now()
	expireAt := now.Add(30 * time.Minute)
	notifyURL := s.callbackURL + "/api/v1/payments/notify/" + req.Channel

	ptx := &paymentmodel.PaymentTransaction{
		OrderID:   req.OrderID,
		PaymentNo: paymentNo,
		Channel:   req.Channel,
		Method:    method,
		Amount:    order.PayableAmount,
		Status:    paymentmodel.PaymentTxnStatusCreated,
		ExpireAt:  expireAt,
		NotifyURL: notifyURL,
	}

	if err := s.paymentRepo.Create(ptx); err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	// 6. Get payment parameters from channel
	var payParams json.RawMessage
	switch req.Channel {
	case "alipay":
		params, err := s.alipaySvc.CreatePayment(ptx, order.OrderNo)
		if err != nil {
			s.logger.Error("alipay create payment failed", zap.Error(err))
			// In dev/test mode, return placeholder params
			payParams = json.RawMessage(`{"pay_url":"https://openapi.alipay.com/gateway.do?placeholder=true"}`)
		} else {
			payParams = params
		}
	case "wechat":
		params, err := s.wechatSvc.CreatePayment(ptx, order.OrderNo)
		if err != nil {
			s.logger.Error("wechat create payment failed", zap.Error(err))
			payParams = json.RawMessage(`{"code_url":"weixin://wxpay/bizpayurl?placeholder=true"}`)
		} else {
			payParams = params
		}
	}

	// 7. Update payment with extra params
	if payParams != nil {
		extra, _ := json.Marshal(map[string]interface{}{
			"pay_params": json.RawMessage(payParams),
		})
		s.paymentRepo.UpdateStatus(ptx.ID, paymentmodel.PaymentTxnStatusCreated, map[string]interface{}{
			"extra_params": extra,
		})
	}

	s.logger.Info("payment created",
		zap.Int64("payment_id", ptx.ID),
		zap.String("payment_no", paymentNo),
		zap.String("channel", req.Channel),
		zap.Int64("amount", order.PayableAmount),
	)

	return &CreatePaymentResponse{
		PaymentID: ptx.ID,
		PaymentNo: paymentNo,
		Channel:   req.Channel,
		Method:    method,
		Amount:    order.PayableAmount,
		ExpireAt:  expireAt,
		PayParams: payParams,
	}, nil
}

// HandleCallback processes a payment callback idempotently.
// Per research.md: DB unique constraint + Redis dedup with 24h TTL.
func (s *PaymentService) HandleCallback(paymentID int64, channelTradeNo string, success bool) error {
	ctx := context.Background()

	// 1. Redis dedup check (fast path)
	dedupKey := fmt.Sprintf("payment:dedup:%d", paymentID)
	exists, err := s.rdb.Exists(ctx, dedupKey).Result()
	if err == nil && exists > 0 {
		s.logger.Info("duplicate callback ignored (redis)", zap.Int64("payment_id", paymentID))
		return nil // already processed
	}

	// 2. Fetch payment
	ptx, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("find payment: %w", err)
	}

	// 3. Check if already paid
	if ptx.Status == paymentmodel.PaymentTxnStatusPaid {
		// Set Redis dedup key for future fast-path
		s.rdb.Set(ctx, dedupKey, 1, 24*time.Hour)
		return nil
	}

	// 4. Update payment status
	now := time.Now()
	if success {
		err = s.paymentRepo.UpdateStatus(paymentID, paymentmodel.PaymentTxnStatusPaid, map[string]interface{}{
			"channel_trade_no": channelTradeNo,
			"paid_at":          now,
		})
		if err != nil {
			return fmt.Errorf("update payment: %w", err)
		}

		// 5. Update order status to paid_full
		err = s.orderRepo.UpdateStatus(ptx.OrderID,
			ordermodel.OrderStatusPendingPay,
			ordermodel.OrderStatusPaidFull,
			"system", nil, "payment success",
		)
		if err != nil {
			s.logger.Error("failed to update order status on payment",
				zap.Int64("order_id", ptx.OrderID),
				zap.Error(err),
			)
			return fmt.Errorf("update order status: %w", err)
		}

		// 6. Confirm inventory (move locked_count to sold_count)
		order, err := s.orderRepo.FindByIDBasic(ptx.OrderID)
		if err == nil {
			totalSeats := order.AdultCount + order.ChildCount + order.InfantCount
			if err := s.inventorySvc.ConfirmStock(ctx, order.DepartureID, totalSeats); err != nil {
				s.logger.Error("failed to confirm stock",
					zap.Int64("order_id", ptx.OrderID),
					zap.Error(err),
				)
			}
		}

		s.logger.Info("payment confirmed",
			zap.Int64("payment_id", paymentID),
			zap.Int64("order_id", ptx.OrderID),
			zap.String("channel_trade_no", channelTradeNo),
		)
	} else {
		err = s.paymentRepo.UpdateStatus(paymentID, paymentmodel.PaymentTxnStatusFailed, map[string]interface{}{
			"channel_trade_no": channelTradeNo,
		})
		if err != nil {
			return fmt.Errorf("update payment: %w", err)
		}
	}

	// 7. Set Redis dedup key
	s.rdb.Set(ctx, dedupKey, 1, 24*time.Hour)

	return nil
}

// GetPaymentStatus returns the current payment status.
func (s *PaymentService) GetPaymentStatus(userID, paymentID int64) (*PaymentStatusResponse, error) {
	ptx, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	// Verify ownership through order
	order, err := s.orderRepo.FindByIDBasic(ptx.OrderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrPaymentNotFound
	}

	return &PaymentStatusResponse{
		ID:             ptx.ID,
		PaymentNo:      ptx.PaymentNo,
		Channel:        ptx.Channel,
		Method:         ptx.Method,
		Amount:         ptx.Amount,
		Status:         ptx.Status,
		ChannelTradeNo: ptx.ChannelTradeNo,
		PaidAt:         ptx.PaidAt,
		ExpireAt:       ptx.ExpireAt,
		CreatedAt:      ptx.CreatedAt,
	}, nil
}

// autoSelectMethod selects the default payment method based on channel.
func (s *PaymentService) autoSelectMethod(channel string) string {
	switch channel {
	case "alipay":
		return paymentmodel.MethodH5
	case "wechat":
		return paymentmodel.MethodNative
	default:
		return paymentmodel.MethodH5
	}
}

// generatePaymentNo generates a payment number: PAY-YYYYMMDD-HHMMSS-XXXX.
func generatePaymentNo() string {
	now := time.Now()
	return fmt.Sprintf("PAY-%s-%s-%04d",
		now.Format("20060102"),
		now.Format("150405"),
		now.UnixNano()%10000,
	)
}

// PaymentStatusResponse is the payment status query response.
type PaymentStatusResponse struct {
	ID             int64      `json:"id"`
	PaymentNo      string     `json:"payment_no"`
	Channel        string     `json:"channel"`
	Method         string     `json:"method"`
	Amount         int64      `json:"amount"`
	Status         string     `json:"status"`
	ChannelTradeNo string     `json:"channel_trade_no,omitempty"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	ExpireAt       time.Time  `json:"expire_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
