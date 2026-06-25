// Package service provides business logic for the Product domain.
//
// This file implements inventory management with a two-phase locking strategy:
//   - Redis atomic DECRBY for O(1) hot-path stock check-and-decrement
//   - PostgreSQL SELECT FOR UPDATE as the consistency guarantee
//   - Scheduled reconciliation to sync Redis counters with DB
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	productmodel "github.com/travel-booking/server/internal/product/model"
)

// Inventory errors.
var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrDepartureNotFound = errors.New("departure not found")
	ErrDepartureNotOpen  = errors.New("departure is not open for booking")
	ErrBookingCutoff     = errors.New("booking cutoff date has passed")
)

// Stock level constants.
const (
	StockLevelAdequate = "adequate"
	StockLevelTight    = "tight"
	StockLevelFull     = "full"
)

// InventoryService manages departure stock with Redis + DB two-phase locking.
type InventoryService struct {
	db     *gorm.DB
	rdb    *redis.Client
	logger *zap.Logger
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(db *gorm.DB, rdb *redis.Client, logger *zap.Logger) *InventoryService {
	return &InventoryService{
		db:     db,
		rdb:    rdb,
		logger: logger,
	}
}

// stockRedisKey returns the Redis key for a departure's available stock.
func stockRedisKey(departureID int64) string {
	return fmt.Sprintf("stock:departure:%d", departureID)
}

// LockStock attempts to lock `count` seats for the given departure.
// Two-phase approach:
//  1. Redis DECRBY atomic check-and-decrement (fast path)
//  2. PostgreSQL SELECT FOR UPDATE + update locked_count (consistency guarantee)
//
// If Redis is unavailable, falls back to DB-only locking.
func (s *InventoryService) LockStock(ctx context.Context, departureID int64, count int) error {
	if count <= 0 {
		return fmt.Errorf("lock count must be positive")
	}

	// Phase 1: Redis atomic decrement
	key := stockRedisKey(departureID)

	// Initialize Redis stock from DB if key doesn't exist
	exists, _ := s.rdb.Exists(ctx, key).Result()
	if exists == 0 {
		if err := s.InitRedisStock(ctx, departureID); err != nil {
			s.logger.Warn("failed to init redis stock, using DB only",
				zap.Int64("departure_id", departureID), zap.Error(err))
			return s.lockStockDB(ctx, departureID, count)
		}
	}

	remaining, err := s.rdb.DecrBy(ctx, key, int64(count)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		s.logger.Warn("redis decr failed, falling back to DB",
			zap.Int64("departure_id", departureID), zap.Error(err))
		// Fallback: DB-only locking
		return s.lockStockDB(ctx, departureID, count)
	}

	// If stock went negative, rollback Redis and return error
	if remaining < 0 {
		s.rdb.IncrBy(ctx, key, int64(count)) // rollback
		return ErrInsufficientStock
	}

	// Phase 2: DB row-level lock
	if err := s.lockStockDB(ctx, departureID, count); err != nil {
		// Rollback Redis on DB failure
		s.rdb.IncrBy(ctx, key, int64(count))
		return err
	}

	return nil
}

// lockStockDB uses PostgreSQL SELECT FOR UPDATE to atomically check and lock stock.
func (s *InventoryService) lockStockDB(ctx context.Context, departureID int64, count int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dep productmodel.DepartureDate
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", departureID).
			First(&dep).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrDepartureNotFound
			}
			return fmt.Errorf("select departure for update: %w", err)
		}

		// Check departure status
		if dep.Status != productmodel.DepartureStatusOpen {
			return ErrDepartureNotOpen
		}

		// Check booking cutoff
		cutoffDate := dep.DepartureDate.AddDate(0, 0, -dep.CutoffDays)
		if time.Now().After(cutoffDate) {
			return ErrBookingCutoff
		}

		// Check available stock
		available := dep.TotalStock - dep.SoldCount - dep.LockedCount
		if available < count {
			return ErrInsufficientStock
		}

		// Increment locked_count
		result := tx.Model(&productmodel.DepartureDate{}).
			Where("id = ?", departureID).
			Update("locked_count", gorm.Expr("locked_count + ?", count))
		if result.Error != nil {
			return fmt.Errorf("update locked_count: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrDepartureNotFound
		}

		return nil
	})
}

