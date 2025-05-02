package handlers

import (
	"strconv"

	"go-backend/internal/models"
	"go-backend/internal/service"
	"go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	deviceService *service.DeviceService
}

func NewDeviceHandler(deviceService *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// ListDevices godoc
// @Summary 获取设备列表
// @Description 获取所有设备列表，支持分页和筛选
// @Tags 设备管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param current query int false "页码(默认1)"
// @Param size query int false "每页数量(默认10)"
// @Param device_type query string false "设备类型筛选"
// @Param status query string false "设备状态筛选"
// @Success 200 {object} utils.Response{data=utils.PageResult{records=[]models.Device}}
// @Router /devices [get]
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	// 获取分页参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 构建过滤条件
	filters := make(map[string]interface{})
	if deviceType := c.Query("device_type"); deviceType != "" {
		filters["device_type"] = deviceType
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	// 获取设备列表
	devices, total, err := h.deviceService.ListDevices(current, size, filters)
	if err != nil {
		utils.Error(c, utils.ERROR, "获取设备列表失败")
		return
	}

	utils.SuccessWithPage(c, devices, current, size, total)
}

// GetDevice godoc
// @Summary 获取设备详情
// @Description 根据ID获取设备详细信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "设备ID"
// @Success 200 {object} utils.Response{data=models.Device}
// @Failure 404 {object} utils.Response
// @Router /devices/{id} [get]
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的设备ID")
		return
	}

	device, err := h.deviceService.GetDevice(uint(id))
	if err != nil {
		utils.Error(c, utils.NOT_FOUND, "设备不存在")
		return
	}

	utils.Success(c, device)
}

// CreateDevice godoc
// @Summary 创建设备
// @Description 创建新的设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param device body models.Device true "设备信息"
// @Success 201 {object} utils.Response{data=models.Device}
// @Failure 400 {object} utils.Response
// @Router /devices [post]
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的设备数据")
		return
	}

	if err := h.deviceService.CreateDevice(&device); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, device, "设备创建成功")
}

// UpdateDevice godoc
// @Summary 更新设备
// @Description 更新设备信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "设备ID"
// @Param device body models.Device true "设备信息"
// @Success 200 {object} utils.Response{data=models.Device}
// @Failure 400,404 {object} utils.Response
// @Router /devices/{id} [put]
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的设备ID")
		return
	}

	var device models.Device
	if err := c.ShouldBindJSON(&device); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的设备数据")
		return
	}
	device.ID = uint(id)

	if err := h.deviceService.UpdateDevice(&device); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, device, "设备更新成功")
}

// DeleteDevice godoc
// @Summary 删除设备
// @Description 删除指定ID的设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "设备ID"
// @Success 200 {object} utils.Response
// @Failure 400,404 {object} utils.Response
// @Router /devices/{id} [delete]
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的设备ID")
		return
	}

	if err := h.deviceService.DeleteDevice(uint(id)); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "设备删除成功")
}
