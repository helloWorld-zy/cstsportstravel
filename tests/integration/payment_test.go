package integration

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/travel-booking/server/internal/common/response"
	ordermodel "github.com/travel-booking/server/internal/order/model"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
	usermodel "github.com/travel-booking/server/internal/user/model"
)

// TestPaymentCallbackIdempotency (T143) verifies that:
// 1. A payment callback is processed correctly the first time
// 2. Duplicate callbacks for the same payment are idempotent (no double-processing)
// 3. Concurrent callbacks for the same payment are handled safely
func TestPaymentCallbackIdempotency(t *testing.T) {
	env := setupTestEnv(t)
	env.registerIdempotentPaymentRoutes()

	// Create a verified user and a pending-pay order.
	user := &usermodel.UserAccount{
		Phone:          "13800138060",
		Nickname:       "幂等测试用户",
		Status:         usermodel.UserStatusActive,
		RealNameStatus: usermodel.RNStatusVerified,
	}
	env.DB.Create(user)

	var orderID int64
	var paymentID int64

	t.Run("Setup_CreatePendingOrder", func(t *testing.T) {
		order := ordermodel.MainOrder{
			OrderNo:       fmt.Sprintf("ORD-IDEMPOTENT-%04d", 1),
			UserID:        user.ID,
			ProductID:     1,
			DepartureID:   1,
			OrderStatus:   ordermodel.OrderStatusPendingPay,
			PaymentStatus: ordermodel.PaymentStatusUnpaid,
			TotalAmount:   399900,
			PayableAmount: 399900,
			AdultCount:    1,
			ContactName:   "幂等测试用户",
			ContactPhone:  "13800138060",
		}
		env.DB.Create(&order)
		orderID = order.ID

		payment := paymentmodel.PaymentTransaction{
			OrderID:   order.ID,
			PaymentNo: fmt.Sprintf("PAY-IDEMPOTENT-%04d", 1),
			Channel:   paymentmodel.ChannelAlipay,
			Method:    paymentmodel.MethodH5,
			Amount:    order.PayableAmount,
			Status:    paymentmodel.PaymentTxnStatusCreated,
			ExpireAt:  time.Now().Add(30 * time.Minute),
		}
		env.DB.Create(&payment)
		paymentID = payment.ID
	})

	t.Run("Step1_FirstCallbackProcessed", func(t *testing.T) {
		if paymentID == 0 {
			t.Skip("skipping — setup failed")
		}

		callbackReq := map[string]interface{}{
			"payment_id":      paymentID,
			"channel_trade_no": "ALIPAY-TRADE-001",
			"status":          "paid",
		}

		w := env.doRequest("POST", "/api/v1/test/payment-callback", callbackReq, "")

		if w.Code != http.StatusOK {
			t.Fatalf("first callback failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		// Verify order status updated.
		var order ordermodel.MainOrder
		env.DB.First(&order, orderID)
		if order.OrderStatus != ordermodel.OrderStatusPaidFull {
			t.Errorf("expected order_status=paid_full after first callback, got %s", order.OrderStatus)
		}

		// Verify payment status updated.
		var payment paymentmodel.PaymentTransaction
		env.DB.First(&payment, paymentID)
		if payment.Status != paymentmodel.PaymentTxnStatusPaid {
			t.Errorf("expected payment status=paid after first callback, got %s", payment.Status)
		}
	})

	t.Run("Step2_DuplicateCallbackIsIdempotent", func(t *testing.T) {
		if paymentID == 0 {
			t.Skip("skipping — setup failed")
		}

		// Send the same callback again.
		callbackReq := map[string]interface{}{
			"payment_id":      paymentID,
			"channel_trade_no": "ALIPAY-TRADE-001",
			"status":          "paid",
		}

		w := env.doRequest("POST", "/api/v1/test/payment-callback", callbackReq, "")

		// Should succeed (idempotent) without error.
		if w.Code != http.StatusOK {
			t.Fatalf("duplicate callback failed: status %d, body: %s", w.Code, w.Body.String())
		}

		// Verify order status is still paid_full (not double-processed).
		var order ordermodel.MainOrder
		env.DB.First(&order, orderID)
		if order.OrderStatus != ordermodel.OrderStatusPaidFull {
			t.Errorf("expected order_status=paid_full after duplicate callback, got %s", order.OrderStatus)
		}

		// Verify exactly one status log for the payment.
		var logCount int64
		env.DB.Model(&ordermodel.OrderStatusLog{}).
			Where("order_id = ? AND to_status = ?", orderID, ordermodel.OrderStatusPaidFull).
			Count(&logCount)
		if logCount != 1 {
			t.Errorf("expected exactly 1 payment status log, got %d", logCount)
		}
	})

	t.Run("Step3_ConcurrentCallbacksHandledSafely", func(t *testing.T) {
		// Create a new order for concurrency testing.
		order := ordermodel.MainOrder{
			OrderNo:       fmt.Sprintf("ORD-CONCURRENT-%04d", 1),
			UserID:        user.ID,
			ProductID:     1,
			DepartureID:   1,
			OrderStatus:   ordermodel.OrderStatusPendingPay,
			PaymentStatus: ordermodel.PaymentStatusUnpaid,
			TotalAmount:   399900,
			PayableAmount: 399900,
			AdultCount:    1,
			ContactName:   "并发测试",
			ContactPhone:  "13800138060",
		}
		env.DB.Create(&order)

		payment := paymentmodel.PaymentTransaction{
			OrderID:   order.ID,
			PaymentNo: fmt.Sprintf("PAY-CONCURRENT-%04d", 1),
			Channel:   paymentmodel.ChannelAlipay,
			Method:    paymentmodel.MethodH5,
			Amount:    order.PayableAmount,
			Status:    paymentmodel.PaymentTxnStatusCreated,
			ExpireAt:  time.Now().Add(30 * time.Minute),
		}
		env.DB.Create(&payment)

		// Send 5 concurrent callbacks for the same payment.
		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				callbackReq := map[string]interface{}{
					"payment_id":       payment.ID,
					"channel_trade_no": fmt.Sprintf("ALIPAY-TRADE-CONCURRENT-%d", idx),
					"status":           "paid",
				}
				env.doRequest("POST", "/api/v1/test/payment-callback", callbackReq, "")
			}(i)
		}
		wg.Wait()

		// Verify order is paid exactly once.
		var updatedOrder ordermodel.MainOrder
		env.DB.First(&updatedOrder, order.ID)
		if updatedOrder.OrderStatus != ordermodel.OrderStatusPaidFull {
			t.Errorf("expected order_status=paid_full after concurrent callbacks, got %s", updatedOrder.OrderStatus)
		}

		// Verify no duplicate status transitions.
		var logCount int64
		env.DB.Model(&ordermodel.OrderStatusLog{}).
			Where("order_id = ? AND to_status = ?", order.ID, ordermodel.OrderStatusPaidFull).
			Count(&logCount)
		if logCount != 1 {
			t.Errorf("expected exactly 1 payment status log after concurrent callbacks, got %d", logCount)
		}
	})
}

