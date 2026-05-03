package models

import "time"

// User represents one region-specific bound game account for a platform user.
type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Platform     string    `gorm:"not null;uniqueIndex:idx_platform_user_region" json:"platform"` // onebot, discord, telegram
	PlatformID   string    `gorm:"not null;uniqueIndex:idx_platform_user_region" json:"platform_id"`
	ServerRegion string    `gorm:"not null;default:'jp';uniqueIndex:idx_platform_user_region" json:"server_region"`
	GameID       string    `json:"game_id"` // PJSK game user ID; keep string to avoid precision loss
	Nickname     string    `json:"nickname"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
