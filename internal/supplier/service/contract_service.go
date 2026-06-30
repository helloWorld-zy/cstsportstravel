// Package service provides business logic for the Supplier domain.
package service

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/supplier/model"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// ContractService handles e-contract generation, signing reminders, and auto-close.
type ContractService struct {
	supplierRepo *repository.SupplierRepository
	logger       *zap.Logger
}

// NewContractService creates a new ContractService.
func NewContractService(supplierRepo *repository.SupplierRepository, logger *zap.Logger) *ContractService {
	return &ContractService{
		supplierRepo: supplierRepo,
		logger:       logger,
	}
}

// ContractTemplate represents an e-contract template.
type ContractTemplate struct {
	ID         int64  `json:"id"`
	TemplateNo string `json:"template_no"`
	Name       string `json:"name"`
	SupplierType string `json:"supplier_type"` // travel_agent, scenic, hotel
	Content    string `json:"content"`
	Version    string `json:"version"`
}

// GenerateContract generates an e-contract PDF for an approved supplier.
// Contract template varies by supplier type.
func (s *ContractService) GenerateContract(tenantID, supplierID int64) (string, error) {
	supplier, err := s.supplierRepo.FindByID(tenantID, supplierID)
	if err != nil {
		return "", fmt.Errorf("supplier not found: %w", err)
	}

	if supplier.Status != model.SupplierStatusActive {
		return "", fmt.Errorf("supplier must be active to generate contract, current status: %s", supplier.Status)
	}

	// TODO: select template based on supplier type
	// TODO: fill template with supplier data
	// TODO: generate PDF
	// TODO: return PDF URL

	contractURL := fmt.Sprintf("/contracts/%s.pdf", supplier.SupplierNo)
	s.logger.Info("contract generated",
		zap.String("supplier_no", supplier.SupplierNo),
		zap.String("contract_url", contractURL),
	)

	return contractURL, nil
}

// RequestSigning sends a signing request to the supplier via CA-certified e-signature.
func (s *ContractService) RequestSigning(tenantID, supplierID int64) error {
	supplier, err := s.supplierRepo.FindByID(tenantID, supplierID)
	if err != nil {
		return err
	}

	// TODO: integrate with CA e-signature service (e.g., e签宝/法大大)
	// TODO: create signing task
	s.logger.Info("signing request sent",
		zap.String("supplier_no", supplier.SupplierNo),
		zap.String("contact_phone", supplier.ContactPhone),
	)

	return nil
}

// CheckSigningTimeout checks for unsigned contracts and handles reminders/closures.
// Business rules: 7-day reminder, 30-day auto-close.
func (s *ContractService) CheckSigningTimeout() error {
	// TODO: query suppliers in "active" status without contract_signed_at
	// For suppliers waiting > 7 days: send reminder
	// For suppliers waiting > 30 days: auto-close application

	s.logger.Info("checking signing timeouts")
	return nil
}

// ConfirmContractSigning marks a contract as signed by the supplier.
func (s *ContractService) ConfirmContractSigning(tenantID, supplierID int64) error {
	supplier, err := s.supplierRepo.FindByID(tenantID, supplierID)
	if err != nil {
		return err
	}

	now := time.Now()
	supplier.ContractSignedAt = &now
	supplier.UpdatedAt = now

	if err := s.supplierRepo.Update(supplier); err != nil {
		return fmt.Errorf("failed to confirm signing: %w", err)
	}

	s.logger.Info("contract signed",
		zap.String("supplier_no", supplier.SupplierNo),
		zap.Time("signed_at", now),
	)

	return nil
}
