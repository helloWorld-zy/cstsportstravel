// Package handler provides HTTP handlers for the Admin domain.
package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/admin/service"
	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
)

// AdminBannerHandler handles HTTP requests for banner management.
type AdminBannerHandler struct {
	bannerSvc *service.BannerService
	logger    *zap.Logger
}

// NewAdminBannerHandler creates a new AdminBannerHandler.
func NewAdminBannerHandler(bannerSvc *service.BannerService, logger *zap.Logger) *AdminBannerHandler {
	return &AdminBannerHandler{
		bannerSvc: bannerSvc,
		logger:    logger,
	}
}

// ListBanners handles GET /api/v1/admin/banners.
func (h *AdminBannerHandler) ListBanners(c *gin.Context) {
	position := c.Query("position")
	status := c.Query("status")

	banners, err := h.bannerSvc.ListBanners(position, status)
	if err != nil {
		h.logger.Error("failed to list banners", zap.Error(err))
		response.ServerError(c, "failed to list banners")
		return
	}

	response.OK(c, banners)
}

// GetBanner handles GET /api/v1/admin/banners/:id.
func (h *AdminBannerHandler) GetBanner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid banner id")
		return
	}

	banner, err := h.bannerSvc.GetBanner(id)
	if err != nil {
		if errors.Is(err, service.ErrBannerNotFound) {
			response.NotFound(c, "banner not found")
			return
		}
		h.logger.Error("failed to get banner", zap.Int64("id", id), zap.Error(err))
		response.ServerError(c, "failed to get banner")
		return
	}

	response.OK(c, banner)
}

// CreateBanner handles POST /api/v1/admin/banners.
func (h *AdminBannerHandler) CreateBanner(c *gin.Context) {
	var req service.CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	operatorID := middleware.GetUserID(c)

	banner, err := h.bannerSvc.CreateBanner(operatorID, req)
	if err != nil {
		h.logger.Error("failed to create banner", zap.Error(err))
		response.ServerError(c, "failed to create banner")
		return
	}

	response.OK(c, banner)
}

// UpdateBanner handles PUT /api/v1/admin/banners/:id.
func (h *AdminBannerHandler) UpdateBanner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid banner id")
		return
	}

	var req service.UpdateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	banner, err := h.bannerSvc.UpdateBanner(id, req)
	if err != nil {
		if errors.Is(err, service.ErrBannerNotFound) {
			response.NotFound(c, "banner not found")
			return
		}
		h.logger.Error("failed to update banner", zap.Int64("id", id), zap.Error(err))
		response.ServerError(c, "failed to update banner")
		return
	}

	response.OK(c, banner)
}

// DeleteBanner handles DELETE /api/v1/admin/banners/:id.
func (h *AdminBannerHandler) DeleteBanner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid banner id")
		return
	}

	if err := h.bannerSvc.DeleteBanner(id); err != nil {
		if errors.Is(err, service.ErrBannerNotFound) {
			response.NotFound(c, "banner not found")
			return
		}
		h.logger.Error("failed to delete banner", zap.Int64("id", id), zap.Error(err))
		response.ServerError(c, "failed to delete banner")
		return
	}

	response.OKMessage(c, "banner deleted")
}
