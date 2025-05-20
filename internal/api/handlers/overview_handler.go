package handlers

import (
	"go-backend/internal/service"
	"go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type OverviewHandler struct {
	deviceService  *service.DeviceService
	networkService *service.NetworkService
	userService    *service.UserService
}

func NewOverviewHandler(deviceService *service.DeviceService, networkService *service.NetworkService, userService *service.UserService) *OverviewHandler {
	return &OverviewHandler{
		deviceService:  deviceService,
		networkService: networkService,
		userService:    userService,
	}
}

// GetOverview godoc
// @Summary 获取系统概览信息
// @Description 获取系统中的设备数量、节点数量等静态指标信息
// @Tags 系统概览
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=service.OverviewStats}
// @Router /overview [get]
func (h *OverviewHandler) GetOverview(c *gin.Context) {
	stats, err := h.deviceService.GetOverviewStats()
	if err != nil {
		utils.Error(c, utils.ERROR, "获取系统概览信息失败")
		return
	}

	utils.Success(c, stats)
}
