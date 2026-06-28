package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	ordermodel "github.com/travel-booking/server/internal/order/model"
	paymentmodel "github.com/travel-booking/server/internal/payment/model"
	productmodel "github.com/travel-booking/server/internal/product/model"
	usermodel "github.com/travel-booking/server/internal/user/model"
)

// TestRefundFlow (T142) verifies the complete refund lifecycle:
// 1. Create a paid order
// 2. User requests refund
// 3. System calculates refund amount based on cancellation rules
// 4. Admin approves refund
// 5. Order status transitions to refunded
func TestRefundFlow(t *testing.T) {
	env := setupTestEnv(t)
	env.registerProductRoutes()
	env.registerOrderRoutes()
	env.registerPaymentRoutes()
	env.registerRefundRoutes()
	env.registerAdminRefundRoutes()

	// Create a verified user.
	user := &usermodel.UserAccount{
		Phone:          "13800138040",
		Nickname:       "退款测试用户",
		Status:         usermodel.UserStatusActive,
		RealNameStatus: usermodel.RNStatusVerified,
		MemberLevel:    1,
	}
	env.DB.Create(user)
	token := env.generateToken(user.ID, "user", nil, nil)

	// Create admin user with refund approval permission.
	adminToken := env.generateToken(1, "admin", []string{"super_admin"}, []string{"refund:approve", "refund:list"})

	// Create a paid order.
	var orderID int64
	t.Run("Setup_CreatePaidOrder", func(t *testing.T) {
		order := ordermodel.MainOrder{
			OrderNo:       fmt.Sprintf("ORD-REFUND-%04d", 1),
			UserID:        user.ID,
			ProductID:     1,
			DepartureID:   1,
			OrderStatus:   ordermodel.OrderStatusPaidFull,
			PaymentStatus: ordermodel.PaymentStatusPaid,
			TotalAmount:   799800,
			PayableAmount: 799800,
			AdultCount:    2,
			ContactName:   "退款测试用户",
			ContactPhone:  "13800138040",
			PaidAt:        timePtr(time.Now()),
		}
		env.DB.Create(&order)
		orderID = order.ID

		// Create payment record.
		env.DB.Create(&paymentmodel.PaymentTransaction{
			OrderID:   order.ID,
			PaymentNo: fmt.Sprintf("PAY-REFUND-%04d", 1),
			Channel:   paymentmodel.ChannelAlipay,
			Method:    paymentmodel.MethodH5,
			Amount:    order.PayableAmount,
			Status:    paymentmodel.PaymentTxnStatusPaid,
			PaidAt:    timePtr(time.Now()),
			ExpireAt:  time.Now().Add(30 * time.Minute),
		})
	})

	t.Run("Step1_UserRequestsRefund", func(t *testing.T) {
		if orderID == 0 {
			t.Skip("skipping — setup failed")
		}

		refundReq := map[string]interface{}{
			"reason":      "行程变更",
			"description": "因个人原因无法出行",
		}

		w := env.doRequest("POST", fmt.Sprintf("/api/v1/orders/%d/refund", orderID), refundReq, token)

		if w.Code != http.StatusOK {
			t.Fatalf("request refund failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		if data["refund_id"] == nil {
			t.Error("response missing refund_id")
		}
		if data["refund_amount"] == nil {
			t.Error("response missing refund_amount")
		}
	})

	t.Run("Step2_VerifyRefundAmountCalculation", func(t *testing.T) {
		if orderID == 0 {
			t.Skip("skipping — setup failed")
		}

		// The departure is 30 days away, so the refund should be 100%.
		var refund paymentmodel.RefundRecord
		env.DB.Where("order_id = ?", orderID).First(&refund)

		if refund.ID == 0 {
			t.Fatal("refund record not created")
		}

		// 30 days before departure → 100% refund.
		expectedRefund := int64(799800)
		if refund.RefundAmount != expectedRefund {
			t.Errorf("expected refund amount %d, got %d", expectedRefund, refund.RefundAmount)
		}
	})

	t.Run("Step3_OrderStatusIsRefunding", func(t *testing.T) {
		if orderID == 0 {
			t.Skip("skipping — setup failed")
		}

		var order ordermodel.MainOrder
		env.DB.First(&order, orderID)

		if order.OrderStatus != ordermodel.OrderStatusRefunding {
			t.Errorf("expected order_status=refunding, got %s", order.OrderStatus)
		}
	})

	t.Run("Step4_AdminViewsPendingRefunds", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/admin/refunds?status=pending", nil, adminToken)

		if w.Code != http.StatusOK {
			t.Fatalf("list refunds failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}
	})

	t.Run("Step5_AdminApprovesRefund", func(t *testing.T) {
		var refund paymentmodel.RefundRecord
		env.DB.Where("order_id = ?", orderID).First(&refund)

		if refund.ID == 0 {
			t.Skip("skipping — refund record not found")
		}

		approveReq := map[string]interface{}{
			"note": "审核通过",
		}

		w := env.doRequest("PUT", fmt.Sprintf("/api/v1/admin/refunds/%d/approve", refund.ID), approveReq, adminToken)

		if w.Code != http.StatusOK {
			t.Fatalf("approve refund failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}
	})

	t.Run("Step6_OrderStatusIsRefunded", func(t *testing.T) {
		if orderID == 0 {
			t.Skip("skipping — setup failed")
		}

		var order ordermodel.MainOrder
		env.DB.First(&order, orderID)

		if order.OrderStatus != ordermodel.OrderStatusRefunded {
			t.Errorf("expected order_status=refunded, got %s", order.OrderStatus)
		}
	})
}

// TestRefundCancellationRuleCalculation verifies that the cancellation rule
// engine correctly calculates refund amounts based on days before departure.
func TestRefundCancellationRuleCalculation(t *testing.T) {
	env := setupTestEnv(t)

	testCases := []struct {
		name              string
		daysBeforeDeparture int
		expectedPercentage float64
	}{
		{"30 days before → 100%", 30, 100},
		{"15 days before → 100%", 15, 100},
		{"10 days before → 80%", 10, 80},
		{"5 days before → 50%", 5, 50},
		{"1 day before → 0%", 1, 0},
		{"0 days before → 0%", 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Find the matching refund rule.
			var rule productmodel.RefundRule
			err := env.DB.Where("is_template = ? AND days_before_min <= ? AND (days_before_max IS NULL OR days_before_max >= ?)",
				true, tc.daysBeforeDeparture, tc.daysBeforeDeparture).
				Order("days_before_min DESC").
				First(&rule).Error

			if err != nil {
				t.Fatalf("find refund rule: %v", err)
			}

			if rule.RefundPercentage != tc.expectedPercentage {
				t.Errorf("expected refund percentage %.0f%%, got %.0f%%",
					tc.expectedPercentage, rule.RefundPercentage)
			}
		})
	}
}

// TestRefundTieredApproval verifies the tiered approval logic:
// ≤1000: operator can approve directly
// 1000-5000: finance supervisor required
// >5000: director required
func TestRefundTieredApproval(t *testing.T) {
	testCases := []struct {
		name          string
		refundAmount  int64
		expectedLevel string
	}{
		{"≤1000 yuan — operator", 100000, "operator"},
		{"3000 yuan — finance_director", 300000, "finance_director"},
		{"6000 yuan — director", 600000, "director"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := determineApprovalLevel(tc.refundAmount)
			if level != tc.expectedLevel {
				t.Errorf("expected approval level %s, got %s", tc.expectedLevel, level)
			}
		})
	}
}

// TestRefundRejectFlow verifies that a rejected refund restores the order status.
func TestRefundRejectFlow(t *testing.T) {
	env := setupTestEnv(t)
	env.registerRefundRoutes()
	env.registerAdminRefundRoutes()

	user := &usermodel.UserAccount{
		Phone:          "13800138050",
		Nickname:       "退款拒绝测试",
		Status:         usermodel.UserStatusActive,
		RealNameStatus: usermodel.RNStatusVerified,
	}
	env.DB.Create(user)
	token := env.generateToken(user.ID, "user", nil, nil)
	adminToken := env.generateToken(1, "admin", []string{"super_admin"}, []string{"refund:approve"})

	// Create a paid order.
	order := ordermodel.MainOrder{
		OrderNo:       fmt.Sprintf("ORD-REJECT-%04d", 1),
		UserID:        user.ID,
		ProductID:     1,
		DepartureID:   1,
		OrderStatus:   ordermodel.OrderStatusPaidFull,
		PaymentStatus: ordermodel.PaymentStatusPaid,
		TotalAmount:   399900,
		PayableAmount: 399900,
		AdultCount:    1,
		ContactName:   "退款拒绝测试",
		ContactPhone:  "13800138050",
		PaidAt:        timePtr(time.Now()),
	}
	env.DB.Create(&order)

	// User requests refund.
	refundReq := map[string]interface{}{
		"reason":      "行程变更",
		"description": "因个人原因",
	}
	env.doRequest("POST", fmt.Sprintf("/api/v1/orders/%d/refund", order.ID), refundReq, token)

	// Admin rejects refund.
	var refund paymentmodel.RefundRecord
	env.DB.Where("order_id = ?", order.ID).First(&refund)

	rejectReq := map[string]interface{}{
		"reason": "退款申请不符合条件",
	}
	w := env.doRequest("PUT", fmt.Sprintf("/api/v1/admin/refunds/%d/reject", refund.ID), rejectReq, adminToken)

	if w.Code != http.StatusOK {
		t.Fatalf("reject refund failed: status %d, body: %s", w.Code, w.Body.String())
	}

	// Verify order status restored to paid_full.
	env.DB.First(&order, order.ID)
	if order.OrderStatus != ordermodel.OrderStatusPaidFull {
		t.Errorf("expected order_status=paid_full after rejection, got %s", order.OrderStatus)
	}
}

// registerRefundRoutes sets up user refund API routes for testing.
func (e *testEnv) registerRefundRoutes() {
	v1 := e.Router.Group("/api/v1")
	orders := v1.Group("/orders")
	orders.Use(middleware.AuthRequired(e.JWT))
	{
		orders.POST("/:id/refund", func(c *gin.Context) {
			orderID := c.Param("id")
			var order ordermodel.MainOrder
			if err := e.DB.First(&order, orderID).Error; err != nil {
				response.NotFound(c, "order not found")
				return
			}

			if order.OrderStatus != ordermodel.OrderStatusPaidFull &&
				order.OrderStatus != ordermodel.OrderStatusPendingTravel {
				response.BusinessError(c, response.CodeBusiness, "order cannot be refunded")
				return
			}

			// Calculate refund amount based on cancellation rules.
			var dep productmodel.DepartureDate
			e.DB.First(&dep, order.DepartureID)
			daysUntilDeparture := int(time.Until(dep.DepartureDate).Hours() / 24)
			if daysUntilDeparture < 0 {
				daysUntilDeparture = 0
			}

			var rule productmodel.RefundRule
			e.DB.Where("is_template = ? AND days_before_min <= ? AND (days_before_max IS NULL OR days_before_max >= ?)",
				true, daysUntilDeparture, daysUntilDeparture).
				Order("days_before_min DESC").
				First(&rule)

			refundAmount := int64(float64(order.PayableAmount) * rule.RefundPercentage / 100)

			// Determine approval level.
			approvalLevel := determineApprovalLevel(refundAmount)

			// Create refund record.
			refund := paymentmodel.RefundRecord{
				OrderID:      order.ID,
				PaymentID:    1, // simplified
				RefundNo:     fmt.Sprintf("REF-%s-%04d", time.Now().Format("20060102150405"), 1),
				RefundAmount: refundAmount,
				RefundReason: "行程变更",
				RefundType:   paymentmodel.RefundTypeFull,
				Status:       paymentmodel.RefundStatusPending,
				ApprovalLevel: approvalLevel,
			}
			e.DB.Create(&refund)

			// Update order status.
			e.DB.Model(&order).Update("order_status", ordermodel.OrderStatusRefunding)

			response.OK(c, map[string]interface{}{
				"refund_id":     refund.ID,
				"refund_amount": refund.RefundAmount,
				"status":        refund.Status,
			})
		})

		orders.GET("/:id/refund-status", func(c *gin.Context) {
			orderID := c.Param("id")
			var refund paymentmodel.RefundRecord
			if err := e.DB.Where("order_id = ?", orderID).First(&refund).Error; err != nil {
				response.NotFound(c, "refund not found")
				return
			}
			response.OK(c, refund)
		})
	}
}

// registerAdminRefundRoutes sets up admin refund management routes for testing.
func (e *testEnv) registerAdminRefundRoutes() {
	v1 := e.Router.Group("/api/v1")
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(e.JWT))
	{
		refunds := admin.Group("/refunds")
		{
			refunds.GET("", func(c *gin.Context) {
				status := c.Query("status")
				var refunds []paymentmodel.RefundRecord
				if status != "" {
					e.DB.Where("status = ?", status).Find(&refunds)
				} else {
					e.DB.Find(&refunds)
				}
				response.OK(c, map[string]interface{}{
					"items": refunds,
					"total": len(refunds),
				})
			})

			refunds.PUT("/:id/approve", func(c *gin.Context) {
				refundID := c.Param("id")
				var refund paymentmodel.RefundRecord
				if err := e.DB.First(&refund, refundID).Error; err != nil {
					response.NotFound(c, "refund not found")
					return
				}

				// Update refund status.
				now := time.Now()
				e.DB.Model(&refund).Updates(map[string]interface{}{
					"status":      paymentmodel.RefundStatusApproved,
					"approved_at": now,
				})

				// Update order status to refunded.
				e.DB.Model(&ordermodel.MainOrder{}).Where("id = ?", refund.OrderID).
					Updates(map[string]interface{}{
						"order_status":   ordermodel.OrderStatusRefunded,
						"payment_status": ordermodel.PaymentStatusRefunded,
					})

				response.OK(c, map[string]interface{}{"status": paymentmodel.RefundStatusApproved})
			})

			refunds.PUT("/:id/reject", func(c *gin.Context) {
				refundID := c.Param("id")
				var refund paymentmodel.RefundRecord
				if err := e.DB.First(&refund, refundID).Error; err != nil {
					response.NotFound(c, "refund not found")
					return
				}

				// Update refund status.
				e.DB.Model(&refund).Update("status", paymentmodel.RefundStatusFailed)

				// Restore order status.
				e.DB.Model(&ordermodel.MainOrder{}).Where("id = ?", refund.OrderID).
					Update("order_status", ordermodel.OrderStatusPaidFull)

				response.OK(c, map[string]interface{}{"status": "rejected"})
			})
		}
	}
}

// determineApprovalLevel returns the required approval level based on refund amount.
func determineApprovalLevel(amountCents int64) string {
	amountYuan := float64(amountCents) / 100
	switch {
	case amountYuan <= 1000:
		return "operator"
	case amountYuan <= 5000:
		return "finance_director"
	default:
		return "director"
	}
}

func timePtr(t time.Time) *time.Time { return &t }
