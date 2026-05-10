package models

import "time"

// AdminUser 控制台本地账号。
//
// Username 与 Nickname 一经创建即不可更改：Username 用于登录认证，
// Nickname 会被注入到 Satori 渲染卡片底部 footer（"Moebot NEXT (deployed by X)"）。
type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Nickname     string    `gorm:"size:64;not null" json:"nickname"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AppMeta 通用键值表，用于持久化 JWT secret 等少量元数据。
type AppMeta struct {
	Key       string    `gorm:"primaryKey;size:64" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
