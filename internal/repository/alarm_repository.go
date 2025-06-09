package repository

import (
	"go-backend/internal/models"

	"gorm.io/gorm"
)

type AlarmRepository struct {
	db *gorm.DB
}

func NewAlarmRepository(db *gorm.DB) *AlarmRepository {
	return &AlarmRepository{db: db}
}

// Create 创建告警
func (r *AlarmRepository) Create(alarm *models.Alarm) error {
	return r.db.Create(alarm).Error
}

// GetByID 根据ID获取告警
func (r *AlarmRepository) GetByID(id uint) (*models.Alarm, error) {
	var alarm models.Alarm
	err := r.db.First(&alarm, id).Error
	if err != nil {
		return nil, err
	}
	return &alarm, nil
}

// Update 更新告警
func (r *AlarmRepository) Update(alarm *models.Alarm) error {
	return r.db.Save(alarm).Error
}

// Delete 删除告警
func (r *AlarmRepository) Delete(id uint) error {
	return r.db.Delete(&models.Alarm{}, id).Error
}

// List 获取告警列表
func (r *AlarmRepository) List(current, size int, filters map[string]interface{}) ([]models.Alarm, int64, error) {
	var alarms []models.Alarm
	var total int64

	query := r.db.Model(&models.Alarm{})

	// 应用过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "name":
				query = query.Where("name LIKE ?", "%"+value.(string)+"%")
			case "description":
				query = query.Where("description LIKE ?", "%"+value.(string)+"%")
			default:
				query = query.Where(key+" = ?", value)
			}
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (current - 1) * size
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&alarms).Error; err != nil {
		return nil, 0, err
	}

	return alarms, total, nil
}

// Count 统计告警数量
func (r *AlarmRepository) Count(filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&models.Alarm{})

	// 应用过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "status":
				query = query.Where("status = ?", value)
			case "event_type":
				query = query.Where("event_type = ?", value)
			}
		}
	}

	err := query.Count(&count).Error
	return count, err
}

// GetByStatus 根据状态获取告警列表
func (r *AlarmRepository) GetByStatus(status models.AlarmStatus) ([]models.Alarm, error) {
	var alarms []models.Alarm
	err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&alarms).Error
	return alarms, err
}

// GetByEventType 根据事件类型获取告警列表
func (r *AlarmRepository) GetByEventType(eventType models.AlarmEvent) ([]models.Alarm, error) {
	var alarms []models.Alarm
	err := r.db.Where("event_type = ?", eventType).Order("created_at DESC").Find(&alarms).Error
	return alarms, err
}

// GetActiveAlarms 获取所有活跃告警
func (r *AlarmRepository) GetActiveAlarms() ([]models.Alarm, error) {
	return r.GetByStatus(models.AlarmStatusActive)
}

// GetRecentAlarms 获取最近的告警（按时间倒序）
func (r *AlarmRepository) GetRecentAlarms(limit int) ([]models.Alarm, error) {
	var alarms []models.Alarm
	err := r.db.Order("created_at DESC").Limit(limit).Find(&alarms).Error
	return alarms, err
}
