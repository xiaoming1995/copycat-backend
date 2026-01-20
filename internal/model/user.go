package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:用户唯一ID(自增)" json:"id"`
	Email     string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null;comment:用户邮箱(用于登录)" json:"email"`
	Password  string    `gorm:"column:password;type:varchar(255);not null;comment:密码哈希" json:"-"` // json:"-" 防止密码泄露
	Nickname  string    `gorm:"column:nickname;type:varchar(100);comment:用户昵称" json:"nickname"`
	Avatar    string    `gorm:"column:avatar;type:varchar(500);comment:头像URL" json:"avatar"`
	Bio       string    `gorm:"column:bio;type:text;comment:个人简介" json:"bio"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
