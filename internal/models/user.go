package models

import "time"

// User represents a bound user account.
type User struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Platform   string    `gorm:"not null;uniqueIndex:idx_platform_user" json:"platform"` // onebot, discord, telegram
	PlatformID string    `gorm:"not null;uniqueIndex:idx_platform_user" json:"platform_id"`
	GameID     string    `json:"game_id"` // PJSK game user ID
	Nickname   string    `json:"nickname"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
