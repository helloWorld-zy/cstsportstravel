// Package repository provides data access for the Payment domain.
package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	paymentmodel "github.com/travel-booking/server/internal/payment/model"
)

// PaymentRepository provides data access for PaymentTransaction and RefundRecord.
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment transaction.
func (r *PaymentRepository) Create(tx *paymentmodel.PaymentTransaction) error {
	return r.db.Create(tx).Error
}

// FindByID returns a payment transaction by ID.
func (r *PaymentRepository) FindByID(id int64) (*paymentmodel.PaymentTransaction, error) {
	var tx paymentmodel.PaymentTransaction
	if err := r.db.First(&tx, id).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

// FindByOrderID returns the latest payment transaction for an order and channel.
func (r *PaymentRepository) FindByOrderID(orderID int64, channel string) (*paymentmodel.PaymentTransaction, error) {
	var tx paymentmodel.PaymentTransaction
	query := r.db.Where("order_id = ?", orderID)
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	err := query.Order("created_at DESC").First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// FindByPaymentNo returns a payment transaction by payment number.
func (r *PaymentRepository) FindByPaymentNo(paymentNo string) (*paymentmodel.PaymentTransaction, error) {
	var tx paymentmodel.PaymentTransaction
	if err := r.db.Where("payment_no = ?", paymentNo).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

// UpdateStatus updates the payment transaction status and related fields.
func (r *PaymentRepository) UpdateStatus(id int64, status string, extra map[string]interface{}) error {
	updates := map[string]interface{}{
		"status": status,
	}
	for k, v := range extra {
		updates[k] = v
	}

	result := r.db.Model(&paymentmodel.PaymentTransaction{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update payment status: %w", result.Error)
	}
	return nil
}

// HasActivePayment checks if an order has an active (created/paying) payment.
func (r *PaymentRepository) HasActivePayment(orderID int64) (bool, error) {
	var count int64
	err := r.db.Model(&paymentmodel.PaymentTransaction{}).
		Where("order_id = ? AND status IN ?", orderID, []string{
			paymentmodel.PaymentTxnStatusCreated,
			paymentmodel.PaymentTxnStatusPaying,
		}).
		Count(&count).Error
	return count > 0, err
}

// CreateRefundRecord creates a new refund record.
func (r *PaymentRepository) CreateRefundRecord(record *paymentmodel.RefundRecord) error {
	return r.db.Create(record).Error
}

// FindRefundByID returns a refund record by ID.
func (r *PaymentRepository) FindRefundByID(id int64) (*paymentmodel.RefundRecord, error) {
	var record paymentmodel.RefundRecord
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// FindRefundByOrderID returns the latest refund record for an order.
func (r *PaymentRepository) FindRefundByOrderID(orderID int64) (*paymentmodel.RefundRecord, error) {
	var record paymentmodel.RefundRecord
	err := r.db.Where("order_id = ?", orderID).
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateRefundStatus updates the refund record status and related fields.
func (r *PaymentRepository) UpdateRefundStatus(id int64, status string, extra map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	for k, v := range extra {
		updates[k] = v
	}

	result := r.db.Model(&paymentmodel.RefundRecord{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update refund status: %w", result.Error)
	}
	return nil
}
