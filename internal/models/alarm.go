package models

import (
	"time"
)

// AlarmStatus 告警状态枚举
type AlarmStatus string

const (
	AlarmStatusActive   AlarmStatus = "pending"  // 活跃状态
	AlarmStatusResolved AlarmStatus = "resolved" // 已解决
)

// AlarmEvent 事件类型枚举
type AlarmEvent string

const (
	AlarmEventHardware    AlarmEvent = "hardware"    // 硬件事件
	AlarmEventNetwork     AlarmEvent = "network"     // 网络事件
	AlarmEventSecurity    AlarmEvent = "security"    // 安全事件
	AlarmEventPerformance AlarmEvent = "performance" // 性能事件
	AlarmEventSystem      AlarmEvent = "system"      // 系统事件
)

// Alarm 告警数据模型
type Alarm struct {
	ID          uint        `json:"id" gorm:"primaryKey;autoIncrement" example:"1"`
	Name        string      `json:"name" gorm:"not null;size:255" validate:"required,min=1,max=255" example:"网络连接超时"`
	EventType   AlarmEvent  `json:"event_type" gorm:"not null;size:50" validate:"required" example:"network"`
	Status      AlarmStatus `json:"status" gorm:"not null;size:20;default:'active'" validate:"required" example:"active"`
	Description string      `json:"description" gorm:"type:text" validate:"max=1000" example:"设备与基站之间的网络连接超时，可能影响通信质量"`
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
}
