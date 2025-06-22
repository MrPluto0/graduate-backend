package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-backend/docs" // 导入生成的swagger文档
	"go-backend/internal/api"
	"go-backend/internal/config"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
)

// @title           网络设备管理系统 API
// @version         1.0
// @description     网络设备管理系统后端API文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description 请在此输入 'Bearer {token}' 格式的 JWT token

func main() {
	// 加载配置文件
	cfg := config.InitConfig()

	// 初始化 JWT 密钥
	utils.InitJWTSecret(cfg.JWT.Secret)

	// 初始化数据库连接
	database.InitDB("./data.db")

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	router := gin.Default()

	// 设置路由
	api.SetupRoutes(router)

	// 添加Swagger文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 设置静态文件目录
	router.Static("/static", "./static")

	// 展示Swagger文档
	log.Println("Swagger文档地址: http://localhost:" + cfg.Port + "/swagger/index.html")

	// 启动服务器
	log.Printf("启动服务器，监听端口 :%s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("无法启动服务器: %s\n", err)
	}

}
