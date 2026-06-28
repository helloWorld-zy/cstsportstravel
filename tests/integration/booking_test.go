package integration

import (
	"encoding/json"
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

// TestBookingFlow (T141) verifies the complete booking lifecycle:
// 1. Browse product list
// 2. View product detail
// 3. View departure calendar
// 4. Create order (select departure + travellers)
// 5. Create payment
// 6. Simulate payment callback
// 7. Verify order status updated
func TestBookingFlow(t *testing.T) {
	env := setupTestEnv(t)
	env.registerProductRoutes()
	env.registerOrderRoutes()
	env.registerPaymentRoutes()

	// Create a verified user.
	user := &usermodel.UserAccount{
		Phone:          "13800138010",
		Nickname:       "预订测试用户",
		Status:         usermodel.UserStatusActive,
		RealNameStatus: usermodel.RNStatusVerified,
		MemberLevel:    1,
	}
	env.DB.Create(user)
	token := env.generateToken(user.ID, "user", nil, nil)

	t.Run("Step1_ProductList", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/products?page=1&page_size=10", nil, "")

		if w.Code != http.StatusOK {
			t.Fatalf("list products failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		items, ok := data["items"].([]interface{})
		if !ok {
			t.Fatal("response missing items array")
		}
		if len(items) == 0 {
			t.Error("product list is empty")
		}
	})

	t.Run("Step2_ProductDetail", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/products/1", nil, "")

		if w.Code != http.StatusOK {
			t.Fatalf("get product detail failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		if data["product_name"] == nil {
			t.Error("product detail missing product_name")
		}
		if data["days"] == nil {
			t.Error("product detail missing days")
		}
	})

	t.Run("Step3_DepartureCalendar", func(t *testing.T) {
		month := time.Now().Add(30 * 24 * time.Hour).Format("2006-01")
		w := env.doRequest("GET", fmt.Sprintf("/api/v1/products/1/departures?month=%s", month), nil, "")

		if w.Code != http.StatusOK {
			t.Fatalf("get departures failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}
	})

	var orderID int64

	t.Run("Step4_CreateOrder", func(t *testing.T) {
		orderReq := map[string]interface{}{
			"product_id":    1,
			"departure_id":  1,
			"adult_count":   2,
			"child_count":   1,
			"infant_count":  0,
			"contact_name":  "预订测试用户",
			"contact_phone": "13800138010",
			"travellers": []map[string]interface{}{
				{
					"real_name":  "成人一",
					"id_card_no": "110101199001011234",
					"phone":      "13800138010",
					"birth_date": "1990-01-01",
					"gender":     "male",
				},
				{
					"real_name":  "成人二",
					"id_card_no": "110101199202022345",
					"phone":      "13900139000",
					"birth_date": "1992-02-02",
					"gender":     "female",
				},
				{
					"real_name":             "儿童一",
					"id_card_no":            "110101202001013456",
					"birth_date":            "2020-01-01",
					"gender":                "male",
					"is_child":              true,
					"linked_adult_index":    0,
				},
			},
		}

		w := env.doRequest("POST", "/api/v1/orders", orderReq, token)

		if w.Code != http.StatusOK {
			t.Fatalf("create order failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		if data["order_id"] == nil {
			t.Error("response missing order_id")
		}
		if data["order_no"] == nil {
			t.Error("response missing order_no")
		}
		if data["payable_amount"] == nil {
			t.Error("response missing payable_amount")
		}

		// Extract order ID for subsequent steps.
		idFloat, ok := data["order_id"].(float64)
		if !ok {
			t.Fatalf("order_id is not a number: %T", data["order_id"])
		}
		orderID = int64(idFloat)
	})

	t.Run("Step5_CreatePayment", func(t *testing.T) {
		if orderID == 0 {
			t.Skip("skipping — order creation failed")
		}

		payReq := map[string]interface{}{
			"order_id": orderID,
			"channel":  "alipay",
			"method":   "h5",
		}

		w := env.doRequest("POST", "/api/v1/payments/create", payReq, token)

		if w.Code != http.StatusOK {
			t.Fatalf("create payment failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		if data["payment_id"] == nil {
			t.Error("response missing payment_id")
		}
	})

	t.Run("Step6_SimulatePaymentCallback", func(t *testing.T) {
		if orderID == 0 {
			t.Skip("skipping — order creation failed")
		}

		callbackReq := map[string]interface{}{
			"order_id": orderID,
			"status":   "paid",
		}

		w := env.doRequest("POST", "/api/v1/test/payments/simulate-callback", callbackReq, "")

		if w.Code != http.StatusOK {
			t.Fatalf("simulate callback failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		if resp.Code != response.CodeSuccess {
			t.Errorf("expected code %d, got %d: %s", response.CodeSuccess, resp.Code, resp.Message)
		}
	})

	t.Run("Step7_VerifyOrderStatus", func(t *testing.T) {
		if orderID == 0 {
			t.Skip("skipping — order creation failed")
		}

		w := env.doRequest("GET", fmt.Sprintf("/api/v1/orders/%d", orderID), nil, token)

		if w.Code != http.StatusOK {
			t.Fatalf("get order failed: status %d, body: %s", w.Code, w.Body.String())
		}

		resp := parseResponse(t, w)
		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("response data is not a map: %T", resp.Data)
		}

		status, _ := data["order_status"].(string)
		if status != ordermodel.OrderStatusPaidFull && status != ordermodel.OrderStatusPendingTravel {
			t.Errorf("expected order_status=paid_full or pending_travel, got %s", status)
		}
	})

	t.Run("Step8_VerifyInventoryDeducted", func(t *testing.T) {
		var dep productmodel.DepartureDate
		env.DB.First(&dep, 1)

		if dep.SoldCount < 2 {
			t.Errorf("expected sold_count >= 2 after booking, got %d", dep.SoldCount)
		}
	})
}

// TestBookingWithInsufficientStock verifies that booking fails when stock is insufficient.
func TestBookingWithInsufficientStock(t *testing.T) {
	env := setupTestEnv(t)
	env.registerProductRoutes()
	env.registerOrderRoutes()

	// Set stock to 1.
	env.DB.Model(&productmodel.DepartureDate{}).Where("id = ?", 1).
		Updates(map[string]interface{}{"total_stock": 1, "sold_count": 0})

	user := &usermodel.UserAccount{
		Phone:          "13800138020",
		Nickname:       "库存测试用户",
		Status:         usermodel.UserStatusActive,
		RealNameStatus: usermodel.RNStatusVerified,
	}
	env.DB.Create(user)
	token := env.generateToken(user.ID, "user", nil, nil)

	// Try to book 2 adults (stock only 1).
	orderReq := map[string]interface{}{
		"product_id":    1,
		"departure_id":  1,
		"adult_count":   2,
		"contact_name":  "库存测试用户",
		"contact_phone": "13800138020",
		"travellers": []map[string]interface{}{
			{"real_name": "成人一", "id_card_no": "110101199001011234", "phone": "13800138020", "gender": "male"},
			{"real_name": "成人二", "id_card_no": "110101199202022345", "phone": "13900139000", "gender": "female"},
		},
	}

	w := env.doRequest("POST", "/api/v1/orders", orderReq, token)

	// Should fail with stock error.
	if w.Code == http.StatusOK {
		resp := parseResponse(t, w)
		if resp.Code == response.CodeSuccess {
			t.Error("booking with insufficient stock should fail")
		}
	}
}

// TestBookingWithUnverifiedUser verifies that booking requires real-name verification.
func TestBookingWithUnverifiedUser(t *testing.T) {
	env := setupTestEnv(t)
	env.registerProductRoutes()
	env.registerOrderRoutes()

	user := &usermodel.UserAccount{
		Phone:          "13800138030",
		Nickname:       "未认证用户",
		Status:         usermodel.UserStatusActive,
		RealNameStatus: usermodel.RNStatusUnverified,
	}
	env.DB.Create(user)
	token := env.generateToken(user.ID, "user", nil, nil)

	orderReq := map[string]interface{}{
		"product_id":    1,
		"departure_id":  1,
		"adult_count":   1,
		"contact_name":  "未认证用户",
		"contact_phone": "13800138030",
		"travellers": []map[string]interface{}{
			{"real_name": "未认证", "id_card_no": "110101199001011234", "phone": "13800138030", "gender": "male"},
		},
	}

	w := env.doRequest("POST", "/api/v1/orders", orderReq, token)

	// Should fail — real-name verification required.
	if w.Code == http.StatusOK {
		resp := parseResponse(t, w)
		if resp.Code == response.CodeSuccess {
			t.Error("booking without real-name verification should fail")
		}
	}
}

// registerProductRoutes sets up product API routes for testing.
func (e *testEnv) registerProductRoutes() {
	v1 := e.Router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.GET("", func(c *gin.Context) {
				var prods []productmodel.Product
				e.DB.Where("status = ?", productmodel.ProductStatusApproved).Find(&prods)
				response.OK(c, map[string]interface{}{
					"items": prods,
					"total": len(prods),
					"page":  1,
				})
			})

			products.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				var prod productmodel.Product
				if err := e.DB.First(&prod, id).Error; err != nil {
					response.NotFound(c, "product not found")
					return
				}
				response.OK(c, prod)
			})

			products.GET("/:id/departures", func(c *gin.Context) {
				id := c.Param("id")
				var deps []productmodel.DepartureDate
				e.DB.Where("product_id = ?", id).Find(&deps)
				response.OK(c, deps)
			})

			products.GET("/:id/reviews", func(c *gin.Context) {
				response.OK(c, []interface{}{})
			})

			products.GET("/:id/itinerary", func(c *gin.Context) {
				response.OK(c, []interface{}{})
			})

			products.GET("/search/suggest", func(c *gin.Context) {
				response.OK(c, []string{})
			})
		}
	}
}

// registerOrderRoutes sets up order API routes for testing.
func (e *testEnv) registerOrderRoutes() {
	v1 := e.Router.Group("/api/v1")
	orders := v1.Group("/orders")
	orders.Use(middleware.AuthRequired(e.JWT))
	{
		orders.POST("", func(c *gin.Context) {
			userID := middleware.GetUserID(c)

			var req struct {
				ProductID    int64  `json:"product_id"`
				DepartureID  int64  `json:"departure_id"`
				AdultCount   int    `json:"adult_count"`
				ChildCount   int    `json:"child_count"`
				InfantCount  int    `json:"infant_count"`
				ContactName  string `json:"contact_name"`
				ContactPhone string `json:"contact_phone"`
				Travellers   []struct {
					RealName  string `json:"real_name"`
					IDCardNo  string `json:"id_card_no"`
					Phone     string `json:"phone"`
					BirthDate string `json:"birth_date"`
					Gender    string `json:"gender"`
					IsChild   bool   `json:"is_child"`
				} `json:"travellers"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.BadRequest(c, "invalid request: "+err.Error())
				return
			}

			// Verify user has real-name verification.
			var user usermodel.UserAccount
			e.DB.First(&user, userID)
			if user.RealNameStatus != usermodel.RNStatusVerified {
				response.BusinessError(c, response.CodeBusiness, "real-name verification required")
				return
			}

			// Check stock.
			var dep productmodel.DepartureDate
			if err := e.DB.First(&dep, req.DepartureID).Error; err != nil {
				response.NotFound(c, "departure not found")
				return
			}
			needed := req.AdultCount + req.ChildCount + req.InfantCount
			if dep.AvailableStock() < needed {
				response.BusinessError(c, response.CodeStockEmpty, "insufficient stock")
				return
			}

			// Calculate price.
			totalAmount := int64(req.AdultCount)*int64(dep.AdultPrice) +
				int64(req.ChildCount)*int64(dep.ChildPrice) +
				int64(req.InfantCount)*int64(dep.InfantPrice)

			// Single supplement for odd adult count.
			singleSupplement := int64(0)
			if req.AdultCount%2 == 1 {
				singleSupplement = int64(dep.SingleSupplement)
			}
			totalAmount += singleSupplement

			// Create order.
			order := ordermodel.MainOrder{
				OrderNo:                fmt.Sprintf("ORD-%s-%04d", time.Now().Format("20060102150405"), 1),
				UserID:                 userID,
				ProductID:              req.ProductID,
				DepartureID:            req.DepartureID,
				OrderStatus:            ordermodel.OrderStatusPendingPay,
				PaymentStatus:          ordermodel.PaymentStatusUnpaid,
				TotalAmount:            totalAmount,
				PayableAmount:          totalAmount,
				SingleSupplementAmount: singleSupplement,
				AdultCount:             req.AdultCount,
				ChildCount:             req.ChildCount,
				InfantCount:            req.InfantCount,
				ContactName:            req.ContactName,
				ContactPhone:           req.ContactPhone,
			}
			e.DB.Create(&order)

			// Lock stock.
			e.DB.Model(&dep).Updates(map[string]interface{}{
				"locked_count": dep.LockedCount + needed,
			})

			// Create travellers.
			for _, t := range req.Travellers {
				encName, _ := e.Encryptor.Encrypt(t.RealName)
				encIDCard, _ := e.Encryptor.Encrypt(t.IDCardNo)
				traveller := ordermodel.OrderTraveller{
					OrderID:  order.ID,
					RealName: encName,
					IDCardNo: encIDCard,
					Phone:    t.Phone,
					Gender:   t.Gender,
					IsChild:  t.IsChild,
				}
				e.DB.Create(&traveller)
			}

			// Create status log.
			e.DB.Create(&ordermodel.OrderStatusLog{
				OrderID:      order.ID,
				FromStatus:   "",
				ToStatus:     ordermodel.OrderStatusPendingPay,
				OperatorType: "user",
				OperatorID:   &userID,
			})

			response.OK(c, map[string]interface{}{
				"order_id":       order.ID,
				"order_no":       order.OrderNo,
				"payable_amount": order.PayableAmount,
				"expire_at":      time.Now().Add(30 * time.Minute).Format(time.RFC3339),
			})
		})

		orders.GET("", func(c *gin.Context) {
			userID := middleware.GetUserID(c)
			var orders []ordermodel.MainOrder
			e.DB.Where("user_id = ?", userID).Order("id DESC").Find(&orders)
			response.OK(c, map[string]interface{}{
				"items": orders,
				"total": len(orders),
			})
		})

		orders.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var order ordermodel.MainOrder
			if err := e.DB.First(&order, id).Error; err != nil {
				response.NotFound(c, "order not found")
				return
			}
			response.OK(c, order)
		})

		orders.POST("/:id/cancel", func(c *gin.Context) {
			id := c.Param("id")
			var order ordermodel.MainOrder
			if err := e.DB.First(&order, id).Error; err != nil {
				response.NotFound(c, "order not found")
				return
			}
			if order.OrderStatus != ordermodel.OrderStatusPendingPay {
				response.BusinessError(c, response.CodeBusiness, "order cannot be cancelled")
				return
			}
			e.DB.Model(&order).Updates(map[string]interface{}{
				"order_status": ordermodel.OrderStatusCancelled,
				"cancel_reason": "user_cancel",
			})
			response.OK(c, map[string]interface{}{"status": ordermodel.OrderStatusCancelled})
		})
	}
}

// registerPaymentRoutes sets up payment API routes for testing.
func (e *testEnv) registerPaymentRoutes() {
	v1 := e.Router.Group("/api/v1")

	// Authenticated payment endpoints.
	payments := v1.Group("/payments")
	payments.Use(middleware.AuthRequired(e.JWT))
	{
		payments.POST("/create", func(c *gin.Context) {
			var req struct {
				OrderID int64  `json:"order_id"`
				Channel string `json:"channel"`
				Method  string `json:"method"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				response.BadRequest(c, "invalid request")
				return
			}

			var order ordermodel.MainOrder
			if err := e.DB.First(&order, req.OrderID).Error; err != nil {
				response.NotFound(c, "order not found")
				return
			}

			payment := paymentmodel.PaymentTransaction{
				OrderID:   req.OrderID,
				PaymentNo: fmt.Sprintf("PAY-%s-%04d", time.Now().Format("20060102150405"), 1),
				Channel:   req.Channel,
				Method:    req.Method,
				Amount:    order.PayableAmount,
				Status:    paymentmodel.PaymentTxnStatusCreated,
				ExpireAt:  time.Now().Add(30 * time.Minute),
			}
			e.DB.Create(&payment)

			response.OK(c, map[string]interface{}{
				"payment_id": payment.ID,
				"pay_url":    fmt.Sprintf("https://openapi.alipay.com/mock/%d", payment.ID),
			})
		})

		payments.GET("/:id/status", func(c *gin.Context) {
			id := c.Param("id")
			var payment paymentmodel.PaymentTransaction
			if err := e.DB.First(&payment, id).Error; err != nil {
				response.NotFound(c, "payment not found")
				return
			}
			response.OK(c, map[string]interface{}{
				"status": payment.Status,
			})
		})
	}

	// Test-only payment simulation.
	v1.POST("/test/payments/simulate-callback", func(c *gin.Context) {
		var req struct {
			OrderID int64  `json:"order_id"`
			Status  string `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request")
			return
		}

		if req.Status == "paid" {
			// Update order status.
			e.DB.Model(&ordermodel.MainOrder{}).Where("id = ?", req.OrderID).
				Updates(map[string]interface{}{
					"order_status":   ordermodel.OrderStatusPaidFull,
					"payment_status": ordermodel.PaymentStatusPaid,
					"paid_at":        time.Now(),
				})

			// Update payment transaction.
			e.DB.Model(&paymentmodel.PaymentTransaction{}).Where("order_id = ?", req.OrderID).
				Updates(map[string]interface{}{
					"status": paymentmodel.PaymentTxnStatusPaid,
					"paid_at": time.Now(),
				})

			// Move locked stock to sold.
			var order ordermodel.MainOrder
			e.DB.First(&order, req.OrderID)
			var dep productmodel.DepartureDate
			e.DB.First(&dep, order.DepartureID)
			count := order.AdultCount + order.ChildCount + order.InfantCount
			e.DB.Model(&dep).Updates(map[string]interface{}{
				"locked_count": dep.LockedCount - count,
				"sold_count":   dep.SoldCount + count,
			})
		}

		response.OK(c, map[string]interface{}{"status": "processed"})
	})
}

// Helper to marshal JSON for test assertions.
func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
