package repository

import (
	"go-backend/internal/models"

	"gorm.io/gorm"
)

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

// Create 创建新链路
func (r *LinkRepository) Create(link *models.Link) error {
	return r.db.Create(link).Error
}

// GetByID 根据ID获取链路
func (r *LinkRepository) GetByID(id uint) (*models.Link, error) {
	var link models.Link
	err := r.db.Preload("Source").Preload("Target").First(&link, id).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// List 获取链路列表（不分页）
func (r *LinkRepository) List(filters map[string]interface{}) ([]models.Link, error) {
	var links []models.Link
	query := r.db.Model(&models.Link{}).Preload("Source").Preload("Target")

	// 应用过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	err := query.Find(&links).Error
	if err != nil {
		return nil, err
	}

	return links, nil
}

// ListWithPage 获取分页的链路列表
func (r *LinkRepository) ListWithPage(offset, limit int, filters map[string]interface{}) ([]models.Link, int64, error) {
	var links []models.Link
	var total int64

	query := r.db.Model(&models.Link{}).Preload("Source").Preload("Target")

	// 应用过滤条件
	for key, value := range filters {
		if key == "name" {
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
			continue
		}
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
	err = query.Offset(offset).Limit(limit).Find(&links).Error
	if err != nil {
		return nil, 0, err
	}

	return links, total, nil
}

// Update 更新链路信息
func (r *LinkRepository) Update(link *models.Link) error {
	return r.db.Save(link).Error
}

// Delete 删除链路
func (r *LinkRepository) Delete(id uint) error {
	return r.db.Delete(&models.Link{}, id).Error
}

// GetByNodes 根据源节点和目标节点获取链路
func (r *LinkRepository) GetByNodes(sourceID, targetID uint) (*models.Link, error) {
	var link models.Link
	err := r.db.Where("source_id = ? AND target_id = ?", sourceID, targetID).
		// Or("source_id = ? AND target_id = ?", targetID, sourceID). 不考虑双向链路
		Preload("Source").Preload("Target").
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// Count 统计链路数量
func (r *LinkRepository) Count(filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&models.Link{})

	// 应用过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	err := query.Count(&count).Error
	return count, err
}
