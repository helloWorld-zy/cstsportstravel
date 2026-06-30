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

// AgreementHandler handles agreement signing operations.
type AgreementHandler struct {
	distributorRepo *repository.DistributorRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewAgreementHandler creates a new AgreementHandler.
func NewAgreementHandler(
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *AgreementHandler {
	return &AgreementHandler{
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
	}
}

// SignAgreement handles POST /api/v2/distributor/agreement/sign
// PRD §8.2.2: 分销商首次登录完成协议签署
func (h *AgreementHandler) SignAgreement(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	distributor, err := h.distributorRepo.FindByUserID(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Distributor not found"})
			return
		}
		h.logger.Error("failed to find distributor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Only pending distributors can sign agreement
	if distributor.Status != domain.DistributorStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Agreement already signed or distributor not in pending status",
		})
		return
	}

	// Record signing time and IP
	now := time.Now()
	clientIP := c.ClientIP()

	distributor.AgreementSignedAt = &now
	distributor.AgreementSignedIP = clientIP
	distributor.Status = domain.DistributorStatusActive
	distributor.UpdatedAt = now

	if err := h.distributorRepo.Update(distributor); err != nil {
		h.logger.Error("failed to sign agreement", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	h.logger.Info("distributor agreement signed",
		zap.Int64("distributor_id", distributor.ID),
		zap.String("ip", clientIP),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Agreement signed successfully",
		"data": gin.H{
			"signed_at": now,
			"status":    distributor.Status,
		},
	})
}

// GetAgreementStatus handles GET /api/v2/distributor/agreement/status
func (h *AgreementHandler) GetAgreementStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	distributor, err := h.distributorRepo.FindByUserID(uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Distributor not found"})
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
			"status":             distributor.Status,
			"agreement_signed":   distributor.AgreementSignedAt != nil,
			"agreement_signed_at": distributor.AgreementSignedAt,
		},
	})
}