// TestPaymentCallbackWithInvalidData verifies that malformed callbacks are rejected.
func TestPaymentCallbackWithInvalidData(t *testing.T) {
	env := setupTestEnv(t)
	env.registerIdempotentPaymentRoutes()

	t.Run("MissingPaymentID", func(t *testing.T) {
		w := env.doRequest("POST", "/api/v1/test/payment-callback",
			map[string]interface{}{"status": "paid"}, "")

		if w.Code == http.StatusOK {
			resp := parseResponse(t, w)
			if resp.Code == response.CodeSuccess {
				t.Error("callback without payment_id should fail")
			}
		}
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		// Create a payment.
		env.DB.Create(&paymentmodel.PaymentTransaction{
			OrderID:   999,
			PaymentNo: "PAY-INVALID-001",
			Channel:   paymentmodel.ChannelAlipay,
			Method:    paymentmodel.MethodH5,
			Amount:    100000,
			Status:    paymentmodel.PaymentTxnStatusCreated,
			ExpireAt:  time.Now().Add(30 * time.Minute),
		})

		w := env.doRequest("POST", "/api/v1/test/payment-callback",
			map[string]interface{}{
				"payment_id": 999,
				"status":     "invalid_status",
			}, "")

		if w.Code == http.StatusOK {
			resp := parseResponse(t, w)
			if resp.Code == response.CodeSuccess {
				t.Error("callback with invalid status should fail")
			}
		}
	})
}

// TestPaymentCallbackForCancelledOrder verifies that callbacks for cancelled orders are rejected.
func TestPaymentCallbackForCancelledOrder(t *testing.T) {
	env := setupTestEnv(t)
	env.registerIdempotentPaymentRoutes()

	// Create a cancelled order.
	env.DB.Create(&ordermodel.MainOrder{
		OrderNo:       "ORD-CANCELLED-001",
		UserID:        1,
		ProductID:     1,
		DepartureID:   1,
		OrderStatus:   ordermodel.OrderStatusCancelled,
		PaymentStatus: ordermodel.PaymentStatusUnpaid,
		TotalAmount:   100000,
		PayableAmount: 100000,
		AdultCount:    1,
		ContactName:   "取消测试",
		ContactPhone:  "13800138000",
		CancelReason:  "user_cancel",
	})

	payment := paymentmodel.PaymentTransaction{
		OrderID:   998,
		PaymentNo: "PAY-CANCELLED-001",
		Channel:   paymentmodel.ChannelAlipay,
		Method:    paymentmodel.MethodH5,
		Amount:    100000,
		Status:    paymentmodel.PaymentTxnStatusClosed,
		ExpireAt:  time.Now().Add(30 * time.Minute),
	}
	env.DB.Create(&payment)

	callbackReq := map[string]interface{}{
		"payment_id":       payment.ID,
		"channel_trade_no": "ALIPAY-LATE-CALLBACK",
		"status":           "paid",
	}

	w := env.doRequest("POST", "/api/v1/test/payment-callback", callbackReq, "")

	// Should reject — payment already closed.
	if w.Code == http.StatusOK {
		resp := parseResponse(t, w)
		if resp.Code == response.CodeSuccess {
			t.Error("callback for closed payment should fail")
		}
	}
}

