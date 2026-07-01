// Package service provides business logic for the Payment domain.
//
// This file implements the payment status proactive query service per FR-168:
//   - All 3 channels: Alipay, WeChat, UnionPay
//   - Trigger after 30s if no callback received
//   - Retry every 60s up to max retries
package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// ChannelQuerier defines the interface for querying payment status from a channel.
type ChannelQuerier interface {
	QueryOrder(paymentNo string) (string, error)
}

// ProactiveQueryService handles proactive payment status queries.
// FR-168: Trigger after 30s, retry every 60s, all 3 channels.
type ProactiveQueryService struct {
	alipayQuerier   ChannelQuerier
	wechatQuerier   ChannelQuerier
	unionpayQuerier ChannelQuerier
	triggerDelay    time.Duration
	retryInterval   time.Duration
	maxRetries      int
	logger          *zap.Logger
}

// NewProactiveQueryService creates a new ProactiveQueryService.
func NewProactiveQueryService(
	alipay, wechat, unionpay ChannelQuerier,
	logger *zap.Logger,
) *ProactiveQueryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ProactiveQueryService{
		alipayQuerier:   alipay,
		wechatQuerier:   wechat,
		unionpayQuerier: unionpay,
		triggerDelay:    30 * time.Second, // FR-168: trigger after 30s
		retryInterval:   60 * time.Second, // FR-168: retry every 60s
		maxRetries:      5,
		logger:          logger,
	}
}

// QueryResult holds the result of a proactive query.
type QueryResult struct {
	PaymentNo     string `json:"payment_no"`
	Channel       string `json:"channel"`
	Status        string `json:"status"`
	ChannelTradeNo string `json:"channel_trade_no,omitempty"`
	IsFinal       bool   `json:"is_final"` // true if status is terminal (paid/failed/closed)
}

// ShouldQuery checks if a payment should be proactively queried.
// Returns true if the payment has been in "paying" status for more than triggerDelay.
func (s *ProactiveQueryService) ShouldQuery(ptx *paymentmodel.PaymentTransaction) bool {
	// Only query payments in "paying" status
	if ptx.Status != paymentmodel.PaymentTxnStatusPaying {
		return false
	}

	// Check if enough time has passed since creation
	elapsed := time.Since(ptx.CreatedAt)
	return elapsed >= s.triggerDelay
}

// QueryPaymentStatus queries the payment status from the appropriate channel.
func (s *ProactiveQueryService) QueryPaymentStatus(ptx *paymentmodel.PaymentTransaction) (*QueryResult, error) {
	querier := s.getQuerier(ptx.Channel)
	if querier == nil {
		return nil, fmt.Errorf("no querier for channel: %s", ptx.Channel)
	}

	s.logger.Info("proactively querying payment status",
		zap.String("payment_no", ptx.PaymentNo),
		zap.String("channel", ptx.Channel),
	)

	channelStatus, err := querier.QueryOrder(ptx.PaymentNo)
	if err != nil {
		return nil, fmt.Errorf("query %s order: %w", ptx.Channel, err)
	}

	// Map channel status to internal status
	internalStatus := s.mapChannelStatus(ptx.Channel, channelStatus)
	isFinal := internalStatus == paymentmodel.PaymentTxnStatusPaid ||
		internalStatus == paymentmodel.PaymentTxnStatusFailed ||
		internalStatus == paymentmodel.PaymentTxnStatusClosed

	return &QueryResult{
		PaymentNo: ptx.PaymentNo,
		Channel:   ptx.Channel,
		Status:    internalStatus,
		IsFinal:   isFinal,
	}, nil
}

// getQuerier returns the appropriate channel querier.
func (s *ProactiveQueryService) getQuerier(channel string) ChannelQuerier {
	switch channel {
	case paymentmodel.ChannelAlipay:
		return s.alipayQuerier
	case paymentmodel.ChannelWechat:
		return s.wechatQuerier
	case paymentmodel.ChannelUnionPay:
		return s.unionpayQuerier
	default:
		return nil
	}
}

// mapChannelStatus maps channel-specific status to internal status.
func (s *ProactiveQueryService) mapChannelStatus(channel, channelStatus string) string {
	switch channel {
	case paymentmodel.ChannelAlipay:
		return mapAlipayStatus(channelStatus)
	case paymentmodel.ChannelWechat:
		return mapWechatStatus(channelStatus)
	case paymentmodel.ChannelUnionPay:
		return mapUnionPayStatus(channelStatus)
	default:
		return paymentmodel.PaymentTxnStatusPaying
	}
}

// mapAlipayStatus maps Alipay trade status to internal status.
func mapAlipayStatus(status string) string {
	switch status {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return paymentmodel.PaymentTxnStatusPaid
	case "TRADE_CLOSED":
		return paymentmodel.PaymentTxnStatusClosed
	case "WAIT_BUYER_PAY":
		return paymentmodel.PaymentTxnStatusPaying
	default:
		return paymentmodel.PaymentTxnStatusPaying
	}
}

// mapWechatStatus maps WeChat trade state to internal status.
func mapWechatStatus(status string) string {
	switch status {
	case "SUCCESS":
		return paymentmodel.PaymentTxnStatusPaid
	case "CLOSED", "REVOKED", "PAYERROR":
		return paymentmodel.PaymentTxnStatusFailed
	case "NOTPAY", "USERPAYING":
		return paymentmodel.PaymentTxnStatusPaying
	default:
		return paymentmodel.PaymentTxnStatusPaying
	}
}

// mapUnionPayStatus maps UnionPay response code to internal status.
func mapUnionPayStatus(respCode string) string {
	switch respCode {
	case "00": // 交易成功
		return paymentmodel.PaymentTxnStatusPaid
	case "A6": // 处理中
		return paymentmodel.PaymentTxnStatusPaying
	default: // 其他为失败
		return paymentmodel.PaymentTxnStatusFailed
	}
}
