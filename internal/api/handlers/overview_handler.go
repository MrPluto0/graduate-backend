package handlers

import (
	"go-backend/internal/service"
	"go-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OverviewHandler struct {
	deviceService  *service.DeviceService
	networkService *service.NetworkService
	userService    *service.UserService
	monitorService *service.MonitorService
	alarmService   *service.AlarmService
}

func NewOverviewHandler(
	deviceService *service.DeviceService,
	networkService *service.NetworkService,
	userService *service.UserService,
	monitorService *service.MonitorService,
	alarmService *service.AlarmService,
) *OverviewHandler {
	return &OverviewHandler{
		deviceService:  deviceService,
		networkService: networkService,
		userService:    userService,
		monitorService: monitorService,
		alarmService:   alarmService,
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

// GetSystemMetrics godoc
// @Summary 获取系统监控指标
// @Description 获取系统CPU、内存等监控指标
// @Tags 系统概览
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=models.SystemMetrics}
// @Router /system/metrics [get]
func (h *OverviewHandler) GetSystemMetrics(c *gin.Context) {
	metrics, err := h.monitorService.GetSystemMetrics()
	if err != nil {
		utils.Error(c, utils.ERROR, "获取系统监控指标失败")
		return
	}

	utils.Success(c, metrics)
}

// GetAlarms godoc
// @Summary 获取告警列表
// @Description 获取系统告警信息列表
// @Tags 系统概览
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param current query int false "当前页" default(1)
// @Param size query int false "每页大小" default(10)
// @Param status query string false "告警状态"
// @Success 200 {object} utils.Response{data=utils.PageResult}
// @Router /alarms [get]
func (h *OverviewHandler) GetAlarms(c *gin.Context) {
	// 获取查询参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")

	// 调用alarmService获取数据
	alarms, total, err := h.alarmService.GetAlarmList(current, size, status)
	if err != nil {
		utils.Error(c, utils.ERROR, "获取告警列表失败")
		return
	}

	utils.SuccessWithPage(c, alarms, current, size, total)
}
