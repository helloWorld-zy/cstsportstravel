package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestMetricsMiddleware_RegistersMetrics verifies that the middleware registers
// the expected Prometheus metrics (HTTP request duration, count, error rate).
func TestMetricsMiddleware_RegistersMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	mw := NewMetricsMiddleware(registry)

	// Verify the middleware is created
	assert.NotNil(t, mw)

	// Make a request to ensure metrics are populated (Prometheus Gather
	// only returns metric families with at least one observation).
	router := gin.New()
	router.Use(mw.Handler())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gather metrics to verify registration
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	registeredNames := make(map[string]bool)
	for _, mf := range metricFamilies {
		registeredNames[mf.GetName()] = true
	}

	// Expect HTTP request duration histogram
	assert.True(t, registeredNames["travel_http_request_duration_seconds"],
		"expected travel_http_request_duration_seconds to be registered")

	// Expect HTTP request total counter
	assert.True(t, registeredNames["travel_http_requests_total"],
		"expected travel_http_requests_total to be registered")
}

// TestMetricsMiddleware_RecordsRequest verifies that the middleware records
// HTTP request metrics (count and duration) for each request.
func TestMetricsMiddleware_RecordsRequest(t *testing.T) {
	registry := prometheus.NewRegistry()
	mw := NewMetricsMiddleware(registry)

	router := gin.New()
	router.Use(mw.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	var requestCount float64
	for _, mf := range metricFamilies {
		if mf.GetName() == "travel_http_requests_total" {
			for _, m := range mf.GetMetric() {
				requestCount += m.GetCounter().GetValue()
			}
		}
	}

	assert.Equal(t, float64(1), requestCount,
		"expected exactly 1 request to be recorded")
}

// TestMetricsMiddleware_RecordsErrorMetrics verifies that the middleware
// increments error counters for 5xx responses.
func TestMetricsMiddleware_RecordsErrorMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	mw := NewMetricsMiddleware(registry)

	router := gin.New()
	router.Use(mw.Handler())
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fail"})
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	var errorCount float64
	for _, mf := range metricFamilies {
		if mf.GetName() == "travel_http_request_errors_total" {
			for _, m := range mf.GetMetric() {
				errorCount += m.GetCounter().GetValue()
			}
		}
	}

	assert.Equal(t, float64(1), errorCount,
		"expected exactly 1 error to be recorded")
}

// TestMetricsMiddleware_MetricsEndpoint verifies that the /metrics endpoint
// returns Prometheus text format output.
func TestMetricsMiddleware_MetricsEndpoint(t *testing.T) {
	registry := prometheus.NewRegistry()
	mw := NewMetricsMiddleware(registry)

	router := gin.New()
	router.Use(mw.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// Make a request to generate metrics
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Now request /metrics
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "travel_http_requests_total",
		"expected metrics endpoint to return prometheus format")
}

// TestMetricsMiddleware_LabelsByMethodAndStatus verifies that metrics are
// labeled by HTTP method, path, and status code.
func TestMetricsMiddleware_LabelsByMethodAndStatus(t *testing.T) {
	registry := prometheus.NewRegistry()
	mw := NewMetricsMiddleware(registry)

	router := gin.New()
	router.Use(mw.Handler())
	router.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	router.POST("/created", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{})
	})

	// GET /ok
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// POST /created
	req = httptest.NewRequest(http.MethodPost, "/created", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if mf.GetName() == "travel_http_requests_total" {
			// Should have at least 2 metric vectors (different method+status combos)
			assert.GreaterOrEqual(t, len(mf.GetMetric()), 2,
				"expected at least 2 labeled metric vectors")
		}
	}
}

// TestBusinessMetrics_RegisterAndIncrement verifies business metrics
// (order count, payment success rate) can be registered and incremented.
func TestBusinessMetrics_RegisterAndIncrement(t *testing.T) {
	registry := prometheus.NewRegistry()
	bm := NewBusinessMetrics(registry)

	assert.NotNil(t, bm)

	// Increment order count
	bm.IncOrderCount("domestic_tour")
	bm.IncOrderCount("domestic_tour")

	// Record payment success/failure
	bm.RecordPaymentResult("alipay", true)
	bm.RecordPaymentResult("alipay", true)
	bm.RecordPaymentResult("wechat", false)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	metrics := make(map[string]float64)
	for _, mf := range metricFamilies {
		name := mf.GetName()
		for _, m := range mf.GetMetric() {
			if m.GetCounter() != nil {
				metrics[name] += m.GetCounter().GetValue()
			}
		}
	}

	assert.Equal(t, float64(2), metrics["travel_orders_total"],
		"expected 2 orders recorded")
	assert.Equal(t, float64(2), metrics["travel_payments_success_total"],
		"expected 2 successful payments")
	assert.Equal(t, float64(1), metrics["travel_payments_failure_total"],
		"expected 1 failed payment")
}
