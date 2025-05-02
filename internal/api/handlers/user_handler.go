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
		utils.Error(c, utils.VALIDATION_ERROR, "无效的用户数据")
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
