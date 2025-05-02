package repository

import (
	"go-backend/internal/models"

	"gorm.io/gorm"
)

type DeviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create 创建新设备
func (r *DeviceRepository) Create(device *models.Device) error {
	return r.db.Create(device).Error
}

// GetByID 根据ID获取设备
func (r *DeviceRepository) GetByID(id uint) (*models.Device, error) {
	var device models.Device
	err := r.db.First(&device, id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// List 获取设备列表，支持分页和过滤
func (r *DeviceRepository) List(offset, limit int, filters map[string]interface{}) ([]models.Device, int64, error) {
	var devices []models.Device
	var total int64

	query := r.db.Model(&models.Device{})

	// 应用过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.Offset(offset).Limit(limit).Find(&devices).Error
	if err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

// Update 更新设备信息
func (r *DeviceRepository) Update(device *models.Device) error {
	return r.db.Save(device).Error
}

// Delete 删除设备
func (r *DeviceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Device{}, id).Error
}

// GetByMAC 根据MAC地址获取设备
func (r *DeviceRepository) GetByMAC(mac string) (*models.Device, error) {
	var device models.Device
	err := r.db.Where("mac = ?", mac).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}
