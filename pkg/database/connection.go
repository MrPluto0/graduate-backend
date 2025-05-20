package database

import (
	"encoding/json"
	"go-backend/internal/models"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(dbPath string) {
	var err error

	// 配置日志选项
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

	// 导入初始测试数据
	importTestData()
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

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}

// loadJSONFile 从JSON文件加载测试数据
func loadJSONFile(filename string, v interface{}) error {
	dataPath := filepath.Join("data", filename)
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// importTestData 导入测试数据
func importTestData() {
	// 检查是否已存在测试数据
	var count int64
	DB.Model(&models.User{}).Where("username <> ?", "admin").Count(&count)
	if count > 0 {
		return // 已存在测试数据，不再重复导入
	}

	// 创建一个事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 导入设备数据
	var deviceData struct {
		Devices []models.Device `json:"devices"`
	}
	if err := loadJSONFile("devices.json", &deviceData); err != nil {
		log.Printf("加载设备数据失败: %v", err)
		tx.Rollback()
		return
	}
	if err := tx.Create(&deviceData.Devices).Error; err != nil {
		log.Printf("导入设备数据失败: %v", err)
		tx.Rollback()
		return
	}

	// 导入节点数据
	var nodeData struct {
		Nodes []models.Node `json:"nodes"`
	}
	if err := loadJSONFile("nodes.json", &nodeData); err != nil {
		log.Printf("加载节点数据失败: %v", err)
		tx.Rollback()
		return
	}

	for _, node := range nodeData.Nodes {
		if err := tx.Create(&node).Error; err != nil {
			log.Printf("导入节点数据失败: %v", err)
			tx.Rollback()
			return
		}
	}

	// 导入链路数据
	var linkData struct {
		Links []models.Link `json:"links"`
	}
	if err := loadJSONFile("links.json", &linkData); err != nil {
		log.Printf("加载链路数据失败: %v", err)
		tx.Rollback()
		return
	}

	if err := tx.Create(&linkData.Links).Error; err != nil {
		log.Printf("导入链路数据失败: %v", err)
		tx.Rollback()
		return
	}

	// 导入用户数据，需要特殊处理密码加密
	var userData struct {
		Users []models.User `json:"users"`
	}
	if err := loadJSONFile("users.json", &userData); err != nil {
		log.Printf("加载用户数据失败: %v", err)
		tx.Rollback()
		return
	}

	// 创建用户并加密密码
	for _, u := range userData.Users {
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		user := models.User{
			Username: u.Username,
			Email:    u.Email,
			Password: string(passwordHash),
			Role:     u.Role,
		}
		if err := tx.Create(&user).Error; err != nil {
			log.Printf("导入用户数据失败: %v", err)
			tx.Rollback()
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("提交测试数据事务失败: %v", err)
		return
	}

	log.Println("成功导入初始测试数据")
}
