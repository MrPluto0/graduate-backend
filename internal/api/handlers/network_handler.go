package handlers

import (
	"go-backend/internal/models"
	"go-backend/internal/service"
	"go-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BatchUpdateNodesPositionRequest 批量更新节点位置请求
type BatchUpdateNodesPositionRequest struct {
	Nodes []models.Node `json:"nodes" binding:"required,dive"` // 节点列表，只需要ID、X、Y字段
}

type NetworkHandler struct {
	networkService *service.NetworkService
}

func NewNetworkHandler(networkService *service.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
	}
}

// ListNodes godoc
// @Summary 获取所有节点
// @Description 获取网络拓扑中的所有节点列表
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param current query int false "页码(默认1)"
// @Param size query int false "每页数量(默认10)"
// @Param search query string false "节点名称搜索关键词"
// @Param node_type query string false "节点类型筛选"
// @Success 200 {object} utils.Response{data=utils.PageResult{records=[]models.Node}}
// @Router /network/nodes [get]
func (h *NetworkHandler) ListNodes(c *gin.Context) {
	// 获取分页参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	offset := (current - 1) * size

	// 构建过滤条件
	filters := make(map[string]interface{})
	if nodeType := c.Query("node_type"); nodeType != "" {
		filters["node_type"] = nodeType
	}
	if search := c.Query("search"); search != "" {
		filters["name"] = search
	}

	nodes, total, err := h.networkService.ListNodesWithPage(offset, size, filters)
	if err != nil {
		utils.Error(c, utils.ERROR, "获取节点列表失败")
		return
	}

	utils.SuccessWithPage(c, nodes, current, size, total)
}

// GetNode godoc
// @Summary 获取节点详情
// @Description 根据ID获取节点详细信息
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "节点ID"
// @Success 200 {object} utils.Response{data=models.Node}
// @Failure 404 {object} utils.Response
// @Router /network/nodes/{id} [get]
func (h *NetworkHandler) GetNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的节点ID")
		return
	}

	node, err := h.networkService.GetNode(uint(id))
	if err != nil {
		utils.Error(c, utils.NOT_FOUND, "节点不存在")
		return
	}

	utils.Success(c, node)
}

// CreateNode godoc
// @Summary 创建节点
// @Description 在网络拓扑中创建新节点
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param node body models.Node true "节点信息"
// @Success 201 {object} utils.Response{data=models.Node}
// @Failure 400 {object} utils.Response
// @Router /network/nodes [post]
func (h *NetworkHandler) CreateNode(c *gin.Context) {
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	if err := h.networkService.CreateNode(&node); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, node, "节点创建成功")
}

// UpdateNode godoc
// @Summary 更新节点
// @Description 更新网络拓扑中的节点信息
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "节点ID"
// @Param node body models.Node true "节点信息"
// @Success 200 {object} utils.Response{data=models.Node}
// @Failure 400,404 {object} utils.Response
// @Router /network/nodes/{id} [put]
func (h *NetworkHandler) UpdateNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的节点ID")
		return
	}

	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}
	node.ID = uint(id)

	if err := h.networkService.UpdateNode(&node); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, node, "节点更新成功")
}

// DeleteNode godoc
// @Summary 删除节点
// @Description 从网络拓扑中删除节点
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "节点ID"
// @Success 200 {object} utils.Response
// @Failure 400,404 {object} utils.Response
// @Router /network/nodes/{id} [delete]
func (h *NetworkHandler) DeleteNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的节点ID")
		return
	}

	if err := h.networkService.DeleteNode(uint(id)); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "节点删除成功")
}

// ListLinks godoc
// @Summary 获取所有链路
// @Description 获取网络拓扑中的所有链路列表
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param current query int false "页码(默认1)"
// @Param size query int false "每页数量(默认10)"
// @Param search query string false "链路名称搜索关键词"
// @Success 200 {object} utils.Response{data=utils.PageResult{records=[]models.Link}}
// @Router /network/links [get]
func (h *NetworkHandler) ListLinks(c *gin.Context) {
	// 获取分页参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	offset := (current - 1) * size

	// 构建过滤条件
	filters := make(map[string]interface{})
	if search := c.Query("search"); search != "" {
		filters["name"] = search
	}

	links, total, err := h.networkService.ListLinks(offset, size, filters)
	if err != nil {
		utils.Error(c, utils.ERROR, "获取链路列表失败")
		return
	}

	utils.SuccessWithPage(c, links, current, size, total)
}

// GetLink godoc
// @Summary 获取链路详情
// @Description 根据ID获取链路详细信息
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "链路ID"
// @Success 200 {object} utils.Response{data=models.Link}
// @Failure 404 {object} utils.Response
// @Router /network/links/{id} [get]
func (h *NetworkHandler) GetLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的链路ID")
		return
	}

	link, err := h.networkService.GetLink(uint(id))
	if err != nil {
		utils.Error(c, utils.NOT_FOUND, "链路不存在")
		return
	}

	utils.Success(c, link)
}

// CreateLink godoc
// @Summary 创建链路
// @Description 在网络拓扑中创建新链路
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param link body models.Link true "链路信息"
// @Success 201 {object} utils.Response{data=models.Link}
// @Failure 400 {object} utils.Response
// @Router /network/links [post]
func (h *NetworkHandler) CreateLink(c *gin.Context) {
	var link models.Link
	if err := c.ShouldBindJSON(&link); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	if err := h.networkService.CreateLink(&link); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, link, "链路创建成功")
}

// UpdateLink godoc
// @Summary 更新链路
// @Description 更新网络拓扑中的链路信息
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "链路ID"
// @Param link body models.Link true "链路信息"
// @Success 200 {object} utils.Response{data=models.Link}
// @Failure 400,404 {object} utils.Response
// @Router /network/links/{id} [put]
func (h *NetworkHandler) UpdateLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的链路ID")
		return
	}

	var link models.Link
	if err := c.ShouldBindJSON(&link); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}
	link.ID = uint(id)

	if err := h.networkService.UpdateLink(&link); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, link, "链路更新成功")
}

// DeleteLink godoc
// @Summary 删除链路
// @Description 从网络拓扑中删除链路
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "链路ID"
// @Success 200 {object} utils.Response
// @Failure 400,404 {object} utils.Response
// @Router /network/links/{id} [delete]
func (h *NetworkHandler) DeleteLink(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的链路ID")
		return
	}

	if err := h.networkService.DeleteLink(uint(id)); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "链路删除成功")
}

// GetTopology godoc
// @Summary 获取网络拓扑
// @Description 获取完整的网络拓扑数据，包括所有节点和链路
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=service.TopologyData}
// @Router /network/topology [get]
func (h *NetworkHandler) GetTopology(c *gin.Context) {
	topology, err := h.networkService.GetTopology()
	if err != nil {
		utils.Error(c, utils.ERROR, "获取网络拓扑失败")
		return
	}

	utils.Success(c, topology)
}

// BatchUpdateNodesPosition godoc
// @Summary 批量更新节点位置
// @Description 批量更新网络拓扑中的节点位置坐标
// @Tags 网络管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body BatchUpdateNodesPositionRequest true "批量更新位置请求"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /network/nodes/batch-position [patch]
func (h *NetworkHandler) BatchUpdateNodesPosition(c *gin.Context) {
	var request BatchUpdateNodesPositionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	if len(request.Nodes) == 0 {
		utils.Error(c, utils.VALIDATION_ERROR, "节点列表不能为空")
		return
	}

	if err := h.networkService.BatchUpdateNodesPosition(request.Nodes); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, nil, "批量更新节点位置成功")
}
