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

// AdminDistributorHandler handles admin distributor management operations.
type AdminDistributorHandler struct {
	distributorRepo *repository.DistributorRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewAdminDistributorHandler creates a new AdminDistributorHandler.
func NewAdminDistributorHandler(
	distributorRepo *repository.DistributorRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *AdminDistributorHandler {
	return &AdminDistributorHandler{
		distributorRepo: distributorRepo,
		db:              db,
		logger:          logger,
	}
}

// ListDistributors handles GET /api/v2/admin/distributors
// PRD §8.6.1: 全平台分销商列表
func (h *AdminDistributorHandler) ListDistributors(c *gin.Context) {
	distributorType := c.DefaultQuery("type", "")
	grade := c.DefaultQuery("grade", "")
	status := c.DefaultQuery("status", "")
	keyword := c.DefaultQuery("keyword", "")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	pageNum := 1
	pageSizeNum := 20
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &pageSizeNum)

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	distributors, total, err := h.distributorRepo.ListByTypeAndGrade(tid, distributorType, grade, status, keyword, pageNum, pageSizeNum)
	if err != nil {
		h.logger.Error("failed to list distributors", zap.Error(err))
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

// UpdateDistributorStatusRequest represents the request body for updating distributor status.
type UpdateDistributorStatusRequest struct {
	Action string `json:"action" binding:"required,oneof=freeze unfreeze cancel"`
	Reason string `json:"reason"`
}

// UpdateDistributorStatus handles PUT /api/v2/admin/distributors/:id/status
func (h *AdminDistributorHandler) UpdateDistributorStatus(c *gin.Context) {
	id := c.GetInt64("id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid distributor ID"})
		return
	}

	var req UpdateDistributorStatusRequest
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

	var targetStatus string
	switch req.Action {
	case "freeze":
		targetStatus = domain.DistributorStatusFrozen
	case "unfreeze":
		targetStatus = domain.DistributorStatusActive
	case "cancel":
		targetStatus = domain.DistributorStatusCancelled
	}

	if !distributor.CanTransitionTo(targetStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("Cannot transition from %s to %s", distributor.Status, targetStatus),
		})
		return
	}

	distributor.Status = targetStatus
	if req.Reason != "" {
		distributor.FrozenReason = req.Reason
	}
	if targetStatus == domain.DistributorStatusFrozen {
		distributor.FrozenUntil = nil // Indefinite freeze
	}
	distributor.UpdatedAt = time.Now()

	if err := h.distributorRepo.Update(distributor); err != nil {
		h.logger.Error("failed to update distributor status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Distributor status updated",
		"data": gin.H{
			"status": distributor.Status,
		},
	})
}

// UpdateDistributorGradeRequest represents the request body for updating distributor grade.
type UpdateDistributorGradeRequest struct {
	Grade string `json:"grade" binding:"required,oneof=normal senior"`
}

// UpdateDistributorGrade handles PUT /api/v2/admin/distributors/:id/grade
func (h *AdminDistributorHandler) UpdateDistributorGrade(c *gin.Context) {
	id := c.GetInt64("id")
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid distributor ID"})
		return
	}

	var req UpdateDistributorGradeRequest
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

	distributor.Grade = req.Grade
	if req.Grade == domain.DistributorGradeSenior {
		validUntil := time.Now().AddDate(0, 0, 90)
		distributor.GradeValidUntil = &validUntil
	} else {
		distributor.GradeValidUntil = nil
	}
	distributor.UpdatedAt = time.Now()

	if err := h.distributorRepo.Update(distributor); err != nil {
		h.logger.Error("failed to update distributor grade", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Distributor grade updated",
		"data": gin.H{
			"grade": distributor.Grade,
		},
	})
}
