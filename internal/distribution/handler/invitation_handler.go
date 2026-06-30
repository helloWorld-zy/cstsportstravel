package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/distribution/domain"
	"github.com/travel-booking/server/internal/distribution/repository"
)

// InvitationHandler handles invitation mechanism operations.
type InvitationHandler struct {
	distributorRepo *repository.DistributorRepository
	relationRepo    *repository.DistributorRelationRepository
	db              *gorm.DB
	logger          *zap.Logger
}

// NewInvitationHandler creates a new InvitationHandler.
func NewInvitationHandler(
	distributorRepo *repository.DistributorRepository,
	relationRepo *repository.DistributorRelationRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *InvitationHandler {
	return &InvitationHandler{
		distributorRepo: distributorRepo,
		relationRepo:    relationRepo,
		db:              db,
		logger:          logger,
	}
}

// GetInviteInfo handles GET /api/v2/distributor/team/invite
// PRD §8.2.5: 一级分销商可通过邀请机制发展二级分销商
func (h *InvitationHandler) GetInviteInfo(c *gin.Context) {
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

	// Only level 1 distributors can invite
	if distributor.Level != domain.DistributorLevel1 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "Only level 1 distributors can invite"})
		return
	}

	if !distributor.IsActive() {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "Distributor is not active"})
		return
	}

	// Generate invite link
	inviteLink := fmt.Sprintf("https://domain.com/distributor/invite?code=%s", distributor.InviteCode)

	// Count children
	childCount, err := h.relationRepo.CountChildren(distributor.ID)
	if err != nil {
		h.logger.Error("failed to count children", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"invite_link":    inviteLink,
			"invite_code":    distributor.InviteCode,
			"click_count":    0, // TODO: Track invite link clicks
			"conversion_count": childCount,
		},
	})
}

// GetTeamMembers handles GET /api/v2/distributor/team
// PRD §8.5.3: 我的团队模块仅限一级分销商访问
func (h *InvitationHandler) GetTeamMembers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	pageNum := 1
	pageSizeNum := 20
	fmt.Sscanf(page, "%d", &pageNum)
	fmt.Sscanf(pageSize, "%d", &pageSizeNum)

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

	// Only level 1 distributors can view team
	if distributor.Level != domain.DistributorLevel1 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "Only level 1 distributors can view team"})
		return
	}

	// Get children
	children, err := h.relationRepo.FindChildren(distributor.ID)
	if err != nil {
		h.logger.Error("failed to find children", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Build team member list
	type TeamMember struct {
		DistributorID    int64   `json:"distributor_id"`
		Nickname         string  `json:"nickname"`
		RegisteredAt     string  `json:"registered_at"`
		Last30DaysOrders int     `json:"last30_days_orders"`
		Last30DaysAmount float64 `json:"last30_days_amount"`
		ContributedCommission float64 `json:"contributed_commission"`
	}

	members := make([]TeamMember, 0)
	for _, child := range children {
		childDist, err := h.distributorRepo.FindByID(child.TenantID, child.DistributorID)
		if err != nil {
			continue
		}

		members = append(members, TeamMember{
			DistributorID:    childDist.ID,
			Nickname:         childDist.RealName,
			RegisteredAt:     child.BindTime.Format("2006-01-02 15:04:05"),
			Last30DaysOrders: 0, // TODO: Calculate from orders
			Last30DaysAmount: 0, // TODO: Calculate from orders
			ContributedCommission: 0, // TODO: Calculate from commissions
		})
	}

	// Paginate
	total := len(members)
	start := (pageNum - 1) * pageSizeNum
	end := start + pageSizeNum
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"team_summary": gin.H{
				"total_members":             total,
				"today_orders":              0, // TODO: Calculate
				"monthly_amount":            0, // TODO: Calculate
				"total_secondary_commission": 0, // TODO: Calculate
			},
			"members":   members[start:end],
			"total":     total,
			"page":      pageNum,
			"page_size": pageSizeNum,
		},
	})
}
