package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/travel-booking/server/internal/common/config"
	"github.com/travel-booking/server/internal/payment/gateway"
)

// --- T110: UnionPay Callback Handler Tests ---

func setupTestRouter(handler *UnionPayNotifyHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v2/payments/notify/unionpay", handler.HandleBackNotification)
	r.POST("/api/v2/payments/notify/unionpay/front", handler.HandleFrontNotification)
	return r
}

func TestHandleBackNotification_Success(t *testing.T) {
	gw := gateway.NewUnionPayGateway(gateway.TestUnionPayConfig(), nil)
	handler := NewUnionPayNotifyHandler(gw, &mockPaymentUpdater{}, nil)
	router := setupTestRouter(handler)

	body := "orderId=PAY-20260701-0001&txnAmt=50000&respCode=00&txnTime=20260701120000"
	req := httptest.NewRequest(http.MethodPost, "/api/v2/payments/notify/unionpay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleBackNotification_MissingOrderId(t *testing.T) {
	gw := gateway.NewUnionPayGateway(gateway.TestUnionPayConfig(), nil)
	handler := NewUnionPayNotifyHandler(gw, &mockPaymentUpdater{}, nil)
	router := setupTestRouter(handler)

	body := "txnAmt=50000&respCode=00"
	req := httptest.NewRequest(http.MethodPost, "/api/v2/payments/notify/unionpay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleBackNotification_VerifyFailure(t *testing.T) {
	// Create a gateway with empty config (not configured) to test stub mode
	// In stub mode, missing respCode causes verification to fail
	gw := gateway.NewUnionPayGateway(config.UnionPayConfig{}, nil)
	handler := NewUnionPayNotifyHandler(gw, &mockPaymentUpdater{}, nil)
	router := setupTestRouter(handler)

	// Missing respCode will cause VerifyNotification to return false
	body := "orderId=PAY-20260701-0001&txnAmt=50000"
	req := httptest.NewRequest(http.MethodPost, "/api/v2/payments/notify/unionpay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleBackNotification_Idempotency(t *testing.T) {
	gw := gateway.NewUnionPayGateway(gateway.TestUnionPayConfig(), nil)
	updater := &mockPaymentUpdater{}
	handler := NewUnionPayNotifyHandler(gw, updater, nil)
	router := setupTestRouter(handler)

	body := "orderId=PAY-20260701-0001&txnAmt=50000&respCode=00"

	// First call
	req1 := httptest.NewRequest(http.MethodPost, "/api/v2/payments/notify/unionpay", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second call (idempotent - should not re-process)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v2/payments/notify/unionpay", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Updater should only be called once
	assert.Equal(t, 1, updater.updateCount)
}

func TestHandleBackNotification_FailedTransaction(t *testing.T) {
	gw := gateway.NewUnionPayGateway(gateway.TestUnionPayConfig(), nil)
	handler := NewUnionPayNotifyHandler(gw, &mockPaymentUpdater{}, nil)
	router := setupTestRouter(handler)

	body := "orderId=PAY-20260701-0002&txnAmt=50000&respCode=05"
	req := httptest.NewRequest(http.MethodPost, "/api/v2/payments/notify/unionpay", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still return 200 to UnionPay even for failed transactions
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleFrontNotification_DisplayOnly(t *testing.T) {
	gw := gateway.NewUnionPayGateway(gateway.TestUnionPayConfig(), nil)
	updater := &mockPaymentUpdater{}
	handler := NewUnionPayNotifyHandler(gw, updater, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v2/payments/notify/unionpay/front", handler.HandleFrontNotification)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/payments/notify/unionpay/front?orderId=PAY-20260701-0001&respCode=00", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// FR-162: frontUrl is display-only, should NOT update payment status
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 0, updater.updateCount)
}

// --- Mock helpers ---

type mockPaymentUpdater struct {
	verifyFail  bool
	updateCount int
}

func (m *mockPaymentUpdater) UpdatePaymentStatus(paymentNo string, status string, channelTradeNo string) error {
	m.updateCount++
	return nil
}

func (m *mockPaymentUpdater) IsProcessed(paymentNo string) bool {
	return m.updateCount > 0
}
