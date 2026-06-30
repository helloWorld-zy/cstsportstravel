// Package handler provides HTTP handlers for the Distribution domain.
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// ApplicationHandler handles distributor application requests.
type ApplicationHandler struct {
	distributorRepo *repository.DistributorRepository
	relationRepo    *repository.DistributorRelationRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewApplicationHandler creates a new ApplicationHandler.
func NewApplicationHandler(
	distributorRepo *repository.DistributorRepository,
	relationRepo *repository.DistributorRelationRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *ApplicationHandler {
	return &ApplicationHandler{
		distributorRepo: distributorRepo,
		relationRepo:    relationRepo,
		db:              db,
		logger:          logger,
	}
}

// SubmitApplicationRequest represents the request body for submitting a distributor application.
type SubmitApplicationRequest struct {
	DistributorType   string `form:"distributorType" binding:"required,oneof=personal enterprise"`
	RealName          string `form:"realName"`
	IDCardNumber      string `form:"idCardNumber"`
	EnterpriseName    string `form:"enterpriseName"`
	CreditCode        string `form:"creditCode"`
	BankName          string `form:"bankName" binding:"required"`
	BankAccountName   string `form:"bankAccountName" binding:"required"`
	BankAccountNumber string `form:"bankAccountNumber" binding:"required"`
	Phone             string `form:"phone" binding:"required"`
	Email             string `form:"email"`
	PromotionChannel  string `form:"promotionChannel"`
	InviteCode        string `form:"inviteCode"`
}

// SubmitApplication handles POST /api/v2/distributors/apply
func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
	var req SubmitApplicationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	// Validate based on distributor type
	if req.DistributorType == domain.DistributorTypePersonal {
		if req.RealName == "" || req.IDCardNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Personal distributor requires realName and idCardNumber"})
			return
		}
		// TODO: Validate ID card checksum
	} else if req.DistributorType == domain.DistributorTypeEnterprise {
		if req.EnterpriseName == "" || req.CreditCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Enterprise distributor requires enterpriseName and creditCode"})
			return
		}
		// TODO: Validate business license via API
	}

	// TODO: Validate bank card via UnionPay BIN check

	// Check if phone already registered
	existing, _ := h.distributorRepo.FindByPhone(req.Phone)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "Phone number already registered as distributor"})
		return
	}

	// Generate distributor number
	distNo, err := h.distributorRepo.GenerateDistributorNo()
	if err != nil {
		h.logger.Error("failed to generate distributor number", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Generate invite code
	inviteCode, err := h.distributorRepo.GenerateInviteCode()
	if err != nil {
		h.logger.Error("failed to generate invite code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Get tenant ID from context (set by middleware)
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	// Get user ID from context
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	distributor := &domain.Distributor{
		TenantID:           tid,
		UserID:             uid,
		DistributorNo:      distNo,
		DistributorType:    req.DistributorType,
		Level:              domain.DistributorLevel1,
		Grade:              domain.DistributorGradeNormal,
		Status:             domain.DistributorStatusPending,
		RealName:           req.RealName,
		IDCardNumber:       req.IDCardNumber, // TODO: Encrypt with AES-256-GCM
		EnterpriseName:     req.EnterpriseName,
		CreditCode:         req.CreditCode,
		BankName:           req.BankName,
		BankAccountName:    req.BankAccountName,
		BankAccountNumber:  req.BankAccountNumber, // TODO: Encrypt with AES-256-GCM
		Phone:              req.Phone,
		Email:              req.Email,
		PromotionChannel:   req.PromotionChannel,
		InviteCode:         inviteCode,
		TotalCommission:    0,
		WithdrawableAmount: 0,
		FrozenAmount:       0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(distributor).Error; err != nil {
			return err
		}

		// If invite code provided, create relation
		if req.InviteCode != "" {
			parent, err := h.distributorRepo.FindByInviteCode(req.InviteCode)
			if err != nil {
				return err
			}

			// Create level 2 relation
			relation := &domain.DistributorRelation{
				TenantID:      tid,
				DistributorID: distributor.ID,
				ParentID:      &parent.ID,
				Level:         domain.DistributorLevel2,
				BindTime:      time.Now(),
				Status:        domain.RelationStatusActive,
				CreatedAt:     time.Now(),
			}
			if err := tx.Create(relation).Error; err != nil {
				return err
			}

			// Update distributor level to 2
			distributor.Level = domain.DistributorLevel2
			if err := tx.Save(distributor).Error; err != nil {
				return err
			}
		} else {
			// Create level 1 relation (no parent)
			relation := &domain.DistributorRelation{
				TenantID:      tid,
				DistributorID: distributor.ID,
				Level:         domain.DistributorLevel1,
				BindTime:      time.Now(),
				Status:        domain.RelationStatusActive,
				CreatedAt:     time.Now(),
			}
			if err := tx.Create(relation).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		h.logger.Error("failed to create distributor application", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to submit application"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Application submitted successfully",
		"data": gin.H{
			"distributor_no": distributor.DistributorNo,
			"status":         distributor.Status,
		},
	})
}

// GetApplicationStatus handles GET /api/v2/distributors/apply/:distributorNo
func (h *ApplicationHandler) GetApplicationStatus(c *gin.Context) {
	distributorNo := c.Param("distributorNo")
	if distributorNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Distributor number is required"})
		return
	}

	distributor, err := h.distributorRepo.FindByDistributorNo(distributorNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Application not found"})
			return
		}
		h.logger.Error("failed to find distributor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"distributor_no": distributor.DistributorNo,
			"status":         distributor.Status,
			"created_at":     distributor.CreatedAt,
		},
	})
}
