package handlers

import (
	"go-backend/internal/algorithm"
	"go-backend/internal/algorithm/define"
	"go-backend/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskSubmitRequest struct {
	UserID   uint    `json:"user_id" binding:"required"`
	DataSize float64 `json:"data_size" binding:"required,gt=0"`
	TaskType string  `json:"task_type"`
}

type TaskResponse struct {
	TaskID     string    `json:"task_id"`
	UserID     uint      `json:"user_id"`
	DataSize   float64   `json:"data_size"`
	TaskType   string    `json:"task_type"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time"`
	TotalDelay float64   `json:"total_delay,omitempty"`
}

type AlgorithmHandler struct {
	system *algorithm.System
}

func NewAlgorithmHandler() *AlgorithmHandler {
	return &AlgorithmHandler{
		system: algorithm.GetSystemInstance(),
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
	var requests []TaskSubmitRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	if len(requests) == 0 {
		utils.Error(c, utils.VALIDATION_ERROR, "任务列表不能为空")
		return
	}

	tasks := make([]*define.Task, 0)
	for _, req := range requests {
		task, err := h.system.SubmitTask(define.TaskBase{
			UserID:   req.UserID,
			DataSize: req.DataSize,
			TaskType: req.TaskType,
		})
		if err != nil {
			utils.Error(c, utils.ERROR, err.Error())
			return
		}
		tasks = append(tasks, task)
	}

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
// @Success 200 {object} utils.Response{data=map[string]interface{}}
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