// ReleaseStock releases `count` previously locked seats for the given departure.
func (s *InventoryService) ReleaseStock(ctx context.Context, departureID int64, count int) error {
	if count <= 0 {
		return fmt.Errorf("release count must be positive")
	}

	// Phase 1: Redis INCRBY (best-effort)
	key := stockRedisKey(departureID)
	s.rdb.IncrBy(ctx, key, int64(count))

	// Phase 2: DB decrement locked_count
	result := s.db.WithContext(ctx).
		Model(&productmodel.DepartureDate{}).
		Where("id = ?", departureID).
		Update("locked_count", gorm.Expr("GREATEST(locked_count - ?, 0)", count))
	if result.Error != nil {
		return fmt.Errorf("release locked_count: %w", result.Error)
	}

	return nil
}

// ConfirmStock converts locked seats to sold seats (called after payment success).
func (s *InventoryService) ConfirmStock(ctx context.Context, departureID int64, count int) error {
	if count <= 0 {
		return fmt.Errorf("confirm count must be positive")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Move from locked_count to sold_count
		result := tx.Model(&productmodel.DepartureDate{}).
			Where("id = ? AND locked_count >= ?", departureID, count).
			Updates(map[string]interface{}{
				"locked_count": gorm.Expr("locked_count - ?", count),
				"sold_count":   gorm.Expr("sold_count + ?", count),
			})
		if result.Error != nil {
			return fmt.Errorf("confirm stock: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrInsufficientStock
		}

		// Check if departure is now full
		var dep productmodel.DepartureDate
		if err := tx.First(&dep, departureID).Error; err != nil {
			return err
		}
		if dep.AvailableStock() <= 0 {
			tx.Model(&productmodel.DepartureDate{}).
				Where("id = ?", departureID).
				Update("status", productmodel.DepartureStatusFull)
		}

		return nil
	})
}

// GetAvailableStock returns the current available stock for a departure.
// Tries Redis first, falls back to DB.
func (s *InventoryService) GetAvailableStock(ctx context.Context, departureID int64) (int, error) {
	key := stockRedisKey(departureID)
	val, err := s.rdb.Get(ctx, key).Int64()
	if err == nil {
		return int(val), nil
	}

	// Fallback to DB
	var dep productmodel.DepartureDate
	if err := s.db.WithContext(ctx).First(&dep, departureID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrDepartureNotFound
		}
		return 0, err
	}

	available := dep.AvailableStock()

	// Sync to Redis for next time
	s.rdb.Set(ctx, key, available, 10*time.Minute)

	return available, nil
}

// GetStockLevel returns the stock level category for a departure.
func (s *InventoryService) GetStockLevel(available int, total int) string {
	if available <= 0 {
		return StockLevelFull
	}
	ratio := float64(available) / float64(total)
	if ratio <= 0.1 {
		return StockLevelTight
	}
	return StockLevelAdequate
}

// SyncRedisStock syncs Redis stock counter from DB for a departure.
// Used by reconciliation tasks to prevent drift.
func (s *InventoryService) SyncRedisStock(ctx context.Context, departureID int64) error {
	var dep productmodel.DepartureDate
	if err := s.db.WithContext(ctx).First(&dep, departureID).Error; err != nil {
		return err
	}

	key := stockRedisKey(departureID)
	available := dep.AvailableStock()
	return s.rdb.Set(ctx, key, available, 10*time.Minute).Err()
}

// InitRedisStock initializes the Redis stock counter from DB if not already set.
func (s *InventoryService) InitRedisStock(ctx context.Context, departureID int64) error {
	key := stockRedisKey(departureID)
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil // already initialized
	}

	return s.SyncRedisStock(ctx, departureID)
}
