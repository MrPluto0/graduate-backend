package api

import (
	"go-backend/internal/api/handlers"
	"go-backend/internal/api/middleware"
	"go-backend/internal/repository"
	"go-backend/internal/service"
	"go-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(router *gin.Engine) {
	// 获取数据库连接
	db := database.GetDB()

	// 初始化仓储层
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	linkRepo := repository.NewLinkRepository(db)

	// 初始化服务层
	userService := service.NewUserService(userRepo)
	deviceService := service.NewDeviceService(deviceRepo, nodeRepo, linkRepo)
	networkService := service.NewNetworkService(nodeRepo, linkRepo)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	deviceHandler := handlers.NewDeviceHandler(deviceService)
	networkHandler := handlers.NewNetworkHandler(networkService)
	overviewHandler := handlers.NewOverviewHandler(deviceService, networkService, userService)
	healthHandler := handlers.NewHealthHandler()
	algorithmHandler := handlers.NewAlgorithmHandler()

	// 公开路由组
	public := router.Group("/api/v1")
	{
		// 健康检查路由
		public.GET("/health", healthHandler.CheckHealth)

		// 认证相关路由（登录和刷新令牌无需认证）
		auth := public.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 算法管理路由
		algorithm := public.Group("/algorithm")
		{
			algorithm.POST("/start", algorithmHandler.StartAlgorithm)
			algorithm.POST("/stop", algorithmHandler.StopAlgorithm)
			algorithm.GET("/info", algorithmHandler.GetSystemInfo)
			algorithm.POST("/clear", algorithmHandler.ClearHistory)
		}
	}

	// 需要认证的路由组
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// 系统概览
		protected.GET("/overview", overviewHandler.GetOverview)

		// 认证相关路由
		auth := protected.Group("/auth")
		{
			auth.GET("/me", authHandler.GetCurrentUser)
		}

		// 用户管理路由
		users := protected.Group("/users")
		{
			users.GET("/:id", userHandler.GetUser)
		}

		// 管理员专用路由
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			// 管理员用户管理
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", userHandler.ListUsers)
				adminUsers.POST("", userHandler.CreateUser)
			}
		}

		// 设备管理路由
		devices := protected.Group("/devices")
		{
			devices.GET("", deviceHandler.ListDevices)
			devices.POST("", deviceHandler.CreateDevice)
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.PUT("/:id", deviceHandler.UpdateDevice)
			devices.DELETE("/:id", deviceHandler.DeleteDevice)
		}

		// 网络管理路由
		network := protected.Group("/network")
		{ // 节点管理
			nodes := network.Group("/nodes")
			{
				nodes.GET("", networkHandler.ListNodes)
				nodes.POST("", networkHandler.CreateNode)
				nodes.GET("/:id", networkHandler.GetNode)
				nodes.PUT("/:id", networkHandler.UpdateNode)
				nodes.DELETE("/:id", networkHandler.DeleteNode)
				nodes.PATCH("/batch-position", networkHandler.BatchUpdateNodesPosition) // 批量更新节点位置
			}

			// 链路管理
			links := network.Group("/links")
			{
				links.GET("", networkHandler.ListLinks)
				links.POST("", networkHandler.CreateLink)
				links.GET("/:id", networkHandler.GetLink)
				links.PUT("/:id", networkHandler.UpdateLink)
				links.DELETE("/:id", networkHandler.DeleteLink)
			}

			// 获取完整网络拓扑
			network.GET("/topology", networkHandler.GetTopology)
		}
	}
}
