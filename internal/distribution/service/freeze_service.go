package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// FreezeService handles commission freeze and thaw operations.
type FreezeService struct {
	commissionRepo  *repository.CommissionRepository
	distributorRepo *repository.DistributorRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewFreezeService creates a new FreezeService.
func NewFreezeService(
	commissionRepo *repository.CommissionRepository,
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *FreezeService {
	return &FreezeService{
		commissionRepo:  commissionRepo,
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
	}
}

// FreezeCommission transitions a commission from pending to frozen status.
func (s *FreezeService) FreezeCommission(commissionID int64) error {
	commission, err := s.commissionRepo.FindByID(commissionID)
	if err != nil {
		return fmt.Errorf("commission not found: %w", err)
	}

	if !commission.CanTransitionTo(domain.CommissionStatusFrozen) {
		return fmt.Errorf("cannot freeze commission in status %s", commission.Status)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		commission.Status = domain.CommissionStatusFrozen
		commission.UpdatedAt = time.Now()

		if err := tx.Save(commission).Error; err != nil {
			return fmt.Errorf("failed to update commission status: %w", err)
		}

		// Update distributor frozen amount
		if err := s.distributorRepo.UpdateCommissionTotals(commission.DistributorID); err != nil {
			return fmt.Errorf("failed to update distributor totals: %w", err)
		}

		s.logger.Info("commission frozen",
			zap.Int64("commission_id", commissionID),
			zap.Int64("distributor_id", commission.DistributorID),
			zap.Float64("amount", commission.CommissionAmount),
			zap.Time("frozen_until", *commission.FrozenUntil),
		)

		return nil
	})
}

// ThawExpiredCommissions processes commissions whose freeze period has expired.
// This should be called by an Asynq scheduled job.
func (s *FreezeService) ThawExpiredCommissions(batchSize int) (int, error) {
	commissions, err := s.commissionRepo.FindFrozenDue(batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to find frozen commissions: %w", err)
	}

	thawed := 0
	for _, commission := range commissions {
		if err := s.ThawCommission(commission.ID); err != nil {
			s.logger.Error("failed to thaw commission",
				zap.Int64("commission_id", commission.ID),
				zap.Error(err),
			)
			continue
		}
		thawed++
	}

	if thawed > 0 {
		s.logger.Info("thawed expired commissions", zap.Int("count", thawed))
	}

	return thawed, nil
}

// ThawCommission transitions a commission from frozen to withdrawable status.
func (s *FreezeService) ThawCommission(commissionID int64) error {
	commission, err := s.commissionRepo.FindByID(commissionID)
	if err != nil {
		return fmt.Errorf("commission not found: %w", err)
	}

	if !commission.CanTransitionTo(domain.CommissionStatusWithdrawable) {
		return fmt.Errorf("cannot thaw commission in status %s", commission.Status)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		commission.Status = domain.CommissionStatusWithdrawable
		commission.SettledAt = &now
		commission.UpdatedAt = now

		if err := tx.Save(commission).Error; err != nil {
			return fmt.Errorf("failed to update commission status: %w", err)
		}

		// Update distributor totals
		if err := s.distributorRepo.UpdateCommissionTotals(commission.DistributorID); err != nil {
			return fmt.Errorf("failed to update distributor totals: %w", err)
		}

		s.logger.Info("commission thawed",
			zap.Int64("commission_id", commissionID),
			zap.Int64("distributor_id", commission.DistributorID),
			zap.Float64("amount", commission.CommissionAmount),
		)

		return nil
	})
}
