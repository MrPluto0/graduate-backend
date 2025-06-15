package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// NodeType 定义节点类型
type NodeType string

const (
	NodeTypeUser NodeType = "user_equipment" // 用户节点
	NodeTypeComm NodeType = "base_station"   // 基站节点
)

// Properties 节点属性
type Properties map[string]interface{}

// Value 实现 driver.Valuer 接口
func (p Properties) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口
func (p *Properties) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid scan source")
	}
	return json.Unmarshal(bytes, p)
}

// Node 表示网络拓扑中的节点
// swagger:model
type Node struct {
	ID          uint       `json:"id" gorm:"primarykey,autoIncrement"`        // 节点ID
	CreatedAt   time.Time  `json:"created_at"`                                // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`                                // 更新时间
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`         // 删除时间
	Name        string     `json:"name" gorm:"size:100;not null;uniqueIndex"` // 节点名称
	NodeType    NodeType   `json:"node_type" gorm:"size:50;index"`            // 节点类型
	X           int        `json:"x"`                                         // X坐标
	Y           int        `json:"y"`                                         // Y坐标
	Properties  Properties `json:"properties,omitempty" gorm:"type:json"`     // 节点属性(JSON格式)
	DeviceID    *uint      `json:"device_id,omitempty" gorm:"index"`          // 关联的设备ID
	Device      *Device    `json:"device,omitempty"`                          // 关联的设备
	Description string     `json:"description" gorm:"size:500"`               // 节点描述
}

// NodeStats 表示网络拓扑中的节点的性能指标
// swagger:model
type NodeStats struct {
	ID         uint       `json:"id" gorm:"primarykey,autoIncrement"` // 节点ID
	CreatedAt  time.Time  `json:"created_at"`                         // 创建时间
	UpdatedAt  time.Time  `json:"updated_at"`                         // 更新时间
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index"`  // 删除时间
	Timeslot   uint       `json:"timeslot" gorm:"index;not null"`     // 时间槽（时间戳）
	NodeID     uint       `json:"node_id" gorm:"index;not null"`      // 关联的节点ID
	Node       Node       `json:"node" gorm:"foreignKey:NodeID"`      // 关联的节点
	CPUUsage   float64    `json:"cpu_usage" gorm:"type:decimal(5,2)"` // CPU使用率（百分比）
	PacketsIn  int64      `json:"packets_in"`                         // 入站数据包数量
	PacketsOut int64      `json:"packets_out"`                        // 出站数据包数量
}
