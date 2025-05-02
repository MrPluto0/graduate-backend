package handlers

import (
	"time"

	"go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// HealthHandler 处理健康检查相关的请求
type HealthHandler struct{}

// NewHealthHandler 创建一个新的健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// CheckHealth godoc
// @Summary      健康检查接口
// @Description  返回API服务的运行状态、版本和时间戳信息
// @Tags         系统监控
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Router       /health [get]
func (h *HealthHandler) CheckHealth(c *gin.Context) {
	status := map[string]interface{}{
		"status":    "up",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "go-backend API",
		"version":   "1.0.0",
	}

	utils.Success(c, status)
}
