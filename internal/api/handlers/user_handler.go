package handlers

import (
	"go-backend/internal/models"
	"go-backend/internal/service"
	"go-backend/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser godoc
// @Summary 创建新用户
// @Description 创建新用户账号
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body models.User true "用户信息"
// @Success 201 {object} utils.Response{data=models.User}
// @Failure 400 {object} utils.Response
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	if err := h.userService.CreateUser(&user); err != nil {
		utils.Error(c, utils.ERROR, err.Error())
		return
	}

	utils.SuccessWithMessage(c, user, "用户创建成功")
}

// GetUser godoc
// @Summary 获取用户信息
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "用户ID"
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 404 {object} utils.Response
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	// 将字符串ID转换为uint
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, "无效的用户ID")
		return
	}

	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		utils.Error(c, utils.NOT_FOUND, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// ListUsers godoc
// @Summary 获取用户列表
// @Description 获取用户列表，仅管理员可访问
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param current query int false "当前页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param role query string false "用户角色筛选" Enums(admin,user)
// @Param search query string false "搜索关键字（用户名或邮箱）"
// @Success 200 {object} utils.Response{data=[]models.User}
// @Failure 403 {object} utils.Response
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 获取分页参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	// 构建过滤条件
	filters := make(map[string]interface{})
	if role := c.Query("role"); role != "" {
		filters["role"] = role
	}

	// 搜索关键字可以同时匹配用户名和邮箱
	if search := c.Query("search"); search != "" {
		filters["username"] = search
		// 这里简化处理，实际可以用OR条件同时搜索用户名和邮箱
	}

	// 获取用户列表
	users, total, err := h.userService.ListUsers(current, size, filters)
	if err != nil {
		utils.Error(c, utils.ERROR, "获取用户列表失败")
		return
	}

	// 清除密码字段
	for i := range users {
		users[i].Password = ""
	}

	utils.SuccessWithPage(c, users, current, size, total)
}
