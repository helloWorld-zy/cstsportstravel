// Package event defines the NATS event bus with unified envelope structure,
// subject naming conventions, and typed event DTOs for inter-service communication.
//
// Event subjects follow the pattern: {domain}.{action}.{version}
// e.g., "order.commission.calculate.v1", "visa.status.changed.v1"
package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Envelope is the unified event structure for all NATS messages.
// All events MUST use this envelope format (Constitution: API-First Design).
type Envelope struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	TraceID   string          `json:"trace_id"`
	ServiceID string          `json:"service_id"` // publishing service name
}

// NewEnvelope creates a new event envelope with auto-generated trace ID.
func NewEnvelope(eventType string, payload interface{}, serviceID string) (*Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal event payload: %w", err)
	}

	return &Envelope{
		EventType: eventType,
		Payload:   data,
		Timestamp: time.Now().UTC(),
		TraceID:   uuid.New().String(),
		ServiceID: serviceID,
	}, nil
}

// Marshal serializes the envelope to JSON bytes.
func (e *Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalEnvelope deserializes JSON bytes into an Envelope.
func UnmarshalEnvelope(data []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}
	return &env, nil
}

// UnmarshalPayload deserializes the payload into the given struct.
func (e *Envelope) UnmarshalPayload(dest interface{}) error {
	return json.Unmarshal(e.Payload, dest)
}

// ────────────────────────────────────────────────────────────────────────────
// NATS Subject Definitions
// ────────────────────────────────────────────────────────────────────────────

const (
	// StreamNames
	StreamOrders      = "ORDERS"
	StreamPayments    = "PAYMENTS"
	StreamVisa        = "VISA"
	StreamDistribution = "DISTRIBUTION"
	StreamMarketing   = "MARKETING"

	// ── Commission Calculation Events ──
	// Published by order-service when an order is paid.
	// Consumed by distribution-service to calculate commissions.
	SubjectOrderPaid        = "order.paid.v1"
	SubjectOrderRefunded    = "order.refunded.v1"

	// Commission calculation result (internal to distribution-service)
	SubjectCommissionCalculated = "commission.calculated.v1"
	SubjectCommissionFrozen     = "commission.frozen.v1"
	SubjectCommissionThawed     = "commission.thawed.v1"
	SubjectCommissionRecovered  = "commission.recovered.v1"

	// ── Visa Status Change Events ──
	// Published by order-service when visa status changes.
	// Consumed by notification service and user-facing services.
	SubjectVisaStatusChanged    = "visa.status.changed.v1"
	SubjectVisaApproved         = "visa.approved.v1"
	SubjectVisaRejected         = "visa.rejected.v1"
	SubjectVisaMaterialRequired = "visa.material.required.v1"

	// ── Reconciliation Events ──
	// Published by payment-service for settlement reconciliation.
	SubjectReconciliationTaskCreated = "reconciliation.task.created.v1"
	SubjectReconciliationCompleted   = "reconciliation.completed.v1"
	SubjectSettlementGenerated       = "settlement.generated.v1"

	// ── Supplier Events ──
	SubjectSupplierApplicationSubmitted = "supplier.application.submitted.v1"
	SubjectSupplierApproved             = "supplier.approved.v1"
	SubjectSupplierSuspended            = "supplier.suspended.v1"

	// ── Marketing Events ──
	SubjectCouponClaimed  = "coupon.claimed.v1"
	SubjectCouponUsed     = "coupon.used.v1"
	SubjectCouponReturned = "coupon.returned.v1"
)

// ────────────────────────────────────────────────────────────────────────────
// Event DTOs (Payloads)
// ────────────────────────────────────────────────────────────────────────────

// OrderPaidPayload is published when an order payment is confirmed.
type OrderPaidPayload struct {
	OrderID          int64   `json:"order_id"`
	OrderNo          string  `json:"order_no"`
	UserID           int64   `json:"user_id"`
	TenantID         int64   `json:"tenant_id"`
	ActualAmount     float64 `json:"actual_amount"`
	PaymentChannel   string  `json:"payment_channel"`
	PaymentMode      string  `json:"payment_mode"` // full, deposit
	DistributorIDl1  *int64  `json:"distributor_id_l1,omitempty"`
	DistributorIDl2  *int64  `json:"distributor_id_l2,omitempty"`
	PromotionCode    string  `json:"promotion_code,omitempty"`
	ProductType      string  `json:"product_type"`
	PaidAt           string  `json:"paid_at"`
}

// OrderRefundedPayload is published when an order refund is processed.
type OrderRefundedPayload struct {
	OrderID        int64   `json:"order_id"`
	OrderNo        string  `json:"order_no"`
	UserID         int64   `json:"user_id"`
	TenantID       int64   `json:"tenant_id"`
	RefundAmount   float64 `json:"refund_amount"`
	RefundType     string  `json:"refund_type"` // full, partial
	RefundReason   string  `json:"refund_reason"`
	RefundedAt     string  `json:"refunded_at"`
}

// VisaStatusChangedPayload is published when a visa order status changes.
type VisaStatusChangedPayload struct {
	VisaOrderID int64  `json:"visa_order_id"`
	OrderID     int64  `json:"order_id"`
	UserID      int64  `json:"user_id"`
	TenantID    int64  `json:"tenant_id"`
	CountryID   int64  `json:"country_id"`
	FromStatus  string `json:"from_status"`
	ToStatus    string `json:"to_status"`
	OperatorID  int64  `json:"operator_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
	ChangedAt   string `json:"changed_at"`
}

// CommissionCalculatedPayload is published after commission calculation.
type CommissionCalculatedPayload struct {
	OrderID          int64   `json:"order_id"`
	OrderActualAmount float64 `json:"order_actual_amount"`
	DistributorIDl1  *int64  `json:"distributor_id_l1,omitempty"`
	DistributorIDl2  *int64  `json:"distributor_id_l2,omitempty"`
	CommissionRateL1 float64 `json:"commission_rate_l1"`
	CommissionRateL2 float64 `json:"commission_rate_l2"`
	CommissionAmtL1  float64 `json:"commission_amount_l1"`
	CommissionAmtL2  float64 `json:"commission_amount_l2"`
	RuleScope        string  `json:"rule_scope"` // global, category, product
	FreezeDays       int     `json:"freeze_days"`
	CalculatedAt     string  `json:"calculated_at"`
}

// ReconciliationTaskCreatedPayload is published when a reconciliation task is created.
type ReconciliationTaskCreatedPayload struct {
	TaskID      int64  `json:"task_id"`
	Channel     string `json:"channel"` // alipay, wechat, unionpay
	TradeDate   string `json:"trade_date"`
	CreatedAt   string `json:"created_at"`
}

// SettlementGeneratedPayload is published when a supplier settlement is generated.
type SettlementGeneratedPayload struct {
	SettlementID int64  `json:"settlement_id"`
	SupplierID   int64  `json:"supplier_id"`
	TenantID     int64  `json:"tenant_id"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
	PayableAmount float64 `json:"payable_amount"`
	GeneratedAt  string `json:"generated_at"`
}
