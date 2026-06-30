// Package handler provides HTTP handlers for the Supplier domain.
package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/shared/middleware"
	"github.com/travel-booking/server/internal/supplier/model"
	"github.com/travel-booking/server/internal/supplier/repository"
)

// ApplicationHandler handles supplier application requests.
type ApplicationHandler struct {
	supplierRepo  *repository.SupplierRepository
	qualRepo      *repository.QualificationRepository
	logger        *zap.Logger
}

// NewApplicationHandler creates a new ApplicationHandler.
func NewApplicationHandler(
	supplierRepo *repository.SupplierRepository,
	qualRepo *repository.QualificationRepository,
	logger *zap.Logger,
) *ApplicationHandler {
	return &ApplicationHandler{
		supplierRepo: supplierRepo,
		qualRepo:     qualRepo,
		logger:       logger,
	}
}

// SubmitApplicationRequest represents the supplier application submission request.
type SubmitApplicationRequest struct {
	CompanyName         string  `form:"companyName" binding:"required"`
	CreditCode          string  `form:"creditCode" binding:"required"`
	RegisteredAddress   string  `form:"registeredAddress"`
	RegisteredCapital   float64 `form:"registeredCapital"`
	EstablishmentDate   string  `form:"establishmentDate"`
	LegalPersonName     string  `form:"legalPersonName" binding:"required"`
	LegalPersonIDCard   string  `form:"legalPersonIdCard" binding:"required"`
	BusinessScope       string  `form:"businessScope"`
	TravelLicenseNo     string  `form:"travelLicenseNo"`
	ContactName         string  `form:"contactName" binding:"required"`
	ContactPhone        string  `form:"contactPhone" binding:"required"`
	ContactEmail        string  `form:"contactEmail"`
	FinanceContactName  string  `form:"financeContactName"`
	FinanceContactPhone string  `form:"financeContactPhone"`
	BankName            string  `form:"bankName"`
	BankAccountName     string  `form:"bankAccountName"`
	BankAccountNumber   string  `form:"bankAccountNumber"`
}

// SubmitApplication handles POST /api/v2/suppliers/apply.
// Accepts multipart form with company info and qualification files.
func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
	var req SubmitApplicationRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, "missing required fields: "+err.Error())
		return
	}

	tenantID := middleware.GetTenantID(c)

	// Check duplicate credit code
	existing, _ := h.supplierRepo.FindByCreditCode(req.CreditCode)
	if existing != nil {
		response.BusinessError(c, response.CodeConflict, "supplier with this credit code already exists")
		return
	}

	// Generate application number
	appNo, err := h.supplierRepo.GenerateApplicationNo()
	if err != nil {
		h.logger.Error("failed to generate application number", zap.Error(err))
		response.ServerError(c, "failed to generate application number")
		return
	}

	// Generate supplier number
	supplierNo, err := h.supplierRepo.GenerateSupplierNo()
	if err != nil {
		h.logger.Error("failed to generate supplier number", zap.Error(err))
		response.ServerError(c, "failed to generate supplier number")
		return
	}

	// Parse establishment date
	var estDate *time.Time
	if req.EstablishmentDate != "" {
		t, err := time.Parse("2006-01-02", req.EstablishmentDate)
		if err == nil {
			estDate = &t
		}
	}

	supplier := &model.Supplier{
		TenantID:               tenantID,
		SupplierNo:             supplierNo,
		CompanyName:            req.CompanyName,
		UnifiedSocialCreditCode: req.CreditCode,
		RegisteredAddress:      req.RegisteredAddress,
		BusinessLicenseURL:     "", // File URL set after upload
		LegalPersonName:        req.LegalPersonName,
		LegalPersonIDCard:      req.LegalPersonIDCard, // TODO: encrypt with AES-256-GCM
		BusinessScope:          req.BusinessScope,
		TravelLicenseNo:        req.TravelLicenseNo,
		ContactName:            req.ContactName,
		ContactPhone:           req.ContactPhone,
		ContactEmail:           req.ContactEmail,
		FinanceContactName:     req.FinanceContactName,
		FinanceContactPhone:    req.FinanceContactPhone,
		BankName:               req.BankName,
		BankAccountName:        req.BankAccountName,
		BankAccountNumber:      req.BankAccountNumber, // TODO: encrypt with AES-256-GCM
		SettlementCycle:        model.SettlementCycleMonthly,
		Status:                 model.SupplierStatusPending,
		ApplicationNo:          appNo,
		AppliedAt:              time.Now(),
	}

	if req.RegisteredCapital > 0 {
		supplier.RegisteredCapital = &req.RegisteredCapital
	}
	if estDate != nil {
		supplier.EstablishmentDate = estDate
	}

	// Handle file uploads
	if err := h.handleFileUploads(c, supplier); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.supplierRepo.Create(supplier); err != nil {
		h.logger.Error("failed to create supplier application", zap.Error(err))
		response.ServerError(c, "failed to submit application")
		return
	}

	// Save qualifications
	h.saveQualifications(c, tenantID, supplier.ID)

	response.OK(c, gin.H{
		"applicationNo": appNo,
		"status":        supplier.Status,
	})
}

