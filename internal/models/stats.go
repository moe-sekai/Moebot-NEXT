package models

import "time"

// CommandStat records a single command invocation.
type CommandStat struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Command    string    `gorm:"not null;index" json:"command"`
	Platform   string    `gorm:"not null" json:"platform"`
	ClientID   string    `gorm:"index;default:''" json:"client_id"`
	UserID     string    `json:"user_id"`
	GroupID    string    `json:"group_id"`
	Region     string    `gorm:"index" json:"region"`
	Args       string    `json:"args"`
	ResponseMs int64     `json:"response_ms"` // response time in milliseconds
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}
