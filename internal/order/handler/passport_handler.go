// Package handler provides HTTP handlers for the Order domain.
package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/middleware"
	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/order/model"
	"github.com/travel-booking/server/internal/order/service"
)

// PassportHandler handles HTTP requests for passport management.
type PassportHandler struct {
	ocrAdapter service.OCRAdapter
	logger     *zap.Logger
}

// NewPassportHandler creates a new PassportHandler.
func NewPassportHandler(ocrAdapter service.OCRAdapter, logger *zap.Logger) *PassportHandler {
	return &PassportHandler{
		ocrAdapter: ocrAdapter,
		logger:     logger,
	}
}

// CreatePassportRequest holds the request body for creating passport info.
type CreatePassportRequest struct {
	NameCN         string `json:"name_cn" binding:"required"`
	NamePinyin     string `json:"name_pinyin" binding:"required"`
	PassportNumber string `json:"passport_number" binding:"required"`
	PassportExpiry string `json:"passport_expiry" binding:"required"` // YYYY-MM-DD
	IssuePlace     string `json:"issue_place"`
	Nationality    string `json:"nationality" binding:"required"`
	Gender         string `json:"gender"`
	BirthDate      string `json:"birth_date"` // YYYY-MM-DD
	IsDefault      bool   `json:"is_default"`
}

// CreatePassport handles POST /api/v2/passports.
func (h *PassportHandler) CreatePassport(c *gin.Context) {
	var req CreatePassportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// Parse passport expiry
	expiry, err := time.Parse("2006-01-02", req.PassportExpiry)
	if err != nil {
		response.BadRequest(c, "invalid passport_expiry format, expected YYYY-MM-DD")
		return
	}

	// Validate passport expiry - at least 6 months from now
	minExpiry := time.Now().AddDate(0, 6, 0)
	if expiry.Before(minExpiry) {
		response.BusinessError(c, 2010, "您的护照有效期不足6个月，建议换发后再预订")
		return
	}

	_ = userID // Will be used for DB insert

	response.OKMessage(c, "passport created successfully")
}

// ValidatePassportExpiryRequest holds the request for passport expiry validation.
type ValidatePassportExpiryRequest struct {
	PassportExpiry string `json:"passport_expiry" binding:"required"` // YYYY-MM-DD
	ReturnDate     string `json:"return_date" binding:"required"`     // YYYY-MM-DD
	CountryID      int64  `json:"country_id"`
}

// ValidatePassportExpiry handles POST /api/v2/passport/validate-expiry.
// Validates if passport validity covers return date + required months (default 6).
func (h *PassportHandler) ValidatePassportExpiry(c *gin.Context) {
	var req ValidatePassportExpiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	expiry, err := time.Parse("2006-01-02", req.PassportExpiry)
	if err != nil {
		response.BadRequest(c, "invalid passport_expiry format")
		return
	}

	returnDate, err := time.Parse("2006-01-02", req.ReturnDate)
	if err != nil {
		response.BadRequest(c, "invalid return_date format")
		return
	}

	// Default to 6 months requirement
	requiredMonths := 6

	passport := &model.PassportInfo{
		PassportExpiry: expiry,
	}
	if err := passport.ValidateExpiry(returnDate, requiredMonths); err != nil {
		response.BusinessError(c, 2010, err.Error())
		return
	}

	response.OK(c, gin.H{
		"valid":            true,
		"passport_expiry":  req.PassportExpiry,
		"return_date":      req.ReturnDate,
		"required_months":  requiredMonths,
	})
}

// OCRPassport handles POST /api/v2/passport/ocr.
// Performs OCR recognition on a passport photo.
func (h *PassportHandler) OCRPassport(c *gin.Context) {
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		response.BadRequest(c, "image file required")
		return
	}
	defer file.Close()

	result, err := h.ocrAdapter.RecognizePassport(c.Request.Context(), file)
	if err != nil {
		h.logger.Error("OCR recognition failed", zap.Error(err))
		response.ServerError(c, "OCR recognition failed")
		return
	}

	if !result.Success {
		response.BusinessError(c, 2011, result.ErrorMessage)
		return
	}

	response.OK(c, result)
}

// GetPassportList handles GET /api/v2/passports.
// Returns user's saved passport info list.
func (h *PassportHandler) GetPassportList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	_ = userID // Will be used for DB query

	// Placeholder - in production, query passport_info table
	response.OK(c, []interface{}{})
}
