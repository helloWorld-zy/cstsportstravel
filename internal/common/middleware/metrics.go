// Package middleware provides Prometheus metrics collection for HTTP requests
// and business operations.
//
// Application-layer metrics follow the RED method (Rate / Errors / Duration).
// Business metrics track order volume, payment success rate, and refund rate.
//
// All metrics use the "travel_" namespace prefix for consistent identification
// in Prometheus and Grafana.
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsMiddleware holds Prometheus metric collectors for HTTP request tracking.
type MetricsMiddleware struct {
	requestDuration *prometheus.HistogramVec
	requestTotal    *prometheus.CounterVec
	requestErrors   *prometheus.CounterVec
}

// NewMetricsMiddleware creates and registers HTTP metrics with the given registry.
//
// Registered metrics:
//   - travel_http_request_duration_seconds (histogram): HTTP request duration by method/path/status
//   - travel_http_requests_total (counter): Total HTTP requests by method/path/status
//   - travel_http_request_errors_total (counter): Total HTTP 5xx errors by method/path
func NewMetricsMiddleware(registry *prometheus.Registry) *MetricsMiddleware {
	m := &MetricsMiddleware{
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "travel",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status"},
		),
		requestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "travel",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		requestErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "travel",
				Subsystem: "http",
				Name:      "request_errors_total",
				Help:      "Total number of HTTP 5xx error responses.",
			},
			[]string{"method", "path"},
		),
	}

	registry.MustRegister(m.requestDuration, m.requestTotal, m.requestErrors)
	return m
}

// Handler returns a Gin middleware that records HTTP request metrics.
// It captures method, normalized path, status code, and request duration.
func (m *MetricsMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath() // Use route pattern, not actual path (avoids high cardinality)

		if path == "" {
			path = "unknown"
		}

		m.requestDuration.WithLabelValues(method, path, status).Observe(duration)
		m.requestTotal.WithLabelValues(method, path, status).Inc()

		// Count 5xx errors
		if c.Writer.Status() >= 500 {
			m.requestErrors.WithLabelValues(method, path).Inc()
		}
	}
}

// BusinessMetrics holds Prometheus metric collectors for business-level tracking.
type BusinessMetrics struct {
	orderTotal        *prometheus.CounterVec
	paymentSuccess    *prometheus.CounterVec
	paymentFailure    *prometheus.CounterVec
	paymentDuration   *prometheus.HistogramVec
	refundTotal       *prometheus.CounterVec
	activeConnections prometheus.Gauge
}

// NewBusinessMetrics creates and registers business metrics with the given registry.
//
// Registered metrics:
//   - travel_orders_total (counter): Total orders by product type
//   - travel_payments_success_total (counter): Successful payments by channel
//   - travel_payments_failure_total (counter): Failed payments by channel
//   - travel_payment_duration_seconds (histogram): Payment processing duration
//   - travel_refunds_total (counter): Total refund requests by reason
//   - travel_active_connections (gauge): Current active database connections
func NewBusinessMetrics(registry *prometheus.Registry) *BusinessMetrics {
	bm := &BusinessMetrics{
		orderTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "travel",
				Name:      "orders_total",
				Help:      "Total number of orders created.",
			},
			[]string{"product_type"},
		),
		paymentSuccess: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "travel",
				Name:      "payments_success_total",
				Help:      "Total number of successful payments.",
			},
			[]string{"channel"},
		),
		paymentFailure: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "travel",
				Name:      "payments_failure_total",
				Help:      "Total number of failed payments.",
			},
			[]string{"channel"},
		),
		paymentDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "travel",
				Name:      "payment_duration_seconds",
				Help:      "Payment processing duration in seconds.",
				Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
			},
			[]string{"channel"},
		),
		refundTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "travel",
				Name:      "refunds_total",
				Help:      "Total number of refund requests.",
			},
			[]string{"reason"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "travel",
				Name:      "active_connections",
				Help:      "Current number of active database connections.",
			},
		),
	}

	registry.MustRegister(
		bm.orderTotal,
		bm.paymentSuccess,
		bm.paymentFailure,
		bm.paymentDuration,
		bm.refundTotal,
		bm.activeConnections,
	)

	return bm
}

// IncOrderCount increments the order counter for the given product type.
func (bm *BusinessMetrics) IncOrderCount(productType string) {
	bm.orderTotal.WithLabelValues(productType).Inc()
}

// RecordPaymentResult records a payment attempt outcome.
func (bm *BusinessMetrics) RecordPaymentResult(channel string, success bool) {
	if success {
		bm.paymentSuccess.WithLabelValues(channel).Inc()
	} else {
		bm.paymentFailure.WithLabelValues(channel).Inc()
	}
}

// ObservePaymentDuration records the duration of a payment operation.
func (bm *BusinessMetrics) ObservePaymentDuration(channel string, duration time.Duration) {
	bm.paymentDuration.WithLabelValues(channel).Observe(duration.Seconds())
}

// IncRefundCount increments the refund counter for the given reason.
func (bm *BusinessMetrics) IncRefundCount(reason string) {
	bm.refundTotal.WithLabelValues(reason).Inc()
}

// SetActiveConnections sets the current active database connection count.
func (bm *BusinessMetrics) SetActiveConnections(count float64) {
	bm.activeConnections.Set(count)
}
