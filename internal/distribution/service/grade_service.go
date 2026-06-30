package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// GradeService handles distributor grade management (auto upgrade/downgrade).
type GradeService struct {
	distributorRepo *repository.DistributorRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewGradeService creates a new GradeService.
func NewGradeService(
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *GradeService {
	return &GradeService{
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
	}
}

// SeniorGradeRequirements defines the requirements for senior grade.
// PRD §8.2.3: 近90天推广订单≥50单且成交总额≥10万元且无违规记录
type SeniorGradeRequirements struct {
	MinOrders    int     `json:"min_orders"`     // ≥50单
	MinAmount    float64 `json:"min_amount"`     // ≥10万元
	ReviewDays   int     `json:"review_days"`    // 90天
	ValidityDays int     `json:"validity_days"`  // 90天有效期
}

// DefaultSeniorGradeRequirements returns the default requirements.
func DefaultSeniorGradeRequirements() SeniorGradeRequirements {
	return SeniorGradeRequirements{
		MinOrders:    50,
		MinAmount:    100000,
		ReviewDays:   90,
		ValidityDays: 90,
	}
}

// GradeCheckResult represents the result of a grade check.
type GradeCheckResult struct {
	DistributorID    int64   `json:"distributor_id"`
	CurrentGrade     string  `json:"current_grade"`
	ShouldUpgrade    bool    `json:"should_upgrade"`
	ShouldDowngrade  bool    `json:"should_downgrade"`
	OrderCount       int     `json:"order_count"`
	TotalAmount      float64 `json:"total_amount"`
	HasViolations    bool    `json:"has_violations"`
}

// CheckGrade checks if a distributor should be upgraded or downgraded.
func (s *GradeService) CheckGrade(distributorID int64) (*GradeCheckResult, error) {
	distributor, err := s.distributorRepo.FindByID(0, distributorID)
	if err != nil {
		return nil, fmt.Errorf("distributor not found: %w", err)
	}

	requirements := DefaultSeniorGradeRequirements()
	since := time.Now().AddDate(0, 0, -requirements.ReviewDays)

	// Count orders and amount in the review period
	var orderCount int
	var totalAmount float64

	err = s.db.Table("commission_detail").
		Where("distributor_id = ? AND commission_level = 1 AND created_at >= ?", distributorID, since).
		Select("COUNT(DISTINCT order_id), COALESCE(SUM(order_actual_amount), 0)").
		Row().Scan(&orderCount, &totalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to query distributor stats: %w", err)
	}

	result := &GradeCheckResult{
		DistributorID: distributorID,
		CurrentGrade:  distributor.Grade,
		OrderCount:    orderCount,
		TotalAmount:   totalAmount,
	}

	// Check upgrade conditions
	if distributor.Grade == domain.DistributorGradeNormal {
		if orderCount >= requirements.MinOrders && totalAmount >= requirements.MinAmount && !result.HasViolations {
			result.ShouldUpgrade = true
		}
	}

	// Check downgrade conditions for senior distributors
	if distributor.Grade == domain.DistributorGradeSenior {
		if distributor.GradeValidUntil != nil && time.Now().After(*distributor.GradeValidUntil) {
			if orderCount < requirements.MinOrders || totalAmount < requirements.MinAmount || result.HasViolations {
				result.ShouldDowngrade = true
			}
		}
	}

	return result, nil
}

// UpgradeToSenior upgrades a distributor to senior grade.
func (s *GradeService) UpgradeToSenior(distributorID int64) error {
	distributor, err := s.distributorRepo.FindByID(0, distributorID)
	if err != nil {
		return fmt.Errorf("distributor not found: %w", err)
	}

	if distributor.Grade == domain.DistributorGradeSenior {
		return fmt.Errorf("distributor is already senior grade")
	}

	validUntil := time.Now().AddDate(0, 0, DefaultSeniorGradeRequirements().ValidityDays)
	distributor.Grade = domain.DistributorGradeSenior
	distributor.GradeValidUntil = &validUntil
	distributor.UpdatedAt = time.Now()

	if err := s.distributorRepo.Update(distributor); err != nil {
		return fmt.Errorf("failed to upgrade distributor: %w", err)
	}

	s.logger.Info("distributor upgraded to senior",
		zap.Int64("distributor_id", distributorID),
		zap.Time("valid_until", validUntil),
	)

	return nil
}

// DowngradeToNormal downgrades a distributor to normal grade.
func (s *GradeService) DowngradeToNormal(distributorID int64) error {
	distributor, err := s.distributorRepo.FindByID(0, distributorID)
	if err != nil {
		return fmt.Errorf("distributor not found: %w", err)
	}

	if distributor.Grade == domain.DistributorGradeNormal {
		return fmt.Errorf("distributor is already normal grade")
	}

	distributor.Grade = domain.DistributorGradeNormal
	distributor.GradeValidUntil = nil
	distributor.UpdatedAt = time.Now()

	if err := s.distributorRepo.Update(distributor); err != nil {
		return fmt.Errorf("failed to downgrade distributor: %w", err)
	}

	s.logger.Info("distributor downgraded to normal",
		zap.Int64("distributor_id", distributorID),
	)

	return nil
}

// AutoReviewGrades automatically reviews and upgrades/downgrades distributors.
// This should be called by an Asynq scheduled job daily.
func (s *GradeService) AutoReviewGrades() (int, int, error) {
	upgraded := 0
	downgraded := 0
	// Get all active distributors
	distributors, _, err := s.distributorRepo.ListByStatus(0, domain.DistributorStatusActive, 1, 10000)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list distributors: %w", err)
	}

	for _, d := range distributors {
		result, err := s.CheckGrade(d.ID)
		if err != nil {
			s.logger.Error("failed to check grade",
				zap.Int64("distributor_id", d.ID),
				zap.Error(err),
			)
			continue
		}

		if result.ShouldUpgrade {
			if err := s.UpgradeToSenior(d.ID); err != nil {
				s.logger.Error("failed to upgrade distributor",
					zap.Int64("distributor_id", d.ID),
					zap.Error(err),
				)
				continue
			}
			upgraded++
		}

		if result.ShouldDowngrade {
			if err := s.DowngradeToNormal(d.ID); err != nil {
				s.logger.Error("failed to downgrade distributor",
					zap.Int64("distributor_id", d.ID),
					zap.Error(err),
				)
				continue
			}
			downgraded++
		}
	}

	s.logger.Info("auto grade review completed",
		zap.Int("upgraded", upgraded),
		zap.Int("downgraded", downgraded),
	)

	return upgraded, downgraded, nil
}