// handleFileUploads processes multipart file uploads for business license, ID cards, etc.
func (h *ApplicationHandler) handleFileUploads(c *gin.Context, supplier *model.Supplier) error {
	// Business license
	if file, err := c.FormFile("businessLicense"); err == nil {
		// TODO: upload to OSS and get URL
		supplier.BusinessLicenseURL = "/uploads/" + file.Filename
	}

	// ID card front
	if _, err := c.FormFile("legalPersonIdCardFront"); err == nil {
		// TODO: upload to OSS
	}

	// ID card back
	if _, err := c.FormFile("legalPersonIdCardBack"); err == nil {
		// TODO: upload to OSS
	}

	// Travel license
	if file, err := c.FormFile("travelLicense"); err == nil {
		supplier.TravelLicenseURL = "/uploads/" + file.Filename
	}

	return nil
}

// saveQualifications saves uploaded qualification files.
func (h *ApplicationHandler) saveQualifications(c *gin.Context, tenantID, supplierID int64) {
	quals := []model.SupplierQualification{}

	if file, err := c.FormFile("businessLicense"); err == nil {
		quals = append(quals, model.SupplierQualification{
			TenantID:          tenantID,
			SupplierID:        supplierID,
			QualificationType: model.QualificationTypeBusinessLicense,
			FileURL:           "/uploads/" + file.Filename,
			FileName:          file.Filename,
			Status:            model.QualificationStatusPending,
		})
	}

	if file, err := c.FormFile("legalPersonIdCardFront"); err == nil {
		quals = append(quals, model.SupplierQualification{
			TenantID:          tenantID,
			SupplierID:        supplierID,
			QualificationType: model.QualificationTypeIDCardFront,
			FileURL:           "/uploads/" + file.Filename,
			FileName:          file.Filename,
			Status:            model.QualificationStatusPending,
		})
	}

	if file, err := c.FormFile("legalPersonIdCardBack"); err == nil {
		quals = append(quals, model.SupplierQualification{
			TenantID:          tenantID,
			SupplierID:        supplierID,
			QualificationType: model.QualificationTypeIDCardBack,
			FileURL:           "/uploads/" + file.Filename,
			FileName:          file.Filename,
			Status:            model.QualificationStatusPending,
		})
	}

	if file, err := c.FormFile("travelLicense"); err == nil {
		quals = append(quals, model.SupplierQualification{
			TenantID:          tenantID,
			SupplierID:        supplierID,
			QualificationType: model.QualificationTypeTravelLicense,
			FileURL:           "/uploads/" + file.Filename,
			FileName:          file.Filename,
			Status:            model.QualificationStatusPending,
		})
	}

	if len(quals) > 0 {
		if err := h.qualRepo.CreateBatch(quals); err != nil {
			h.logger.Error("failed to save qualifications", zap.Error(err))
		}
	}
}

// GetApplicationStatus handles GET /api/v2/suppliers/apply/:applicationNo.
func (h *ApplicationHandler) GetApplicationStatus(c *gin.Context) {
	appNo := c.Param("applicationNo")
	if appNo == "" {
		response.BadRequest(c, "application number required")
		return
	}

	supplier, err := h.supplierRepo.FindByApplicationNo(appNo)
	if err != nil {
		response.NotFound(c, "application not found")
		return
	}

	// Get qualifications
	quals, _ := h.qualRepo.FindBySupplierID(supplier.TenantID, supplier.ID)

	response.OK(c, gin.H{
		"applicationNo": supplier.ApplicationNo,
		"companyName":   supplier.CompanyName,
		"status":        supplier.Status,
		"appliedAt":     supplier.AppliedAt,
		"approvedAt":    supplier.ApprovedAt,
		"qualifications": quals,
	})
}

// parseID parses an integer ID from URL parameter.
func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}
