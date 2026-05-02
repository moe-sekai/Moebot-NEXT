package models

import "time"

// ImageCache indexes cached rendered images on disk.
type ImageCache struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CacheKey   string    `gorm:"uniqueIndex;not null" json:"cache_key"` // e.g. "card_detail_123_v2"
	FilePath   string    `gorm:"not null" json:"file_path"`
	SizeBytes  int64     `json:"size_bytes"`
	CreatedAt  time.Time `json:"created_at"`
	LastAccess time.Time `json:"last_access"`
}
