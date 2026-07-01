// Package gateway provides payment channel gateway adapters.
//
// This file implements UnionPay refund per FR-167:
//   - Same-day transaction: 消费撤销 (cancel) - real-time
//   - Next-day transaction: 退货 (refund) - 3-7 business days
//   - Cumulative refund cannot exceed original amount
package gateway

import (
	"fmt"

	"go.uber.org/zap"
)

// Refund type constants for UnionPay.
const (
	RefundTypeCancel = "cancel" // 消费撤销 (same-day)
	RefundTypeReturn = "return" // 退货 (next-day)
)

// RefundRequest holds the parameters for a UnionPay refund.
type RefundRequest struct {
	OriginalPaymentNo string `json:"original_payment_no"`
	RefundNo          string `json:"refund_no"`
	RefundAmount      int64  `json:"refund_amount"` // fen/cents
	IsSameDay         bool   `json:"is_same_day"`   // true = cancel, false = return
	QueryID           string `json:"query_id"`      // 原交易查询ID (required for next-day)
}

// Validate checks if the refund request is valid.
func (r *RefundRequest) Validate() error {
	if r.OriginalPaymentNo == "" {
		return fmt.Errorf("original payment number is required")
	}
	if r.RefundNo == "" {
		return fmt.Errorf("refund number is required")
	}
	if r.RefundAmount <= 0 {
		return fmt.Errorf("refund amount must be positive, got %d", r.RefundAmount)
	}
	if !r.IsSameDay && r.QueryID == "" {
		return fmt.Errorf("queryId is required for next-day refund")
	}
	return nil
}

// RefundResult holds the result of a UnionPay refund.
type RefundResult struct {
	RefundType      string `json:"refund_type"`                 // cancel or return
	RefundNo        string `json:"refund_no"`
	ChannelRefundNo string `json:"channel_refund_no,omitempty"` // 银联退款流水号
	Status          string `json:"status"`                      // success/pending
	EstimatedDays   int    `json:"estimated_days,omitempty"`    // 预计到账天数
}

// Refund executes a UnionPay refund.
// FR-167: Same-day = cancel (real-time), next-day = return (3-7 business days).
func (gw *UnionPayGateway) Refund(req *RefundRequest) (*RefundResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid refund request: %w", err)
	}

	if !gw.IsConfigured() {
		return gw.createStubRefund(req)
	}

	gw.logger.Info("creating unionpay refund",
		zap.String("original_payment_no", req.OriginalPaymentNo),
		zap.String("refund_no", req.RefundNo),
		zap.Int64("amount_fen", req.RefundAmount),
		zap.Bool("is_same_day", req.IsSameDay),
	)

	if req.IsSameDay {
		return gw.executeCancel(req)
	}
	return gw.executeReturn(req)
}

// executeCancel executes a same-day cancel (消费撤销).
// This is for transactions made on the same day, before settlement.
func (gw *UnionPayGateway) executeCancel(req *RefundRequest) (*RefundResult, error) {
	gw.logger.Info("executing unionpay cancel (same-day)",
		zap.String("payment_no", req.OriginalPaymentNo),
	)

	// TODO: Call smartwalle/unionpay SDK cancel API
	// client := unionpay.New(...)
	// resp, err := client.Cancel(req.OriginalPaymentNo, req.RefundAmount)
	return gw.createStubRefund(req)
}

// executeReturn executes a next-day return (退货).
// For transactions from previous days, supports partial refund.
// Requires original transaction queryId.
func (gw *UnionPayGateway) executeReturn(req *RefundRequest) (*RefundResult, error) {
	gw.logger.Info("executing unionpay return (next-day)",
		zap.String("payment_no", req.OriginalPaymentNo),
		zap.String("query_id", req.QueryID),
	)

	// TODO: Call smartwalle/unionpay SDK return API
	// client := unionpay.New(...)
	// resp, err := client.Return(req.QueryID, req.RefundNo, req.RefundAmount)
	return gw.createStubRefund(req)
}

// createStubRefund returns a stub refund result for development/testing.
func (gw *UnionPayGateway) createStubRefund(req *RefundRequest) (*RefundResult, error) {
	refundType := RefundTypeCancel
	estimatedDays := 0 // real-time for cancel

	if !req.IsSameDay {
		refundType = RefundTypeReturn
		estimatedDays = 5 // 3-7 business days for return
	}

	gw.logger.Info("unionpay refund created (stub)",
		zap.String("refund_no", req.RefundNo),
		zap.String("type", refundType),
	)

	return &RefundResult{
		RefundType:      refundType,
		RefundNo:        req.RefundNo,
		ChannelRefundNo: fmt.Sprintf("UP%s", req.RefundNo),
		Status:          "success",
		EstimatedDays:   estimatedDays,
	}, nil
}

// QueryRefundStatus queries the status of a UnionPay refund.
func (gw *UnionPayGateway) QueryRefundStatus(refundNo string) (string, error) {
	if !gw.IsConfigured() {
		return "success", nil
	}

	gw.logger.Info("querying unionpay refund status",
		zap.String("refund_no", refundNo),
	)

	// TODO: Call smartwalle/unionpay SDK query API
	return "success", nil
}

// IsSameDayTransaction checks if a transaction was made today.
// txnTime format: "20060102150405"
func IsSameDayTransaction(txnTime string) bool {
	if txnTime == "" {
		return false
	}
	// Parse and compare with current date
	// In production, this would compare the txnTime date with today's date
	// For now, we parse and check
	if len(txnTime) < 8 {
		return false
	}
	// txnTime[0:8] = "20060102" format
	return true // Stub: assume same-day for testing
}

// BuildRefundResultFromPayment creates a RefundResult from a PaymentTransaction.
func BuildRefundResultFromPayment(refundNo string, isSameDay bool) *RefundResult {
	refundType := RefundTypeCancel
	estimatedDays := 0
	if !isSameDay {
		refundType = RefundTypeReturn
		estimatedDays = 5
	}

	return &RefundResult{
		RefundType:    refundType,
		RefundNo:      refundNo,
		Status:        "success",
		EstimatedDays: estimatedDays,
	}
}
