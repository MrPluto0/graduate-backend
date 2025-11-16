package handlers

import (
	"go-backend/internal/models"
	"go-backend/internal/service"
	"go-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AlarmHandler struct {
	alarmService *service.AlarmService
}

func NewAlarmHandler(alarmService *service.AlarmService) *AlarmHandler {
	return &AlarmHandler{
		alarmService: alarmService,
	}
}

// GetAlarms godoc
// @Summary 获取告警列表
// @Description 获取系统告警信息列表（支持分页和状态筛选）
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param current query int false "当前页" default(1)
// @Param size query int false "每页大小" default(10)
// @Param status query string false "告警状态(pending/resolved)"
// @Success 200 {object} utils.Response{data=utils.PageResult}
// @Router /alarms [get]
func (h *AlarmHandler) GetAlarms(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")

	alarms, total, err := h.alarmService.GetAlarmList(current, size, status)
	if err != nil {
		utils.Error(c, utils.ERROR, "获取告警列表失败")
		return
	}

	utils.SuccessWithPage(c, alarms, current, size, total)
}

// GetAlarm godoc
// @Summary 获取告警详情
// @Description 根据ID获取单个告警的详细信息
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Success 200 {object} utils.Response{data=models.Alarm}
// @Failure 404 {object} utils.Response
// @Router /alarms/{id} [get]
func (h *AlarmHandler) GetAlarm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的告警ID")
		return
	}

	alarm, err := h.alarmService.GetAlarm(uint(id))
	if err != nil {
		utils.Error(c, utils.NOT_FOUND, "告警不存在")
		return
	}

	utils.Success(c, alarm)
}

// ResolveAlarm godoc
// @Summary 解决告警
// @Description 将告警标记为已解决状态
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /alarms/{id}/resolve [post]
func (h *AlarmHandler) ResolveAlarm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的告警ID")
		return
	}

	if err := h.alarmService.ResolveAlarm(uint(id)); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "告警已解决")
}

// ReactivateAlarm godoc
// @Summary 重新激活告警
// @Description 将已解决的告警重新激活为活跃状态
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /alarms/{id}/reactivate [post]
func (h *AlarmHandler) ReactivateAlarm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的告警ID")
		return
	}

	if err := h.alarmService.ReactivateAlarm(uint(id)); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "告警已重新激活")
}

// DeleteAlarm godoc
// @Summary 删除告警
// @Description 删除指定的告警记录
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /alarms/{id} [delete]
func (h *AlarmHandler) DeleteAlarm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的告警ID")
		return
	}

	if err := h.alarmService.DeleteAlarm(uint(id)); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "告警已删除")
}

// BatchResolveAlarms godoc
// @Summary 批量解决告警
// @Description 批量将多个告警标记为已解决状态
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param ids body []uint true "告警ID列表"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /alarms/batch/resolve [post]
func (h *AlarmHandler) BatchResolveAlarms(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的请求参数")
		return
	}

	if err := h.alarmService.BatchResolveAlarms(ids); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "批量解决成功")
}

// BatchDeleteAlarms godoc
// @Summary 批量删除告警
// @Description 批量删除多个告警记录
// @Tags 告警管理
// @Accept json
// @Produce json
// @Param ids body []uint true "告警ID列表"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /alarms/batch/delete [post]
func (h *AlarmHandler) BatchDeleteAlarms(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的请求参数")
		return
	}

	if err := h.alarmService.BatchDeleteAlarms(ids); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "批量删除成功")
}

// GetAlarmStats godoc
// @Summary 获取告警统计
// @Description 获取告警的统计信息（总数、活跃数、已解决数）
// @Tags 告警管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=service.AlarmStats}
// @Router /alarms/stats [get]
func (h *AlarmHandler) GetAlarmStats(c *gin.Context) {
	stats, err := h.alarmService.GetAlarmStats()
	if err != nil {
		utils.Error(c, utils.ERROR, "获取告警统计失败")
		return
	}

	utils.Success(c, stats)
}
