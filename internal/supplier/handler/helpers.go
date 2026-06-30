package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/travel-booking/server/internal/shared/middleware"
)

// getSupplierID extracts the supplier ID from Gin context.
func getSupplierID(c *gin.Context) int64 {
	if sid, exists := c.Get(middleware.ContextKeySupplierID); exists {
		if id, ok := sid.(int64); ok {
			return id
		}
		if s, ok := sid.(string); ok {
			if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				return id
			}
		}
	}
	return 0
}

// parseIntDefault parses an integer query parameter with a default value.
func parseIntDefault(c *gin.Context, key string, defaultVal int) int {
	s := c.DefaultQuery(key, "")
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
