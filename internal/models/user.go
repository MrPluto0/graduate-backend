package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole 定义用户角色类型
type UserRole string

const (
	RoleAdmin UserRole = "admin" // 管理员
	RoleUser  UserRole = "user"  // 普通用户
)

// User 表示系统用户
// swagger:model
type User struct {
	ID        uint       `json:"id" gorm:"primarykey,autoIncrement"`            // 用户ID
	CreatedAt time.Time  `json:"created_at"`                                    // 创建时间
	UpdatedAt time.Time  `json:"updated_at"`                                    // 更新时间
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`             // 删除时间
	Username  string     `json:"username" gorm:"size:100;not null;uniqueIndex"` // 用户名
	Email     string     `json:"email" gorm:"size:100;not null;uniqueIndex"`    // 电子邮件
	Password  string     `json:"password" gorm:"size:100;not null"`             // 密码（JSON序列化时不返回）
	Role      UserRole   `json:"role" gorm:"size:20;default:user"`              // 用户角色
}

// BeforeCreate 在创建用户前的钩子函数
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 如果角色为空，设置为默认普通用户
	if u.Role == "" {
		u.Role = RoleUser
	}
	return nil
}
