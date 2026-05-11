package filter

import (
	"errors"

	"moebot-next/internal/database"
	"moebot-next/internal/models"

	"gorm.io/gorm"
)

// InternalAppName 由插件名生成 filter app 名字，统一使用 `plugin:<name>` 命名空间，
// 避免与用户手动新建的下游 app 冲突。
func InternalAppName(pluginName string) string {
	return "plugin:" + pluginName
}

// BuiltinTransportName 是 main.go 在启动时种入的"内置传输闸门"FilterApp 名字。
//
// 该 App 是网关把上游 OneBot 事件转发给 Bot 主进程的唯一通道（ws 下游）。
// 它的规则被运行时强制锁定为全开（ModeOn），所有过滤都交给各个 plugin:<name>
// internal app 去做，避免与插件级规则形成令人困惑的串联语义。
//
// 控制台 Filter 页面对该名字的 App 隐藏规则/模板编辑器（仅允许改 URI /
// AccessToken / Enabled 等传输参数）。
const BuiltinTransportName = "moebot-builtin"

// IsBuiltinTransport 判定 App 名是否为内置传输闸门。
func IsBuiltinTransport(name string) bool { return name == BuiltinTransportName }

// SeedInternalApp 确保某个插件对应的 internal FilterApp 存在；缺失时创建一个
// 默认引用 `default` 模板、Internal=true、Builtin=true 的行，从而在控制台
// Filter 页面可见、可分配模板，但不会开 ws 客户端。
//
// 已存在时仅修补 Internal/Builtin 标志，不覆盖用户的规则配置。
func SeedInternalApp(db *database.DB, pluginName, displayName string) error {
	if db == nil {
		return errors.New("filter: nil database")
	}
	name := InternalAppName(pluginName)
	app, err := db.GetFilterAppByName(name)
	if err == nil && app != nil {
		dirty := false
		if !app.Internal {
			app.Internal = true
			dirty = true
		}
		if !app.Builtin {
			app.Builtin = true
			dirty = true
		}
		if dirty {
			return db.UpdateFilterApp(app)
		}
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	defaultTpl, err := db.GetDefaultFilterTemplate()
	if err != nil {
		return err
	}
	tplID := defaultTpl.ID
	row := &models.FilterApp{
		Name:                name,
		URI:                 "internal://" + pluginName,
		Enabled:             true,
		Builtin:             true,
		Internal:            true,
		SortOrder:           100, // 排在用户自定义 app 之后
		TemplateID:          &tplID,
		UserIDRules:         EncodeIDRule(IDRule{Mode: ModeDefault}),
		GroupIDRules:        EncodeIDRule(IDRule{Mode: ModeDefault}),
		MessageRules:        EncodeMessageRule(MessageRule{Mode: ModeOn}),
		PrivateMessageRules: EncodeMessageRule(MessageRule{Mode: ModeDefault}),
		GroupMessageRules:   EncodeMessageRule(MessageRule{Mode: ModeDefault}),
	}
	return db.CreateFilterApp(row)
}
