package middleware

import (
	"go-backend/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 是一个基于 JWT 的认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, utils.UNAUTHORIZED, "未提供认证信息")
			c.Abort()
			return
		}

		// 检查 Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Error(c, utils.UNAUTHORIZED, "认证格式错误")
			c.Abort()
			return
		}

		// 解析 token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.Error(c, utils.UNAUTHORIZED, "无效的token")
			c.Abort()
			return
		}

		// 验证令牌类型
		if claims.TokenType != "access" {
			utils.Error(c, utils.UNAUTHORIZED, "令牌类型错误")
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware 管理员权限检查中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先检查是否已经通过了认证
		role, exists := c.Get("role")
		if !exists {
			utils.Error(c, utils.UNAUTHORIZED, "用户未登录")
			c.Abort()
			return
		}

		// 检查是否为管理员
		if role.(string) != string("admin") {
			utils.Error(c, utils.FORBIDDEN, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
