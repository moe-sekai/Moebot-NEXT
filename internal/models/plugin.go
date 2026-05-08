package models

import "time"

// PluginState 持久化每个插件的启用状态。
// 启用状态由 WebUI / 启动时种子化共同维护，记录变更时间便于审计。
type PluginState struct {
	Name      string    `gorm:"primaryKey;size:128" json:"name"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}
