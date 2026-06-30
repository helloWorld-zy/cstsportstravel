package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// AuditHandler handles distributor audit operations.
type AuditHandler struct {
	distributorRepo *repository.DistributorRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *AuditHandler {
	return &AuditHandler{
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
	}
}

// AuditRequest represents the request body for auditing a distributor application.
type AuditRequest struct {
	Action         string   `json:"action" binding:"required,oneof=approve reject supplement"`
	Reason         string   `json:"reason"`
	SupplementItems []string `json:"supplementItems"`
}

// AuditApplication handles POST /api/v2/admin/distributors/:id/audit
func (h *AuditHandler) AuditApplication(c *gin.Context) {
	id := c.GetInt64("id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid distributor ID"})
		return
	}

	var req AuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	distributor, err := h.distributorRepo.FindByID(tid, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Distributor not found"})
			return
		}
		h.logger.Error("failed to find distributor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	if distributor.Status != domain.DistributorStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Application is not in pending status"})
		return
	}

	switch req.Action {
	case "approve":
		// Generate distributor number if not exists (should already exist)
		if distributor.DistributorNo == "" {
			distNo, err := h.distributorRepo.GenerateDistributorNo()
			if err != nil {
				h.logger.Error("failed to generate distributor number", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
				return
			}
			distributor.DistributorNo = distNo
		}

		// Set status to active (after approval, distributor needs to sign agreement)
		// Actually, per PRD §8.2.2, after approval status should be "pending" until agreement signed
		// But the data model doesn't have a "pending_activation" state
		// We'll use "pending" status and the agreement signing will transition to "active"
		distributor.Status = domain.DistributorStatusPending
		distributor.UpdatedAt = time.Now()

		if err := h.distributorRepo.Update(distributor); err != nil {
			h.logger.Error("failed to approve distributor", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
			return
		}

		// TODO: Send approval notification (SMS + in-app)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Application approved",
			"data": gin.H{
				"distributor_no": distributor.DistributorNo,
				"status":         distributor.Status,
			},
		})

	case "reject":
		if req.Reason == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Rejection reason is required"})
			return
		}

		distributor.Status = domain.DistributorStatusCancelled
		distributor.FrozenReason = req.Reason
		distributor.UpdatedAt = time.Now()

		if err := h.distributorRepo.Update(distributor); err != nil {
			h.logger.Error("failed to reject distributor", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
			return
		}

		// TODO: Send rejection notification

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Application rejected",
		})

	case "supplement":
		if len(req.SupplementItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Supplement items are required"})
			return
		}

		// TODO: Update application status to "supplement_required"
		// TODO: Send supplement notification with 7-day deadline

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Supplement request sent",
			"data": gin.H{
				"supplement_items": req.SupplementItems,
				"deadline":         time.Now().AddDate(0, 0, 7),
			},
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid action"})
	}
}

// ListApplications handles GET /api/v2/admin/distributors/applications
func (h *AuditHandler) ListApplications(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")

	pageNum := 1
	pageSizeNum := 20
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &pageSizeNum)

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	distributors, total, err := h.distributorRepo.ListByStatus(tid, status, pageNum, pageSizeNum)
	if err != nil {
		h.logger.Error("failed to list applications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items":     distributors,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSizeNum,
		},
	})
}
