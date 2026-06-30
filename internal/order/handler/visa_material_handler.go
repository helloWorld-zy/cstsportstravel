// Package handler provides HTTP handlers for the Order domain.
package handler

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/order/model"
	"github.com/travel-booking/server/internal/order/repository"
	productrepo "github.com/travel-booking/server/internal/product/repository"
)

// VisaMaterialHandler handles HTTP requests for visa material management.
type VisaMaterialHandler struct {
	materialRepo  *repository.VisaMaterialRepository
	visaOrderRepo *repository.VisaOrderRepository
	templateRepo  *productrepo.VisaMaterialTemplateRepository
	logger        *zap.Logger
}

// NewVisaMaterialHandler creates a new VisaMaterialHandler.
func NewVisaMaterialHandler(
	materialRepo *repository.VisaMaterialRepository,
	visaOrderRepo *repository.VisaOrderRepository,
	templateRepo *productrepo.VisaMaterialTemplateRepository,
	logger *zap.Logger,
) *VisaMaterialHandler {
	return &VisaMaterialHandler{
		materialRepo:  materialRepo,
		visaOrderRepo: visaOrderRepo,
		templateRepo:  templateRepo,
		logger:        logger,
	}
}

// GetMaterialChecklist handles GET /api/v2/visa-orders/:visaOrderId/materials.
// Returns the material checklist based on occupation type.
func (h *VisaMaterialHandler) GetMaterialChecklist(c *gin.Context) {
	visaOrderID, err := parseID(c, "visaOrderId")
	if err != nil {
		response.BadRequest(c, "invalid visa order id")
		return
	}

	// Get the visa order to determine country and occupation
	order, err := h.visaOrderRepo.FindByID(visaOrderID)
	if err != nil {
		response.NotFound(c, "visa order not found")
		return
	}

	userID := middleware.GetUserID(c)
	if userID != order.UserID {
		response.Forbidden(c, "access denied")
		return
	}

	// Get existing materials
	materials, err := h.materialRepo.FindByVisaOrderID(visaOrderID)
	if err != nil {
		h.logger.Error("failed to get materials", zap.Error(err))
		response.ServerError(c, "failed to get materials")
		return
	}

	response.OK(c, gin.H{
		"visa_order_id":   visaOrderID,
		"occupation_type": order.OccupationType,
		"materials":       materials,
	})
}

// UploadMaterialRequest holds the request for uploading a material file.
type UploadMaterialRequest struct {
	File []byte `json:"-"` // handled via multipart form
}

// UploadMaterial handles POST /api/v2/visa-orders/:visaOrderId/materials/:materialId/upload.
// Uploads a file for a specific material (≤10MB).
func (h *VisaMaterialHandler) UploadMaterial(c *gin.Context) {
	visaOrderID, err := parseID(c, "visaOrderId")
	if err != nil {
		response.BadRequest(c, "invalid visa order id")
		return
	}

	materialID, err := parseID(c, "materialId")
	if err != nil {
		response.BadRequest(c, "invalid material id")
		return
	}

	// Get the material to verify ownership
	material, err := h.materialRepo.FindByID(materialID)
	if err != nil {
		response.NotFound(c, "material not found")
		return
	}
	if material.VisaOrderID != visaOrderID {
		response.BadRequest(c, "material does not belong to this visa order")
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file required")
		return
	}
	defer file.Close()

	// Validate file size (≤10MB)
	if header.Size > model.MaxFileSize {
		response.BusinessError(c, 2012, fmt.Sprintf("文件大小超过限制（最大%dMB）", model.MaxFileSize/(1024*1024)))
		return
	}

	// Validate file format
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := false
	for _, format := range model.AllowedFileFormats {
		if ext == format {
			allowed = true
			break
		}
	}
	if !allowed {
		response.BusinessError(c, 2013, "不支持的文件格式，仅支持 JPG/PNG/PDF")
		return
	}

	_ = file // In production, save to storage and get URL

	// Update material record
	fileURL := fmt.Sprintf("/uploads/visa/%d/%d%s", visaOrderID, materialID, ext)
	if err := h.materialRepo.UpdateFileURL(materialID, fileURL, header.Size); err != nil {
		h.logger.Error("failed to update material", zap.Error(err))
		response.ServerError(c, "failed to upload material")
		return
	}

	response.OK(c, gin.H{
		"material_id": materialID,
		"file_url":    fileURL,
		"file_size":   header.Size,
		"status":      model.VisaMaterialStatusSubmitted,
	})
}

// SubmitMaterials handles POST /api/v2/visa-orders/:visaOrderId/materials/submit.
// Submits all materials for review, changes visa order status to reviewing.
func (h *VisaMaterialHandler) SubmitMaterials(c *gin.Context) {
	visaOrderID, err := parseID(c, "visaOrderId")
	if err != nil {
		response.BadRequest(c, "invalid visa order id")
		return
	}

	order, err := h.visaOrderRepo.FindByID(visaOrderID)
	if err != nil {
		response.NotFound(c, "visa order not found")
		return
	}

	userID := middleware.GetUserID(c)
	if userID != order.UserID {
		response.Forbidden(c, "access denied")
		return
	}

	if order.Status != model.VisaStatusPendingSubmit {
		response.BusinessError(c, 2014, "当前状态不允许提交材料")
		return
	}

	// Check completeness
	complete, missing, err := h.materialRepo.CheckCompleteness(visaOrderID)
	if err != nil {
		h.logger.Error("failed to check completeness", zap.Error(err))
		response.ServerError(c, "failed to check materials")
		return
	}

	if !complete {
		response.BusinessError(c, 2015, fmt.Sprintf("请先上传必填材料：%v", missing))
		return
	}

	// Transition to reviewing
	progress, err := order.TransitionTo(model.VisaStatusReviewing, userID, "用户提交材料")
	if err != nil {
		response.BusinessError(c, 2016, err.Error())
		return
	}

	if err := h.visaOrderRepo.UpdateStatus(order, progress); err != nil {
		h.logger.Error("failed to update visa order status", zap.Error(err))
		response.ServerError(c, "failed to submit materials")
		return
	}

	response.OK(c, gin.H{
		"visa_order_id": visaOrderID,
		"status":        order.Status,
		"status_name":   model.VisaStatusName(order.Status),
	})
}

// parseID extracts an int64 ID from a URL parameter.
func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}
