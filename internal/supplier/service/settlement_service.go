package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/supplier/model"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// SettlementService implements the 5-step settlement flow.
// Step 1: Auto-generate settlement statement PDF
// Step 2: Supplier 7-day review (confirm or dispute)
// Step 3: Payment approval (tiered by amount)
// Step 4: Execute bank transfer and upload voucher
// Step 5: Archive
type SettlementService struct {
	settlementRepo *repository.SettlementRepository
	supplierRepo   *repository.SupplierRepository
	logger         *zap.Logger
}

// NewSettlementService creates a new SettlementService.
func NewSettlementService(
	settlementRepo *repository.SettlementRepository,
	supplierRepo *repository.SupplierRepository,
	logger *zap.Logger,
) *SettlementService {
	return &SettlementService{
		settlementRepo: settlementRepo,
		supplierRepo:   supplierRepo,
		logger:         logger,
	}
}

// GenerateSettlementRequest contains the data needed to generate a settlement.
type GenerateSettlementRequest struct {
	TenantID    int64
	SupplierID  int64
	PeriodStart time.Time
	PeriodEnd   time.Time
}

// GenerateResult contains the result of settlement generation.
type GenerateResult struct {
	SettlementNo string  `json:"settlement_no"`
	OrderCount   int     `json:"order_count"`
	TotalAmount  float64 `json:"total_amount"`
	PayableAmount float64 `json:"payable_amount"`
}

// GenerateSettlement creates a new settlement statement (Step 1).
// Called by Asynq scheduled task at the end of each settlement cycle.
func (s *SettlementService) GenerateSettlement(req GenerateSettlementRequest) (*GenerateResult, error) {
	supplier, err := s.supplierRepo.FindByID(req.TenantID, req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	settlementNo, err := s.settlementRepo.GenerateSettlementNo(supplier.SupplierNo)
	if err != nil {
		return nil, fmt.Errorf("generate settlement number: %w", err)
	}

	// TODO: aggregate order data for the period
	// TODO: calculate commissions, refunds, payable amount
	// TODO: generate PDF

	settlement := &model.SettlementStatement{
		TenantID:           req.TenantID,
		SettlementNo:       settlementNo,
		SupplierID:         req.SupplierID,
		PeriodStart:        req.PeriodStart,
		PeriodEnd:          req.PeriodEnd,
		OrderCount:         0,  // TODO: aggregate
		OrderTotalAmount:   0,  // TODO: aggregate
		RefundAmount:       0,  // TODO: aggregate
		PlatformCommission: 0,  // TODO: calculate
		PayableAmount:      0,  // TODO: calculate
		Status:             model.SettlementStatusPending,
	}

	if err := s.settlementRepo.Create(settlement); err != nil {
		return nil, fmt.Errorf("create settlement: %w", err)
	}

	s.logger.Info("settlement generated",
		zap.String("settlement_no", settlementNo),
		zap.Int64("supplier_id", req.SupplierID),
	)

	return &GenerateResult{
		SettlementNo:  settlementNo,
		OrderCount:    settlement.OrderCount,
		TotalAmount:   settlement.OrderTotalAmount,
		PayableAmount: settlement.PayableAmount,
	}, nil
}

// CheckPendingReviewTimeout checks for settlements pending review > 7 days.
// Auto-escalates to operations manager if overdue.
func (s *SettlementService) CheckPendingReviewTimeout() error {
	settlements, err := s.settlementRepo.ListPendingReview(0, 7)
	if err != nil {
		return err
	}

	for _, settlement := range settlements {
		s.logger.Warn("settlement review timeout",
			zap.String("settlement_no", settlement.SettlementNo),
			zap.Int64("supplier_id", settlement.SupplierID),
		)
		// TODO: send escalation notification
	}

	return nil
}

// ApprovePayment approves a confirmed settlement for payment (Step 3).
// Tiered approval: ≤1万 财务专员, 1-5万 财务主管, >5万 总监联合审批.
func (s *SettlementService) ApprovePayment(tenantID, settlementID, approverID int64) error {
	settlement, err := s.settlementRepo.FindByID(tenantID, settlementID)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettlementStatusConfirmed {
		return fmt.Errorf("settlement must be confirmed before payment approval, current status: %s", settlement.Status)
	}

	now := time.Now()
	settlement.ApprovedBy = &approverID
	settlement.ApprovedAt = &now

	// TODO: check approval authority based on payable amount

	return s.settlementRepo.Update(settlement)
}

// ExecutePayment records the bank transfer and uploads the voucher (Step 4).
func (s *SettlementService) ExecutePayment(tenantID, settlementID int64, voucherURL string) error {
	settlement, err := s.settlementRepo.FindByID(tenantID, settlementID)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettlementStatusConfirmed {
		return fmt.Errorf("settlement must be confirmed before payment, current status: %s", settlement.Status)
	}

	if err := s.settlementRepo.UpdateStatus(tenantID, settlementID, model.SettlementStatusPaid); err != nil {
		return err
	}

	settlement.PaymentVoucherURL = voucherURL
	now := time.Now()
	settlement.PaidAt = &now
	return s.settlementRepo.Update(settlement)
}

// ArchiveSettlement archives a paid settlement (Step 5).
func (s *SettlementService) ArchiveSettlement(tenantID, settlementID int64) error {
	return s.settlementRepo.UpdateStatus(tenantID, settlementID, model.SettlementStatusArchived)
}
