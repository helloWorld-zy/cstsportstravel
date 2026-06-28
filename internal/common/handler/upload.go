// Package handler provides HTTP handlers for shared services.
package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/common/service"
)

// UploadHandler handles HTTP requests for file upload operations.
type UploadHandler struct {
	uploadSvc *service.UploadService
	logger    *zap.Logger
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uploadSvc *service.UploadService, logger *zap.Logger) *UploadHandler {
	return &UploadHandler{
		uploadSvc: uploadSvc,
		logger:    logger,
	}
}

// UploadImageRequest is the request for server-side image upload.
type UploadImageRequest struct {
	Category string `form:"category"` // e.g., "banner", "product", "avatar"
}

// UploadImage handles POST /api/v1/admin/upload/image.
// For server-side upload (when client cannot use STS directly).
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	category := c.PostForm("category")

	// Validate format
	if err := h.uploadSvc.ValidateImageFormat(header.Filename); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate size
	if err := h.uploadSvc.ValidateImageSize(header.Size); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Generate upload key
	key := h.uploadSvc.GenerateUploadKey(header.Filename, category)
	url := h.uploadSvc.GetCDNURL(key)

	// In a real implementation, upload the file to OSS here
	// For MVP, return the generated URL
	result := service.UploadResult{
		URL:      url,
		Key:      key,
		Filename: header.Filename,
		Size:     header.Size,
		MimeType: header.Header.Get("Content-Type"),
	}

	h.logger.Info("image uploaded",
		zap.String("key", key),
		zap.String("filename", header.Filename),
		zap.Int64("size", header.Size),
	)

	response.OK(c, result)
}

// GetSTSTokenRequest is the request for STS token generation.
type GetSTSTokenRequest struct {
	Category string `json:"category" binding:"omitempty"` // e.g., "banner", "product"
}

// GetSTSToken handles POST /api/v1/admin/upload/sts-token.
// Returns temporary credentials for client-side upload to OSS.
func (h *UploadHandler) GetSTSToken(c *gin.Context) {
	var req GetSTSTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Category is optional, default to empty
		req.Category = ""
	}

	token, err := h.uploadSvc.GenerateSTSToken(req.Category)
	if err != nil {
		h.logger.Error("failed to generate STS token", zap.Error(err))
		response.ServerError(c, "failed to generate upload token")
		return
	}

	response.OK(c, token)
}
