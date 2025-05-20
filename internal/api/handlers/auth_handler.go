package handlers

import (
	"go-backend/internal/service"
	"go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // 过期时间（秒）
}

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录并返回访问令牌和刷新令牌
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "登录信息"
// @Success 200 {object} utils.Response{data=TokenResponse}
// @Failure 401 {object} utils.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.VALIDATION_ERROR, err.Error())
		return
	}

	// 验证用户
	user, err := h.userService.ValidateUser(req.Username, req.Password)
	if err != nil {
		utils.Error(c, utils.UNAUTHORIZED, err.Error())
		return
	}

	// 生成访问令牌
	accessToken, err := utils.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		utils.Error(c, utils.ERROR, "生成访问令牌失败")
		return
	}

	// 生成刷新令牌
	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		utils.Error(c, utils.ERROR, "生成刷新令牌失败")
		return
	}

	utils.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    24 * 3600, // 24小时的秒数
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// RefreshToken godoc
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=TokenResponse}
// @Failure 401 {object} utils.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从请求头获取刷新令牌
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) <= 7 || authHeader[:7] != "Bearer " {
		utils.Error(c, utils.UNAUTHORIZED, "无效的令牌格式")
		return
	}
	refreshToken := authHeader[7:]

	// 解析刷新令牌
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		utils.Error(c, utils.UNAUTHORIZED, "无效的刷新令牌")
		return
	}

	// 验证令牌类型
	if claims.TokenType != "refresh" {
		utils.Error(c, utils.UNAUTHORIZED, "令牌类型错误")
		return
	}

	// 生成新的访问令牌
	accessToken, err := utils.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		utils.Error(c, utils.ERROR, "生成新的访问令牌失败")
		return
	}

	utils.Success(c, gin.H{
		"access_token": accessToken,
		"expires_in":   24 * 3600, // 24小时的秒数
	})
}

// GetCurrentUser godoc
// @Summary 获取当前登录用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// 从上下文中获取用户ID（由 AuthMiddleware 中间件设置）
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, utils.UNAUTHORIZED, "用户未登录")
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		utils.Error(c, utils.NOT_FOUND, "获取用户信息失败")
		return
	}

	utils.Success(c, user)
}
