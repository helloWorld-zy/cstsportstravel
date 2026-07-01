// Package model defines GORM models for the Payment domain.
package model

import (
	"encoding/json"
	"time"
)

// PaymentTransaction represents a payment attempt.
type PaymentTransaction struct {
	ID               int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID          int64           `gorm:"column:order_id;not null;index:idx_payment_order" json:"order_id"`
	PaymentNo        string          `gorm:"column:payment_no;size:30;uniqueIndex;not null" json:"payment_no"`
	Channel          string          `gorm:"column:channel;size:20;not null;index:idx_payment_order" json:"channel"`
	Method           string          `gorm:"column:method;size:30;not null" json:"method"`
	Amount           int64           `gorm:"column:amount;not null" json:"amount"` // cents
	Status           string          `gorm:"column:status;size:20;not null;default:created" json:"status"`
	ChannelTradeNo   string          `gorm:"column:channel_trade_no;size:100" json:"channel_trade_no,omitempty"`
	PaidAt           *time.Time      `gorm:"column:paid_at" json:"paid_at,omitempty"`
	ExpireAt         time.Time       `gorm:"column:expire_at;not null" json:"expire_at"`
	NotifyURL        string          `gorm:"column:notify_url;size:500;not null" json:"notify_url"`
	ExtraParams      json.RawMessage `gorm:"column:extra_params;type:jsonb" json:"extra_params,omitempty"`
	PaymentType      string          `gorm:"column:payment_type;size:20;default:full" json:"payment_type"`           // deposit/balance/full/refund
	UnionpayTradeNo  string          `gorm:"column:unionpay_trade_no;size:64" json:"unionpay_trade_no,omitempty"`   // 银联交易号
	UnionpayQueryID  string          `gorm:"column:unionpay_query_id;size:64" json:"unionpay_query_id,omitempty"`   // 银联查询ID
	CreatedAt        time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (PaymentTransaction) TableName() string {
	return "payment_transaction"
}

// RefundRecord represents a refund request and execution.
type RefundRecord struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID         int64      `gorm:"column:order_id;not null;index" json:"order_id"`
	PaymentID       int64      `gorm:"column:payment_id;not null" json:"payment_id"`
	RefundNo        string     `gorm:"column:refund_no;size:30;uniqueIndex;not null" json:"refund_no"`
	RefundAmount    int64      `gorm:"column:refund_amount;not null" json:"refund_amount"` // cents
	RefundReason    string     `gorm:"column:refund_reason;size:500;not null" json:"refund_reason"`
	RefundType      string     `gorm:"column:refund_type;size:20;not null" json:"refund_type"`
	Status          string     `gorm:"column:status;size:20;not null;default:pending;index" json:"status"`
	ApprovalLevel   string     `gorm:"column:approval_level;size:30;not null" json:"approval_level"`
	ApprovedBy      *int64     `gorm:"column:approved_by" json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	ChannelRefundNo string     `gorm:"column:channel_refund_no;size:100" json:"channel_refund_no,omitempty"`
	CompletedAt     *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;default:now()" json:"updated_at"`
}

// TableName overrides the table name.
func (RefundRecord) TableName() string {
	return "refund_record"
}

// Payment status constants.
const (
	PaymentTxnStatusCreated  = "created"
	PaymentTxnStatusPaying   = "paying"
	PaymentTxnStatusPaid     = "paid"
	PaymentTxnStatusFailed   = "failed"
	PaymentTxnStatusClosed   = "closed"
	PaymentTxnStatusRefunded = "refunded"
)

// Payment channel constants.
const (
	ChannelAlipay    = "alipay"
	ChannelWechat    = "wechat"
	ChannelUnionPay  = "unionpay"
)

// Payment method constants.
const (
	MethodNative  = "native"
	MethodJSAPI   = "jsapi"
	MethodH5      = "h5"
	MethodWAP     = "wap"
	MethodGateway = "gateway" // UnionPay gateway payment (PC)
)

// Refund status constants.
const (
	RefundStatusPending    = "pending"
	RefundStatusApproved   = "approved"
	RefundStatusProcessing = "processing"
	RefundStatusSuccess    = "success"
	RefundStatusFailed     = "failed"
)

// Refund type constants.
const (
	RefundTypeFull    = "full"
	RefundTypePartial = "partial"
)

// Payment type constants (deposit/balance/full/refund).
const (
	PaymentTypeFull     = "full"
	PaymentTypeDeposit  = "deposit"
	PaymentTypeBalance  = "balance"
	PaymentTypeRefund   = "refund"
)

// Approval level constants.
const (
	ApprovalLevelOperator       = "operator"
	ApprovalLevelFinanceDirector = "finance_director"
	ApprovalLevelDirector        = "director"
)
