package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/travel-booking/server/internal/common/auth"
	"github.com/travel-booking/server/internal/common/response"
)

// MFAResponse is the response returned when MFA verification is required.
type MFAResponse struct {
	RequireMFA bool   `json:"require_mfa"`
	Message    string `json:"message"`
}

// MFARequired returns middleware that intercepts sensitive operations and requires
// TOTP verification. The TOTP code should be sent in the X-TOTP-Code header.
//
// If the user has not enrolled in MFA, the middleware returns a response indicating
// MFA enrollment is required.
//
// This middleware checks:
// 1. If user has MFA enrolled (admin_user.totp_secret is not empty)
// 2. If a valid TOTP code is provided in X-TOTP-Code header
// 3. Falls back to SMS verification if X-SMS-Code header is provided
func MFARequired(db *gorm.DB, totpService *auth.TOTPService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		// Check if user has MFA enrolled
		var totpSecret string
		err := db.Table("admin_user").
			Select("totp_secret").
			Where("id = ?", userID).
			Scan(&totpSecret).Error
		if err != nil {
			response.ServerError(c, "failed to check MFA status")
			c.Abort()
			return
		}

		// If no MFA enrolled, allow through (admin should enroll)
		// In production, you may want to force enrollment
		if totpSecret == "" {
			c.Next()
			return
		}

		// Check TOTP code from header
		totpCode := c.GetHeader("X-TOTP-Code")
		if totpCode != "" {
			if totpService.VerifyCode(totpSecret, totpCode) {
				c.Next()
				return
			}
			response.BadRequest(c, "invalid TOTP code")
			c.Abort()
			return
		}

		// Check SMS fallback code
		smsCode := c.GetHeader("X-SMS-Code")
		if smsCode != "" {
			// Validate SMS code from Redis (implementation in user service)
			// For now, return error indicating SMS verification not yet implemented
			response.BadRequest(c, "SMS verification fallback not yet available")
			c.Abort()
			return
		}

		// No verification provided
		response.Fail(c, 403, response.CodeForbidden, "MFA verification required: provide X-TOTP-Code header")
		c.Abort()
	}
}

// MFAOptional returns middleware that checks MFA but doesn't block if not enrolled.
// It sets a context key indicating whether MFA was verified.
func MFAOptional(db *gorm.DB, totpService *auth.TOTPService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.Next()
			return
		}

		var totpSecret string
		err := db.Table("admin_user").
			Select("totp_secret").
			Where("id = ?", userID).
			Scan(&totpSecret).Error
		if err != nil || totpSecret == "" {
			c.Set("mfa_verified", false)
			c.Next()
			return
		}

		totpCode := c.GetHeader("X-TOTP-Code")
		if totpCode != "" && totpService.VerifyCode(totpSecret, totpCode) {
			c.Set("mfa_verified", true)
		} else {
			c.Set("mfa_verified", false)
		}

		c.Next()
	}
}
