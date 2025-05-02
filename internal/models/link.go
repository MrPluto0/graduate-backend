package models

import (
	"time"

	"gorm.io/gorm"
)

// LinkStatus 定义链路状态
type LinkStatus string

const (
	LinkStatusUp      LinkStatus = "up"      // 连通
	LinkStatusDown    LinkStatus = "down"    // 断开
	LinkStatusUnknown LinkStatus = "unknown" // 未知
)

// Link 表示网络拓扑中的链路
// swagger:model
type Link struct {
	ID          uint       `json:"id" gorm:"primarykey,autoIncrement"`    // 链路ID
	CreatedAt   time.Time  `json:"created_at"`                            // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`                            // 更新时间
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`     // 删除时间
	Name        string     `json:"name" gorm:"size:100;index"`            // 链路名称
	Status      LinkStatus `json:"status" gorm:"size:20;not null"`        // 链路状态
	SourceID    uint       `json:"source_id" gorm:"not null;index"`       // 源节点ID
	Source      Node       `json:"source" gorm:"foreignKey:SourceID"`     // 源节点
	TargetID    uint       `json:"target_id" gorm:"not null;index"`       // 目标节点ID
	Target      Node       `json:"target" gorm:"foreignKey:TargetID"`     // 目标节点
	Properties  string     `json:"properties,omitempty" gorm:"type:text"` // 链路属性(JSON格式)
	Description string     `json:"description" gorm:"size:500"`           // 链路描述
}

// BeforeCreate 创建前的钩子函数
func (l *Link) BeforeCreate(tx *gorm.DB) error {
	// 如果没有指定状态，默认为未知状态
	if l.Status == "" {
		l.Status = LinkStatusUnknown
	}
	return nil
}
