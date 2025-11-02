package handlers

import (
	"fmt"
	"go-backend/internal/algorithm"
	"go-backend/internal/algorithm/define"
	"go-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AlgorithmHandler struct {
	system *algorithm.SystemAdapter
}

func NewAlgorithmHandler() *AlgorithmHandler {
	return &AlgorithmHandler{
		system: algorithm.GetAdaptedSystem(),
	}
}

// StartAlgorithm godoc
// @Summary 提交任务
// @Description 提交计算任务
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body []TaskSubmitRequest true "任务列表"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /algorithm/start [post]
func (h *AlgorithmHandler) StartAlgorithm(c *gin.Context) {
	var requests []define.TaskBase
	if err := c.ShouldBindJSON(&requests); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	if len(requests) == 0 {
		utils.Error(c, utils.VALIDATION_ERROR, "任务列表不能为空")
		return
	}

	tasks, _ := h.system.SubmitBatchTasks(requests)

	utils.SuccessWithMessage(c, tasks, "任务提交成功")
}

// StopAlgorithm godoc
// @Summary 停止算法
// @Description 停止当前运行的算法
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /algorithm/stop [post]
func (h *AlgorithmHandler) StopAlgorithm(c *gin.Context) {
	h.system.StopAlgorithm()
	utils.SuccessWithMessage(c, nil, "算法已停止")
}

// GetSystemInfo godoc
// @Summary 获取系统信息
// @Description 获取当前系统状态信息
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=define.SystemInfo}
// @Router /algorithm/info [get]
func (h *AlgorithmHandler) GetSystemInfo(c *gin.Context) {
	info := h.system.GetSystemInfo()
	utils.Success(c, info)
}

// ClearHistory godoc
// @Summary 清除历史记录
// @Description 清除算法执行的历史状态记录
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response
// @Router /algorithm/clear [post]
func (h *AlgorithmHandler) ClearHistory(c *gin.Context) {
	h.system.ClearHistory()
	utils.SuccessWithMessage(c, nil, "历史记录已清除")
}

// GetTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取所有任务或根据筛选条件获取任务
// @Tags Algorithm
// @Accept json
// @Produce json
// @Param current query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param user_id query int false "用户ID"
// @Param status query string false "任务状态"
// @Success 200 {object} utils.PageResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /algorithm/tasks [get]
func (h *AlgorithmHandler) GetTasks(c *gin.Context) {
	// 获取分页参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	offset := (current - 1) * size

	// 解析筛选条件
	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(id)
			userID = &uid
		}
	}

	var status *define.TaskStatus
	if statusStr := c.Query("status"); statusStr != "" {
		var s int
		if _, err := fmt.Sscanf(statusStr, "%d", &s); err == nil {
			taskStatus := define.TaskStatus(s)
			status = &taskStatus
		}
	}

	// 获取任务列表
	tasks, total := h.system.GetTasksWithPage(offset, size, userID, status)

	utils.SuccessWithPage(c, tasks, current, size, total)
}

// GetTaskByID godoc
// @Summary 获取单个任务详情
// @Description 根据任务ID获取任务详细信息
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Success 200 {object} utils.Response{data=define.Task}
// @Failure 404 {object} utils.Response
// @Router /algorithm/tasks/{id} [get]
func (h *AlgorithmHandler) GetTaskByID(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.Error(c, utils.VALIDATION_ERROR, "任务ID不能为空")
		return
	}

	task := h.system.GetTaskByID(taskID)
	if task == nil {
		utils.Error(c, utils.NOT_FOUND, "任务不存在")
		return
	}

	utils.Success(c, task)
}

// SubmitTask godoc
// @Summary 提交单个任务
// @Description 提交单个计算任务到调度系统
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body define.TaskBase true "任务信息"
// @Success 200 {object} utils.Response{data=define.Task}
// @Failure 400 {object} utils.Response
// @Router /algorithm/tasks [post]
func (h *AlgorithmHandler) SubmitTask(c *gin.Context) {
	var request define.TaskBase
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	// 提交单个任务（根据是否有优先级选择对应方法）
	var task *define.Task
	var err error

	if request.Priority != 0 {
		task, err = h.system.SubmitTaskWithPriority(
			request.UserID,
			request.DataSize,
			request.Type,
			request.Priority,
		)
	} else {
		task, err = h.system.SubmitTask(
			request.UserID,
			request.DataSize,
			request.Type,
		)
	}

	if err != nil {
		utils.Error(c, utils.ERROR, fmt.Sprintf("任务提交失败: %v", err))
		return
	}

	utils.SuccessWithMessage(c, task, "任务提交成功")
}

// DeleteTask godoc
// @Summary 删除任务
// @Description 根据任务ID删除任务
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "任务ID"
// @Success 200 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /algorithm/tasks/{id} [delete]
func (h *AlgorithmHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		utils.Error(c, utils.VALIDATION_ERROR, "任务ID不能为空")
		return
	}

	err := h.system.DeleteTask(taskID)
	if err != nil {
		utils.Error(c, utils.NOT_FOUND, "任务不存在")
		return
	}

	utils.SuccessWithMessage(c, nil, "任务删除成功")
}
