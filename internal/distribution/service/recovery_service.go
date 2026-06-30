package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// RecoveryService handles commission recovery on refunds.
type RecoveryService struct {
	commissionRepo  *repository.CommissionRepository
	distributorRepo *repository.DistributorRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewRecoveryService creates a new RecoveryService.
func NewRecoveryService(
	commissionRepo *repository.CommissionRepository,
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *RecoveryService {
	return &RecoveryService{
		commissionRepo:  commissionRepo,
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
	}
}

// RecoveryInput represents the input for commission recovery.
type RecoveryInput struct {
	OrderID        int64   `json:"order_id"`
	RefundAmount   float64 `json:"refund_amount"`   // 退款金额
	OrderPaidAmount float64 `json:"order_paid_amount"` // 订单实付金额
}

// RecoveryResult represents the result of commission recovery.
type RecoveryResult struct {
	RecoveredCommissions []*domain.CommissionDetail `json:"recovered_commissions"`
	TotalRecovered       float64                    `json:"total_recovered"`
}

// ProcessRefundRecovery processes commission recovery when an order is refunded.
// PRD §8.7.3:
// - 全额退款: 全额追回佣金
// - 部分退款: 按退款比例追回 (应追回佣金 = 原佣金 × 退款金额 ÷ 实付金额)
// - 冻结期内退款: 直接调整
// - 冻结期外退款: 从可提现余额扣除
func (s *RecoveryService) ProcessRefundRecovery(input RecoveryInput) (*RecoveryResult, error) {
	if input.RefundAmount <= 0 {
		return nil, fmt.Errorf("refund amount must be positive")
	}
	if input.OrderPaidAmount <= 0 {
		return nil, fmt.Errorf("order paid amount must be positive")
	}
	if input.RefundAmount > input.OrderPaidAmount {
		return nil, fmt.Errorf("refund amount (%.2f) cannot exceed order paid amount (%.2f)", input.RefundAmount, input.OrderPaidAmount)
	}

	// Get all commission details for this order
	commissions, err := s.commissionRepo.FindByOrderID(input.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find commissions for order: %w", err)
	}

	if len(commissions) == 0 {
		return &RecoveryResult{}, nil
	}

	// Calculate recovery ratio
	refundRatio := input.RefundAmount / input.OrderPaidAmount
	isFullRefund := input.RefundAmount >= input.OrderPaidAmount

	result := &RecoveryResult{
		RecoveredCommissions: make([]*domain.CommissionDetail, 0),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, commission := range commissions {
			// Skip already recovered commissions
			if commission.Status == domain.CommissionStatusRecovered {
				continue
			}

			var recoverAmount float64
			if isFullRefund {
				// 全额退款: 全额追回
				recoverAmount = commission.CommissionAmount
			} else {
				// 部分退款: 按比例追回
				recoverAmount = commission.CommissionAmount * refundRatio
			}

			// Round to 2 decimal places
			recoverAmount = float64(int(recoverAmount*100)) / 100

			if recoverAmount <= 0 {
				continue
			}

			// Determine how to recover based on commission status
			switch commission.Status {
			case domain.CommissionStatusPending, domain.CommissionStatusFrozen:
				// 冻结期内退款: 直接调整（状态变为已追回）
				commission.Status = domain.CommissionStatusRecovered
				commission.RecoveredAmount = &recoverAmount
				commission.UpdatedAt = time.Now()

				if err := tx.Save(commission).Error; err != nil {
					return fmt.Errorf("failed to update commission %d: %w", commission.ID, err)
				}

			case domain.CommissionStatusWithdrawable:
				// 冻结期外退款: 从可提现余额扣除
				commission.Status = domain.CommissionStatusRecovered
				commission.RecoveredAmount = &recoverAmount
				commission.UpdatedAt = time.Now()

				if err := tx.Save(commission).Error; err != nil {
					return fmt.Errorf("failed to update commission %d: %w", commission.ID, err)
				}

				// Deduct from distributor's withdrawable balance
				if err := s.deductFromBalance(tx, commission.DistributorID, recoverAmount); err != nil {
					return err
				}

			case domain.CommissionStatusWithdrawn:
				// 已提现的佣金: 从可提现余额扣除（可能产生负值）
				commission.Status = domain.CommissionStatusRecovered
				commission.RecoveredAmount = &recoverAmount
				commission.UpdatedAt = time.Now()

				if err := tx.Save(commission).Error; err != nil {
					return fmt.Errorf("failed to update commission %d: %w", commission.ID, err)
				}

				// Deduct from distributor's balance (may go negative)
				if err := s.deductFromBalance(tx, commission.DistributorID, recoverAmount); err != nil {
					return err
				}
			}

			result.RecoveredCommissions = append(result.RecoveredCommissions, &commission)
			result.TotalRecovered += recoverAmount

			s.logger.Info("commission recovered",
				zap.Int64("commission_id", commission.ID),
				zap.Int64("distributor_id", commission.DistributorID),
				zap.Float64("recover_amount", recoverAmount),
				zap.String("original_status", commission.Status),
			)
		}

		// Update distributor totals
		distIDs := make(map[int64]bool)
		for _, c := range result.RecoveredCommissions {
			distIDs[c.DistributorID] = true
		}
		for distID := range distIDs {
			if err := s.distributorRepo.UpdateCommissionTotals(distID); err != nil {
				return fmt.Errorf("failed to update distributor %d totals: %w", distID, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// deductFromBalance deducts an amount from the distributor's withdrawable balance.
func (s *RecoveryService) deductFromBalance(tx *gorm.DB, distributorID int64, amount float64) error {
	return tx.Model(&domain.Distributor{}).
		Where("id = ?", distributorID).
		Updates(map[string]interface{}{
			"withdrawable_amount": gorm.Expr("withdrawable_amount - ?", amount),
			"updated_at":          time.Now(),
		}).Error
}
