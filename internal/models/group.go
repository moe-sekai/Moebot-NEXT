package models

import "time"

// Group represents a chat group's configuration.
type Group struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Platform  string    `gorm:"not null;uniqueIndex:idx_platform_group" json:"platform"`
	GroupID   string    `gorm:"not null;uniqueIndex:idx_platform_group" json:"group_id"`
	Name      string    `json:"name"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Config    string    `gorm:"type:text;default:'{}'" json:"config"` // JSON: feature toggles, permissions, etc.
	CreatedAt time.Time `json:"created_at"`
}
