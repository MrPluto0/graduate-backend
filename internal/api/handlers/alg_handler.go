package handlers

import (
	"go-backend/internal/algorithm"
	"go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AlgorithmRequest 算法请求结构
type AlgorithmRequest struct {
	UserDataList []algorithm.UserData `json:"user_data_list"` // 各用户新产生的数据
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
// @Summary 启动算法
// @Description 接收用户新产生的数据，启动算法计算
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body AlgorithmRequest true "算法请求参数"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /algorithm/start [post]
func (h *AlgorithmHandler) StartAlgorithm(c *gin.Context) {
	var request AlgorithmRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	// 验证请求参数
	if len(request.UserDataList) == 0 {
		utils.Error(c, utils.VALIDATION_ERROR, "用户数据列表不能为空")
		return
	}

	// 启动算法
	if err := h.system.Start(request.UserDataList); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "算法启动成功，开始轮询执行")
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

// GetStateHistory godoc
// @Summary 获取状态历史
// @Description 获取算法执行的状态历史记录
// @Tags 算法管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=[]algorithm.State}
// @Router /algorithm/history [get]
func (h *AlgorithmHandler) GetStateHistory(c *gin.Context) {
	history := h.system.GetStateHistory()
	utils.Success(c, history)
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
