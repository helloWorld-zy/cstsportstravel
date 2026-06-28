// Package repository provides data access for the Order domain.
package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	ordermodel "github.com/travel-booking/server/internal/order/model"
)

// OrderFilter holds optional filter criteria for order listing.
type OrderFilter struct {
	UserID  int64
	Status  string // "all" or specific status
	Page    int
	PerPage int
}

// OrderRepository provides data access for MainOrder and related entities.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order with travellers and initial status log in a transaction.
func (r *OrderRepository) Create(order *ordermodel.MainOrder, travellers []ordermodel.OrderTraveller, statusLog *ordermodel.OrderStatusLog) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the order
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		// Create travellers with order_id set
		for i := range travellers {
			travellers[i].OrderID = order.ID
		}
		if len(travellers) > 0 {
			if err := tx.Create(&travellers).Error; err != nil {
				return fmt.Errorf("create travellers: %w", err)
			}
		}

		// Create initial status log
		if statusLog != nil {
			statusLog.OrderID = order.ID
			if err := tx.Create(statusLog).Error; err != nil {
				return fmt.Errorf("create status log: %w", err)
			}
		}

		return nil
	})
}

// FindByID returns an order with all relations preloaded.
func (r *OrderRepository) FindByID(id int64) (*ordermodel.MainOrder, error) {
	var order ordermodel.MainOrder
	err := r.db.
		Preload("Travellers", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("SubOrders").
		Preload("StatusLogs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByUserID returns a paginated order list for a user with optional status filter.
func (r *OrderRepository) FindByUserID(filter OrderFilter) ([]ordermodel.MainOrder, int64, error) {
	query := r.db.Model(&ordermodel.MainOrder{}).Where("user_id = ?", filter.UserID)

	// Apply status filter
	if filter.Status != "" && filter.Status != "all" {
		// Map visible statuses to internal statuses
		switch filter.Status {
		case "pending_travel":
			query = query.Where("order_status IN ?", []string{
				ordermodel.OrderStatusPaidFull,
				ordermodel.OrderStatusPendingTravel,
			})
		default:
			query = query.Where("order_status = ?", filter.Status)
		}
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	// Paginate
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var orders []ordermodel.MainOrder
	err := query.
		Preload("Travellers").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&orders).Error
	if err != nil {
		return nil, 0, fmt.Errorf("find orders: %w", err)
	}

	return orders, total, nil
}

// UpdateStatus updates the order status and creates a status log entry.
func (r *OrderRepository) UpdateStatus(orderID int64, fromStatus, toStatus, operatorType string, operatorID *int64, reason string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update order status
		updates := map[string]interface{}{
			"order_status": toStatus,
			"updated_at":   time.Now(),
		}

		// Set timestamp fields based on target status
		switch toStatus {
		case ordermodel.OrderStatusPaidFull:
			now := time.Now()
			updates["paid_at"] = now
			updates["payment_status"] = ordermodel.PaymentStatusPaid
		case ordermodel.OrderStatusCancelled:
			now := time.Now()
			updates["cancelled_at"] = now
			updates["cancel_reason"] = reason
		case ordermodel.OrderStatusCompleted:
			now := time.Now()
			updates["completed_at"] = now
		}

		result := tx.Model(&ordermodel.MainOrder{}).
			Where("id = ? AND order_status = ?", orderID, fromStatus).
			Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("update order status: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("order not found or status mismatch")
		}

		// Create status log
		log := &ordermodel.OrderStatusLog{
			OrderID:      orderID,
			FromStatus:   fromStatus,
			ToStatus:     toStatus,
			OperatorType: operatorType,
			OperatorID:   operatorID,
			Reason:       reason,
		}
		if err := tx.Create(log).Error; err != nil {
			return fmt.Errorf("create status log: %w", err)
		}

		return nil
	})
}

// CreateStatusLog creates a status log entry.
func (r *OrderRepository) CreateStatusLog(log *ordermodel.OrderStatusLog) error {
	return r.db.Create(log).Error
}

// FindPendingPayOrders finds orders that are still pending payment and past their expiry.
func (r *OrderRepository) FindExpiredPendingPayOrders(before time.Time) ([]ordermodel.MainOrder, error) {
	var orders []ordermodel.MainOrder
	err := r.db.
		Where("order_status = ? AND created_at < ?", ordermodel.OrderStatusPendingPay, before).
		Find(&orders).Error
	return orders, err
}

// FindByIDBasic returns an order without preloads.
func (r *OrderRepository) FindByIDBasic(id int64) (*ordermodel.MainOrder, error) {
	var order ordermodel.MainOrder
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// CountByUserAndStatus returns the count of orders for a user with a specific status.
func (r *OrderRepository) CountByUserAndStatus(userID int64, status string) (int64, error) {
	var count int64
	err := r.db.Model(&ordermodel.MainOrder{}).
		Where("user_id = ? AND order_status = ?", userID, status).
		Count(&count).Error
	return count, err
}
