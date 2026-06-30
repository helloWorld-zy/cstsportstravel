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

// PromotionHandler handles promotion link operations.
type PromotionHandler struct {
	distributorRepo   *repository.DistributorRepository
	promotionLinkRepo *repository.PromotionLinkRepository
	db                *gorm.DB
	logger            *zap.Logger
}

// NewPromotionHandler creates a new PromotionHandler.
func NewPromotionHandler(
	distributorRepo *repository.DistributorRepository,
	promotionLinkRepo *repository.PromotionLinkRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *PromotionHandler {
	return &PromotionHandler{
		distributorRepo:   distributorRepo,
		promotionLinkRepo: promotionLinkRepo,
		db:                db,
		logger:            logger,
	}
}

// CreatePromotionLinkRequest represents the request body for creating a promotion link.
type CreatePromotionLinkRequest struct {
	ProductID int64 `json:"productId" binding:"required"`
}

// CreatePromotionLink handles POST /api/v2/distributor/promotion-links
// PRD §8.3.1: 每个分销商可为平台上的任意上架产品生成专属推广链接
func (h *PromotionHandler) CreatePromotionLink(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	var req CreatePromotionLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request: " + err.Error()})
		return
	}

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

	if !distributor.IsActive() {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "Distributor is not active"})
		return
	}

	// Check if link already exists for this product
	existingLink, _ := h.promotionLinkRepo.FindByDistributorAndProduct(distributor.ID, req.ProductID)
	if existingLink != nil {
		// Return existing link
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Promotion link already exists",
			"data": gin.H{
				"short_link":  existingLink.ShortLink,
				"qr_code_url": existingLink.QRCodeURL,
				"qr_code_sizes": []gin.H{
					{"size": 300, "url": fmt.Sprintf("%s?size=300", existingLink.QRCodeURL)},
					{"size": 500, "url": fmt.Sprintf("%s?size=500", existingLink.QRCodeURL)},
					{"size": 800, "url": fmt.Sprintf("%s?size=800", existingLink.QRCodeURL)},
				},
			},
		})
		return
	}

	// Generate short link
	shortLink, err := h.promotionLinkRepo.GenerateShortLink()
	if err != nil {
		h.logger.Error("failed to generate short link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	// Create promotion link
	link := &domain.PromotionLink{
		TenantID:      distributor.TenantID,
		DistributorID: distributor.ID,
		ProductID:     req.ProductID,
		ShortLink:     shortLink,
		QRCodeURL:     fmt.Sprintf("https://domain.com/qr/%s", shortLink), // TODO: Generate actual QR code
		ClickPV:       0,
		ClickUV:       0,
		OrderCount:    0,
		OrderAmount:   0,
		Status:        domain.PromotionLinkStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.promotionLinkRepo.Create(link); err != nil {
		h.logger.Error("failed to create promotion link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Promotion link created",
		"data": gin.H{
			"short_link":  link.ShortLink,
			"qr_code_url": link.QRCodeURL,
			"qr_code_sizes": []gin.H{
				{"size": 300, "url": fmt.Sprintf("%s?size=300", link.QRCodeURL)},
				{"size": 500, "url": fmt.Sprintf("%s?size=500", link.QRCodeURL)},
				{"size": 800, "url": fmt.Sprintf("%s?size=800", link.QRCodeURL)},
			},
		},
	})
}

// ListPromotionLinks handles GET /api/v2/distributor/promotion-links
func (h *PromotionHandler) ListPromotionLinks(c *gin.Context) {
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

	links, total, err := h.promotionLinkRepo.FindByDistributorID(distributor.ID, pageNum, pageSizeNum)
	if err != nil {
		h.logger.Error("failed to list promotion links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data": gin.H{
			"items":     links,
			"total":     total,
			"page":      pageNum,
			"page_size": pageSizeNum,
		},
	})
}
