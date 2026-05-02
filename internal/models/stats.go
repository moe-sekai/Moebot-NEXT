package models

import "time"

// CommandStat records a single command invocation.
type CommandStat struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Command    string    `gorm:"not null;index" json:"command"`
	Platform   string    `gorm:"not null" json:"platform"`
	UserID     string    `json:"user_id"`
	GroupID    string    `json:"group_id"`
	Args       string    `json:"args"`
	ResponseMs int64     `json:"response_ms"` // response time in milliseconds
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}
