package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"size:64;uniqueIndex;not null"` // 用户名
	Password  string    `json:"-" gorm:"size:128;not null"`                   // 密码（加密后）
	Nickname  string    `json:"nickname" gorm:"size:128"`                     // 昵称
	Email     string    `json:"email" gorm:"size:128"`                        // 邮箱
	Role      string    `json:"role" gorm:"size:32;default:'user'"`           // 角色：admin, user
	Status    string    `json:"status" gorm:"size:32;default:'active'"`       // 状态：active, disabled
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`             // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`             // 更新时间
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// UserRole 用户角色常量
const (
	RoleAdmin = "admin" // 管理员
	RoleUser  = "user"  // 普通用户
)

// UserStatus 用户状态常量
const (
	StatusActive   = "active"   // 激活
	StatusDisabled = "disabled" // 禁用
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	User      *User  `json:"user"`
	ExpiresIn int64  `json:"expires_in"` // 过期时间（秒）
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string  `json:"nickname"`
	Email    string  `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty,min=6,max=32"`
	Role     string  `json:"role"`
	Status   string  `json:"status"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

// InitSystemRequest 系统初始化请求
type InitSystemRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname"`
	Email    string `json:"email" binding:"omitempty,email"`
}
