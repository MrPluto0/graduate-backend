package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(dbPath string) {
	var err error

	// 配置日志选项 - Silent 模式下不显示任何日志
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// 连接SQLite数据库
	DB, err = gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移数据库表结构
	migrateDB()

	// 初始化默认管理员账户
	createDefaultAdmin()
}

// 自动迁移数据库表结构
func migrateDB() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.Node{},
		&models.Link{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
}

// createDefaultAdmin 创建默认管理员账户
func createDefaultAdmin() {
	// 检查是否已存在管理员账户
	var count int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)

	// 如果没有管理员账户，则创建默认管理员
	if count == 0 {
		// 默认密码加密
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("生成密码哈希失败: %v", err)
		}

		// 创建管理员用户
		admin := models.User{
			Username: "admin",
			Password: string(passwordHash),
			Email:    "admin@example.com",
			Role:     models.RoleAdmin,
		}

		result := DB.Create(&admin)
		if result.Error != nil {
			log.Fatalf("创建默认管理员账户失败: %v", result.Error)
		} else {
			log.Println("已成功创建默认管理员账户 (用户名: admin, 密码: admin123)")
		}
	}
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}
