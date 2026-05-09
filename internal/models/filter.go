package models

import "time"

// FilterGateway holds the OneBot reverse-WS gateway settings.
// Only one row is expected; we identify it by the fixed ID = 1.
//
// Default ID-rule fallbacks are stored on the seeded built-in `default` FilterTemplate
// (see FilterTemplate) — they used to live here as DefaultUserIDRules / DefaultGroupIDRules
// but were migrated out as part of the template refactor.
type FilterGateway struct {
	ID uint `gorm:"primarykey" json:"id"`
	// Enabled toggles the gateway on/off.
	Enabled bool   `gorm:"default:true" json:"enabled"`
	Host    string `gorm:"default:'0.0.0.0'" json:"host"`
	Port    int    `gorm:"default:3939" json:"port"`
	Suffix  string `gorm:"default:'/ws'" json:"suffix"`
	BotID   string `gorm:"default:'10000'" json:"bot_id"`
	// AccessToken, if non-empty, is required from upstream OneBot clients
	// connecting to the gateway. Accepted via header `Authorization: Bearer <t>`
	// / `Authorization: Token <t>` or query string `?access_token=<t>`.
	AccessToken string    `gorm:"default:''" json:"access_token"`
	UserAgent   string    `gorm:"default:'Moebot'" json:"user_agent"`
	BufferSize  int       `gorm:"default:4096" json:"buffer_size"`
	SleepTime   float32   `gorm:"default:5" json:"sleep_time"`
	Debug       bool      `json:"debug"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FilterTemplate is a reusable bundle of filter rules that one or more
// FilterApp records can reference via FilterApp.TemplateID. The seeded
// built-in template named "default" doubles as the global ID-rule fallback
// (used when an app's own user_id/group_id rule has mode=="default").
type FilterTemplate struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"not null;uniqueIndex" json:"name"`
	Description string `json:"description"`
	Builtin     bool   `gorm:"default:false" json:"builtin"`

	UserIDRules         string `gorm:"type:text;default:'{}'" json:"user_id_rules"`
	GroupIDRules        string `gorm:"type:text;default:'{}'" json:"group_id_rules"`
	MessageRules        string `gorm:"type:text;default:'{}'" json:"message_rules"`
	PrivateMessageRules string `gorm:"type:text;default:'{}'" json:"private_message_rules"`
	GroupMessageRules   string `gorm:"type:text;default:'{}'" json:"group_message_rules"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FilterApp represents a downstream OneBot bot application that the gateway forwards messages to.
// All filter rule fields are JSON-encoded so the schema stays simple yet flexible.
//
// When TemplateID is non-nil, the app's own rule fields are ignored at compile
// time and the referenced FilterTemplate's rules are used instead.
type FilterApp struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"not null;uniqueIndex" json:"name"`
	URI         string `gorm:"not null" json:"uri"`
	AccessToken string `json:"access_token"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	Builtin     bool   `gorm:"default:false" json:"builtin"` // true for the Moebot built-in plugin row
	// Internal=true 表示该 App 不开 WS 客户端，仅作为规则容器供插件查询
	// （插件级独立模板）。URI/AccessToken 字段对内部 App 无意义。
	Internal  bool `gorm:"default:false" json:"internal"`
	SortOrder int  `gorm:"default:0" json:"sort_order"`

	// TemplateID points to a FilterTemplate; when set, the rule fields below are ignored.
	TemplateID *uint `gorm:"index" json:"template_id,omitempty"`

	// JSON: {"mode":"default|on|off|whitelist|blacklist","ids":[...]}
	UserIDRules  string `gorm:"type:text;default:'{}'" json:"user_id_rules"`
	GroupIDRules string `gorm:"type:text;default:'{}'" json:"group_id_rules"`

	// JSON: {"mode":"default|on|off|whitelist|blacklist","filters":[...],"prefix":[...],"prefix_replace":""}
	MessageRules        string `gorm:"type:text;default:'{}'" json:"message_rules"`
	PrivateMessageRules string `gorm:"type:text;default:'{}'" json:"private_message_rules"`
	GroupMessageRules   string `gorm:"type:text;default:'{}'" json:"group_message_rules"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
