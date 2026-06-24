package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditLog is the GORM model for audit_log table entries.
type AuditLog struct {
	ID           int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OperatorID   *int64          `gorm:"column:operator_id" json:"operator_id"`
	OperatorType string          `gorm:"column:operator_type;size:20;not null" json:"operator_type"`
	Action       string          `gorm:"column:action;size:100;not null" json:"action"`
	TargetType   string          `gorm:"column:target_type;size:50;not null" json:"target_type"`
	TargetID     *int64          `gorm:"column:target_id" json:"target_id"`
	Detail       json.RawMessage `gorm:"column:detail;type:jsonb" json:"detail"`
	IPAddress    string          `gorm:"column:ip_address;size:45" json:"ip_address"`
	UserAgent    string          `gorm:"column:user_agent;size:500" json:"user_agent"`
	CreatedAt    time.Time       `gorm:"column:created_at;not null;default:now()" json:"created_at"`
}

// TableName overrides the table name for AuditLog.
func (AuditLog) TableName() string {
	return "audit_log"
}

// bodyLogWriter wraps the response body to capture it for audit logging.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditMiddleware returns middleware that records write operations (POST/PUT/DELETE/PATCH)
// to the audit_log table with user_id, action, resource, IP address, and request/response details.
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit write operations
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			c.Next()
			return
		}

		// Read request body
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// Capture response body
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// Build audit entry
		operatorID := GetUserID(c)
		operatorType := "system"
		if operatorID > 0 {
			if GetUserType(c) == "admin" {
				operatorType = "admin"
			} else {
				operatorType = "user"
			}
		}

		detail := map[string]interface{}{
			"method":        method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status_code":   c.Writer.Status(),
			"request_body":  truncateBody(reqBody, 4096),
			"response_body": truncateBody(blw.body.Bytes(), 4096),
		}
		detailJSON, _ := json.Marshal(detail)

		audit := AuditLog{
			Action:       method + " " + c.Request.URL.Path,
			TargetType:   extractResourceType(c.Request.URL.Path),
			Detail:       detailJSON,
			IPAddress:    c.ClientIP(),
			UserAgent:    truncateString(c.GetHeader("User-Agent"), 500),
			OperatorType: operatorType,
		}

		if operatorID > 0 {
			audit.OperatorID = &operatorID
		}

		// Write audit log asynchronously to avoid blocking the response
		go func() {
			if err := db.Create(&audit).Error; err != nil {
				log.Printf("ERROR: audit log write failed for %s: %v", audit.Action, err)
			}
		}()
	}
}

// extractResourceType extracts the resource type from the URL path.
// e.g., "/api/v1/admin/products" -> "product"
func extractResourceType(path string) string {
	segments := splitPath(path)
	if len(segments) >= 4 {
		return segments[3] // e.g., "products", "orders"
	}
	return "unknown"
}

// splitPath splits a URL path into segments, ignoring empty segments.
func splitPath(path string) []string {
	var segments []string
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				segments = append(segments, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		segments = append(segments, current)
	}
	return segments
}

// truncateBody truncates a byte slice to maxLen bytes.
func truncateBody(body []byte, maxLen int) string {
	if len(body) > maxLen {
		return string(body[:maxLen]) + "...[truncated]"
	}
	return string(body)
}

// truncateString truncates a string to maxLen characters.
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
