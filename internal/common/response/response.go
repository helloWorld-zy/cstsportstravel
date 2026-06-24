package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes
const (
	CodeSuccess       = 0
	CodeBadRequest    = 1001
	CodeUnauthorized  = 1002
	CodeForbidden     = 1003
	CodeNotFound      = 1004
	CodeValidation    = 1005
	CodeConflict      = 1006
	CodeTooManyReq    = 1007

	CodeBusiness      = 2000
	CodeInsufficient  = 2001
	CodePaymentFailed = 2002
	CodeOrderClosed   = 2003
	CodeStockEmpty    = 2004

	CodeServer        = 5000
	CodeDBError       = 5001
	CodeCacheError    = 5002
	CodeThirdParty    = 5003
)

// Response is the unified API response envelope.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id"`
}

// OK sends a success response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
		TraceID: getTraceID(c),
	})
}

// OKMessage sends a success response with a custom message.
func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		TraceID: getTraceID(c),
	})
}

// Fail sends an error response with a specific code.
func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		TraceID: getTraceID(c),
	})
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, CodeNotFound, message)
}

// ServerError sends a 500 error response.
func ServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, CodeServer, message)
}

// BusinessError sends a business logic error response.
func BusinessError(c *gin.Context, code int, message string) {
	Fail(c, http.StatusOK, code, message)
}

// getTraceID extracts trace_id from Gin context, or generates a placeholder.
func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}
