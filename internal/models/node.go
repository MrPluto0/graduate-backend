package models

import (
	"time"
)

// NodeType 定义节点类型
type NodeType string

const (
	NodeTypeDevice NodeType = "device" // 设备节点
	NodeTypeRouter NodeType = "router" // 路由器节点
	NodeTypeSwitch NodeType = "switch" // 交换机节点
	NodeTypeHost   NodeType = "host"   // 主机节点
)

// Node 表示网络拓扑中的节点
// swagger:model
type Node struct {
	ID          uint       `json:"id" gorm:"primarykey,autoIncrement"`    // 节点ID
	CreatedAt   time.Time  `json:"created_at"`                            // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`                            // 更新时间
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`     // 删除时间
	Name        string     `json:"name" gorm:"size:100;not null;index"`   // 节点名称
	NodeType    NodeType   `json:"node_type" gorm:"size:50;index"`        // 节点类型
	X           int        `json:"x"`                                     // X坐标
	Y           int        `json:"y"`                                     // Y坐标
	Properties  string     `json:"properties,omitempty" gorm:"type:text"` // 节点属性(JSON格式)
	DeviceID    *uint      `json:"device_id,omitempty" gorm:"index"`      // 关联的设备ID
	Device      *Device    `json:"device,omitempty"`                      // 关联的设备
	Description string     `json:"description" gorm:"size:500"`           // 节点描述
}
