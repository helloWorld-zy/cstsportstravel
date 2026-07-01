package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/marketing/domain"
	"github.com/travel-booking/server/internal/marketing/repository"
)

// ActivityHandler handles promotion activity CRUD requests.
type ActivityHandler struct {
	activityRepo *repository.PromotionActivityRepository
	db           *gorm.DB
	logger       *zap.Logger
}

// NewActivityHandler creates a new ActivityHandler.
func NewActivityHandler(
	activityRepo *repository.PromotionActivityRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *ActivityHandler {
	return &ActivityHandler{
		activityRepo: activityRepo,
		db:           db,
		logger:       logger,
	}
}

// ActivityCreateRequest represents the request body for creating a promotion activity.
type ActivityCreateRequest struct {
	ActivityName        string      `json:"activityName" binding:"required"`
	ActivityType        string      `json:"activityType" binding:"required,oneof=flash_sale full_reduction early_bird"`
	StartTime           string      `json:"startTime" binding:"required"`
	EndTime             string      `json:"endTime" binding:"required"`
	ApplicableProducts  []int64     `json:"applicableProducts"`
	ApplicableCategories []int64    `json:"applicableCategories"`
	Rules               interface{} `json:"rules" binding:"required"`
	ActivityStock       *int        `json:"activityStock"`
	PerUserLimit        *int        `json:"perUserLimit"`
	StackableWithCoupon bool        `json:"stackableWithCoupon"`
}

// CreateActivity handles POST /api/v2/admin/marketing/activities
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req ActivityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)
	createdBy, _ := c.Get("user_id")
	cid, _ := createdBy.(int64)

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid startTime format"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid endTime format"})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "endTime must be after startTime"})
		return
	}

	// Marshal rules to JSON
	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid rules format"})
		return
	}

	activity := &domain.PromotionActivity{
		TenantID:             tid,
		ActivityName:         req.ActivityName,
		ActivityType:         req.ActivityType,
		StartTime:            startTime,
		EndTime:              endTime,
		ApplicableProducts:   req.ApplicableProducts,
		ApplicableCategories: req.ApplicableCategories,
		Rules:                domain.JSONB(rulesJSON),
		ActivityStock:        req.ActivityStock,
		PerUserLimit:         req.PerUserLimit,
		StackableWithCoupon:  req.StackableWithCoupon,
		Status:               domain.ActivityStatusDraft,
		CreatedBy:            cid,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := h.activityRepo.Create(activity); err != nil {
		h.logger.Error("failed to create activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Activity created successfully",
		"data":    activity,
	})
}

// ListActivities handles GET /api/v2/admin/marketing/activities
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	activityType := c.Query("type")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	activities, total, err := h.activityRepo.List(tid, activityType, status, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list activities", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"list":     activities,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// GetActivityDetail handles GET /api/v2/admin/marketing/activities/:id
func (h *ActivityHandler) GetActivityDetail(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid activity ID"})
		return
	}

	activity, err := h.activityRepo.FindByID(tid, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Activity not found"})
			return
		}
		h.logger.Error("failed to find activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    activity,
	})
}

// UpdateActivity handles PUT /api/v2/admin/marketing/activities/:id
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid activity ID"})
		return
	}

	activity, err := h.activityRepo.FindByID(tid, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Activity not found"})
			return
		}
		h.logger.Error("failed to find activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Only allow updating draft activities
	if activity.Status != domain.ActivityStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Can only update draft activities"})
		return
	}

	var req ActivityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid startTime format"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid endTime format"})
		return
	}

	rulesJSON, err := json.Marshal(req.Rules)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid rules format"})
		return
	}

	activity.ActivityName = req.ActivityName
	activity.ActivityType = req.ActivityType
	activity.StartTime = startTime
	activity.EndTime = endTime
	activity.ApplicableProducts = req.ApplicableProducts
	activity.ApplicableCategories = req.ApplicableCategories
	activity.Rules = domain.JSONB(rulesJSON)
	activity.ActivityStock = req.ActivityStock
	activity.PerUserLimit = req.PerUserLimit
	activity.StackableWithCoupon = req.StackableWithCoupon

	if err := h.activityRepo.Update(activity); err != nil {
		h.logger.Error("failed to update activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to update activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Activity updated successfully",
		"data":    activity,
	})
}

// CancelActivity handles POST /api/v2/admin/marketing/activities/:id/cancel
func (h *ActivityHandler) CancelActivity(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(int64)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid activity ID"})
		return
	}

	if err := h.activityRepo.UpdateStatus(tid, id, domain.ActivityStatusCancelled); err != nil {
		h.logger.Error("failed to cancel activity", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Activity cancelled successfully",
	})
}
