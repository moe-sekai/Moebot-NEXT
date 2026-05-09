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

// UserDefaultRegion stores a platform user's preferred default game region.
// Used when a moesekai command is invoked without a region prefix and the user
// has explicitly set a default via /pjsk服务器.
type UserDefaultRegion struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Platform     string    `gorm:"not null;uniqueIndex:idx_user_default_region" json:"platform"`
	PlatformID   string    `gorm:"not null;uniqueIndex:idx_user_default_region" json:"platform_id"`
	ServerRegion string    `gorm:"not null;default:'jp'" json:"server_region"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SuiteSetting stores per-region Suite data privacy and source preference.
type SuiteSetting struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Platform     string    `gorm:"not null;uniqueIndex:idx_suite_setting_user_region" json:"platform"`
	PlatformID   string    `gorm:"not null;uniqueIndex:idx_suite_setting_user_region" json:"platform_id"`
	ServerRegion string    `gorm:"not null;default:'jp';uniqueIndex:idx_suite_setting_user_region" json:"server_region"`
	Mode         string    `gorm:"not null;default:'latest'" json:"mode"`
	Hidden       bool      `json:"hidden"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
