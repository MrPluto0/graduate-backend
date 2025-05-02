package models

import (
	"time"

	"gorm.io/gorm"
)

// DeviceType 定义设备类型
type DeviceType string

const (
	UserDevice    DeviceType = "user_device"    // 用户设备
	ComputeDevice DeviceType = "compute_device" // 网络设备(路由器、交换机等)
	NetworkDevice DeviceType = "network_device" // 核心网网元
)

// DeviceStatus 定义设备状态
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"  // 在线
	DeviceStatusOffline DeviceStatus = "offline" // 离线
	DeviceStatusFault   DeviceStatus = "fault"   // 故障
)

// Device 表示网络设备
// swagger:model
type Device struct {
	ID          uint         `json:"id" gorm:"primarykey,autoIncrement"`        // 设备ID
	CreatedAt   time.Time    `json:"created_at"`                                // 创建时间
	UpdatedAt   time.Time    `json:"updated_at"`                                // 更新时间
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" gorm:"index"`         // 删除时间
	Name        string       `json:"name" gorm:"size:100;not null;index"`       // 设备名称
	DeviceType  DeviceType   `json:"device_type" gorm:"size:50;not null;index"` // 设备类型
	MAC         string       `json:"mac" gorm:"size:50;unique;index"`           // MAC地址
	IP          string       `json:"ip" gorm:"size:50;index"`                   // IP地址
	Status      DeviceStatus `json:"status" gorm:"size:20;not null"`            // 设备状态
	Vendor      string       `json:"vendor" gorm:"size:100"`                    // 设备厂商
	Version     string       `json:"version" gorm:"size:50"`                    // 设备版本
	Location    string       `json:"location" gorm:"size:200"`                  // 设备位置
	Description string       `json:"description" gorm:"size:500"`               // 设备描述
	Config      string       `json:"config_json,omitempty" gorm:"type:text"`    // 设备配置(JSON格式)
	MetaData    string       `json:"meta_data,omitempty" gorm:"type:text"`      // 元数据(JSON格式)
	UserID      *uint        `json:"user_id,omitempty" gorm:"index"`            // 所属用户ID
}

// BeforeCreate 创建前的钩子函数
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	// 如果没有指定状态，默认为离线状态
	if d.Status == "" {
		d.Status = DeviceStatusOffline
	}
	return nil
}
