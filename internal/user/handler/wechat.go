package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/travel-booking/server/internal/common/response"
	"github.com/travel-booking/server/internal/user/service"
)

// WechatHandler handles HTTP requests for WeChat OAuth login.
type WechatHandler struct {
	wechatService *service.WechatService
	logger        *zap.Logger
}

// NewWechatHandler creates a new WechatHandler.
func NewWechatHandler(wechatService *service.WechatService, logger *zap.Logger) *WechatHandler {
	return &WechatHandler{
		wechatService: wechatService,
		logger:        logger,
	}
}

// Login handles POST /api/v1/auth/wechat.
func (h *WechatHandler) Login(c *gin.Context) {
	var req service.WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "code is required")
		return
	}

	result, err := h.wechatService.Login(req)
	if err != nil {
		h.logger.Error("wechat login failed", zap.Error(err))
		response.Unauthorized(c, "wechat login failed")
		return
	}

	response.OK(c, result)
}
