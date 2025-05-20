package repository

import (
	"go-backend/internal/models"

	"gorm.io/gorm"
)

type NodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Create 创建新节点
func (r *NodeRepository) Create(node *models.Node) error {
	return r.db.Create(node).Error
}

// GetByID 根据ID获取节点
func (r *NodeRepository) GetByID(id uint) (*models.Node, error) {
	var node models.Node
	err := r.db.Preload("Device").First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetByDeviceID 根据设备ID获取节点
func (r *NodeRepository) GetByDeviceID(deviceID uint) (*models.Node, error) {
	var node models.Node
	err := r.db.Where("device_id = ?", deviceID).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// ListWithPage 获取分页的节点列表
func (r *NodeRepository) List(filters map[string]interface{}) ([]models.Node, error) {
	var nodes []models.Node

	query := r.db.Model(&models.Node{}).Preload("Device")

	// 应用过滤条件
	for key, value := range filters {
		if key == "name" && value != "" {
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
			continue
		}
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	// 获取分页数据
	err := query.Find(&nodes).Error
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// ListWithPage 获取分页的节点列表
func (r *NodeRepository) ListWithPage(offset, limit int, filters map[string]interface{}) ([]models.Node, int64, error) {
	var nodes []models.Node
	var total int64

	query := r.db.Model(&models.Node{}).Preload("Device")

	// 应用过滤条件
	for key, value := range filters {
		if key == "name" && value != "" {
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
	err = query.Offset(offset).Limit(limit).Find(&nodes).Error
	if err != nil {
		return nil, 0, err
	}

	return nodes, total, nil
}

// Update 更新节点信息
func (r *NodeRepository) Update(node *models.Node) error {
	return r.db.Save(node).Error
}

// Delete 删除节点
func (r *NodeRepository) Delete(id uint) error {
	return r.db.Delete(&models.Node{}, id).Error
}

// Count 统计节点数量
func (r *NodeRepository) Count(filters map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.Model(&models.Node{})

	// 应用过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	err := query.Count(&count).Error
	return count, err
}