// registerIdempotentPaymentRoutes sets up payment callback routes with idempotency checks.
func (e *testEnv) registerIdempotentPaymentRoutes() {
	// Track processed payments for idempotency.
	processedPayments := &sync.Map{}

	v1 := e.Router.Group("/api/v1")
	v1.POST("/test/payment-callback", func(c *gin.Context) {
		var req struct {
			PaymentID      int64  `json:"payment_id"`
			ChannelTradeNo string `json:"channel_trade_no"`
			Status         string `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request")
			return
		}

		if req.PaymentID == 0 {
			response.BadRequest(c, "payment_id is required")
			return
		}

		// Check if already processed (idempotency).
		if _, loaded := processedPayments.LoadOrStore(req.PaymentID, true); loaded {
			// Already processed — return success (idempotent).
			response.OK(c, map[string]interface{}{"status": "already_processed"})
			return
		}

		// Load payment.
		var payment paymentmodel.PaymentTransaction
		if err := e.DB.First(&payment, req.PaymentID).Error; err != nil {
			processedPayments.Delete(req.PaymentID)
			response.NotFound(c, "payment not found")
			return
		}

		// Reject if payment is already closed or paid.
		if payment.Status == paymentmodel.PaymentTxnStatusClosed ||
			payment.Status == paymentmodel.PaymentTxnStatusPaid {
			processedPayments.Delete(req.PaymentID)
			response.BusinessError(c, response.CodeBusiness, "payment already processed")
			return
		}

		// Validate status.
		if req.Status != "paid" && req.Status != "failed" {
			processedPayments.Delete(req.PaymentID)
			response.BadRequest(c, "invalid status")
			return
		}

		// Load order.
		var order ordermodel.MainOrder
		if err := e.DB.First(&order, payment.OrderID).Error; err != nil {
			processedPayments.Delete(req.PaymentID)
			response.NotFound(c, "order not found")
			return
		}

		// Reject if order is cancelled.
		if order.OrderStatus == ordermodel.OrderStatusCancelled {
			processedPayments.Delete(req.PaymentID)
			response.BusinessError(c, response.CodeBusiness, "order is cancelled")
			return
		}

		// Process payment (atomic update with optimistic locking).
		now := time.Now()
		result := e.DB.Model(&paymentmodel.PaymentTransaction{}).
			Where("id = ? AND status = ?", req.PaymentID, paymentmodel.PaymentTxnStatusCreated).
			Updates(map[string]interface{}{
				"status":          paymentmodel.PaymentTxnStatusPaid,
				"channel_trade_no": req.ChannelTradeNo,
				"paid_at":         now,
			})

		if result.RowsAffected == 0 {
			// Another goroutine already processed this payment.
			processedPayments.Delete(req.PaymentID)
			response.OK(c, map[string]interface{}{"status": "already_processed"})
			return
		}

		// Update order status.
		e.DB.Model(&ordermodel.MainOrder{}).Where("id = ?", order.ID).
			Updates(map[string]interface{}{
				"order_status":   ordermodel.OrderStatusPaidFull,
				"payment_status": ordermodel.PaymentStatusPaid,
				"paid_at":        now,
			})

		// Create status log.
		e.DB.Create(&ordermodel.OrderStatusLog{
			OrderID:      order.ID,
			FromStatus:   ordermodel.OrderStatusPendingPay,
			ToStatus:     ordermodel.OrderStatusPaidFull,
			OperatorType: "system",
			Reason:       "payment callback",
		})

		// Move locked stock to sold.
		var dep struct{ LockedCount, SoldCount int }
		e.DB.Table("departure_date").Select("locked_count, sold_count").
			Where("id = ?", order.DepartureID).Scan(&dep)
		count := order.AdultCount + order.ChildCount + order.InfantCount
		e.DB.Exec("UPDATE departure_date SET locked_count = locked_count - ?, sold_count = sold_count + ? WHERE id = ?",
			count, count, order.DepartureID)

		response.OK(c, map[string]interface{}{
			"status":      "processed",
			"order_status": ordermodel.OrderStatusPaidFull,
		})
	})
}
