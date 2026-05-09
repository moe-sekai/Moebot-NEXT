// Package autochat 是 Moebot NEXT 的官方 AutoChat 插件：在群里以 LLM 进行
// 角色扮演式对话，支持 OpenAI 兼容与 Anthropic Messages API 两类 provider，
// 并基于 SQLite + sqlite-vec 做用户记忆 / 历史总结的 RAG 检索。
//
// 该插件读 `<plugin_data_dir>/autochat.yml` 作为子配置；状态（白名单 / 冷却 /
// token 统计）落在 `<plugin_data_dir>/autochat/db.json`，群级人设记忆落在
// `<plugin_data_dir>/autochat/memory/<group>.json`。
package autochat

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"moebot-next/internal/database"
	"moebot-next/internal/filter"
	"moebot-next/internal/plugin"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const PluginName = "autochat"

// pluginImpl 实现 plugin.Plugin / plugin.Configurable，并持有运行期状态。
// 字段集合较多是因为整个 autochat 子系统都挂在它上面（白名单 / 冷却 /
// 缓冲 / 阈值 / 会话 / 记忆 / Token 统计 / 当前 ZeroBot Engine）。
type pluginImpl struct {
	mu         sync.RWMutex
	configPath string

	fileDB        *FileDB
	chatWhiteList *GroupWhiteList
	autoWhiteList *GroupWhiteList
	chatCD        *ColdDown
	tokenStats    *TokenStats
	messageBuffer *MessageBuffer
	sessions      *SessionManager
	memory        *MemoryManager
	thresholds    map[int64]float64

	filterMgr *filter.Manager // 用于查询本插件的 internal FilterApp 规则
	engine    *zero.Engine    // 独立 ZeroBot Engine，禁用插件时调用 Delete 注销
}

// filterAppName 返回本插件在 filter 网关中的 internal app 名字。
func (p *pluginImpl) filterAppName() string { return filter.InternalAppName(PluginName) }

// allowedByFilter 查询 filter 网关：当前消息是否被本插件的 internal app 放行。
// 当 filter 未启用 / 该 app 未 seed 时返回 true（不阻塞）。
func (p *pluginImpl) allowedByFilter(groupID, userID int64, isPrivate bool, raw string) bool {
	if p.filterMgr == nil {
		return true
	}
	return p.filterMgr.AllowMessage(p.filterAppName(), groupID, userID, isPrivate, raw)
}

func (p *pluginImpl) Manifest() plugin.Manifest {
	return plugin.Manifest{
		Name:          PluginName,
		Title:         "AutoChat (LLM 群聊)",
		Version:       "0.1.0",
		Author:        "Moebot Team",
		Category:      plugin.CategoryOfficial,
		Description:   "在群聊中扮演角色的 LLM 对话插件，支持 OpenAI 兼容 + Anthropic + sqlite-vec RAG。",
		Homepage:      "https://github.com/moe-sekai/Moebot-NEXT",
		SettingsRoute: "/plugins/autochat",
		Tags:          []string{"llm", "official"},
		Settings: []plugin.SettingField{
			{Key: "openai_base_url", Label: "OpenAI 兼容 BaseURL", Type: "string", Group: "OpenAI",
				Description: "可指向 OpenAI / Azure / SiliconFlow / Ollama / 任一 OpenAI 兼容 endpoint。"},
			{Key: "openai_api_key", Label: "OpenAI 兼容 API Key", Type: "string", Group: "OpenAI"},
			{Key: "anthropic_api_key", Label: "Anthropic API Key", Type: "string", Group: "Anthropic"},
			{Key: "primary_model", Label: "首选模型", Type: "string", Group: "对话",
				Description: "格式 <provider>:<model>，如 openai:gpt-4o-mini 或 anthropic:claude-3-5-haiku-20241022。"},
			{Key: "system_prompt", Label: "默认 System Prompt（人设）", Type: "textarea", Group: "对话"},
		},
	}
}

// GetSettings 返回 5 项常用配置当前值。
func (p *pluginImpl) GetSettings() (map[string]any, error) {
	c := GetConfig()
	if c == nil {
		return map[string]any{}, nil
	}
	primaryModel := ""
	if len(c.LLM.Models) > 0 {
		primaryModel = c.LLM.Models[0]
	}
	return map[string]any{
		"openai_base_url":   c.LLM.Providers.OpenAI.BaseURL,
		"openai_api_key":    c.LLM.Providers.OpenAI.APIKey,
		"anthropic_api_key": c.LLM.Providers.Anthropic.APIKey,
		"primary_model":     primaryModel,
		"system_prompt":     c.Chat.Prompt.Persona["default"],
	}, nil
}

// UpdateSettings 写回 5 项常用配置到 yaml + 内存。其余字段留给 YAML 编辑器。
func (p *pluginImpl) UpdateSettings(values map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.configPath == "" {
		return errors.New("autochat: plugin not initialized")
	}
	c := GetConfig()
	if c == nil {
		return errors.New("autochat: config not loaded")
	}
	asString := func(k string) (string, bool) {
		if v, ok := values[k]; ok {
			if s, ok := v.(string); ok {
				return s, true
			}
		}
		return "", false
	}
	if v, ok := asString("openai_base_url"); ok {
		c.LLM.Providers.OpenAI.BaseURL = v
	}
	if v, ok := asString("openai_api_key"); ok {
		c.LLM.Providers.OpenAI.APIKey = v
	}
	if v, ok := asString("anthropic_api_key"); ok {
		c.LLM.Providers.Anthropic.APIKey = v
	}
	if v, ok := asString("primary_model"); ok && v != "" {
		// 把 v 提前到 models 列表头
		newModels := []string{v}
		for _, m := range c.LLM.Models {
			if m != v {
				newModels = append(newModels, m)
			}
		}
		c.LLM.Models = newModels
	}
	if v, ok := asString("system_prompt"); ok {
		if c.Chat.Prompt.Persona == nil {
			c.Chat.Prompt.Persona = map[string]string{}
		}
		c.Chat.Prompt.Persona["default"] = v
	}
	if err := plugin.WriteYAMLFrom(p.configPath, c); err != nil {
		return err
	}
	// 更新内存中 Provider 客户端
	p.applyProviders(c)
	return nil
}

func (p *pluginImpl) applyProviders(c *Config) {
	clearProviders()
	registerProvider(newOpenAIProvider(
		c.LLM.Providers.OpenAI.BaseURL,
		c.LLM.Providers.OpenAI.APIKey,
		c.LLM.Providers.OpenAI.Timeout,
	))
	registerProvider(newAnthropicProvider(
		c.LLM.Providers.Anthropic.BaseURL,
		c.LLM.Providers.Anthropic.APIKey,
		c.LLM.Providers.Anthropic.AnthropicVersion,
		c.LLM.Providers.Anthropic.Timeout,
	))
	initEmbeddingClient(c)
	initRerankClient(c)
}

// Init 加载配置 → 初始化客户端 / 状态 → 注册处理器 → 登记关闭钩子。
func (p *pluginImpl) Init(ctx *plugin.Context) error {
	db, _ := ctx.DB.(*database.DB)
	if db == nil || db.DB == nil {
		return errors.New("autochat: database not available in plugin context")
	}
	filterMgr, _ := ctx.Filter.(*filter.Manager)

	// 0) 在 Filter 网关中 seed 本插件对应的 internal app；让控制台「Filter」
	//    页面能够独立分配模板/规则。已存在时不覆盖用户配置。
	if err := filter.SeedInternalApp(db, PluginName, "AutoChat"); err != nil {
		log.Warn().Err(err).Msg("[autochat] 创建 internal filter app 失败")
	} else if filterMgr != nil && filterMgr.IsRunning() {
		_ = filterMgr.Reload(ctx.Ctx)
	}

	// 1) 读子配置（缺失时落默认到磁盘）
	var c Config
	if err := plugin.ReadYAMLInto(ctx.PluginConfigPath, &c); err != nil {
		log.Warn().Err(err).Str("path", ctx.PluginConfigPath).Msg("[autochat] 读取子配置失败，使用默认")
	}
	applyDefaults(&c)
	if _, err := os.Stat(ctx.PluginConfigPath); os.IsNotExist(err) {
		_ = plugin.WriteYAMLFrom(ctx.PluginConfigPath, &c)
	}
	setConfig(&c)

	p.mu.Lock()
	p.configPath = ctx.PluginConfigPath
	p.mu.Unlock()

	// 2) Provider / Embedding / Rerank
	p.applyProviders(&c)

	// 3) 数据目录与状态
	dataDir := filepath.Join(ctx.PluginDataDir, "autochat")
	if err := os.MkdirAll(filepath.Join(dataDir, "memory"), 0o755); err != nil {
		log.Warn().Err(err).Str("dir", dataDir).Msg("[autochat] 创建数据目录失败")
	}
	p.fileDB = NewFileDB(filepath.Join(dataDir, "db.json"))
	p.chatWhiteList = NewGroupWhiteList(p.fileDB, "chat")
	p.autoWhiteList = NewGroupWhiteList(p.fileDB, "autochat")
	p.chatCD = NewColdDown(c.Chat.ChatCDSeconds)
	p.tokenStats = NewTokenStats(p.fileDB)
	p.messageBuffer = NewMessageBuffer(c.Chat.BufferLimit)
	p.sessions = NewSessionManager(12 * time.Hour)
	p.memory = newMemoryManager(dataDir)
	p.thresholds = map[int64]float64{}
	p.filterMgr = filterMgr

	// 4) 向量库（共享 moebot 的主 SQLite）
	if err := initVectorClient(&c, db.DB); err != nil {
		log.Warn().Err(err).Msg("[autochat] 向量库初始化失败，RAG 将禁用")
	}

	// 5) 注册 ZeroBot 处理器
	p.engine = p.registerHandlers()

	// 6) 关闭钩子
	ctx.OnShutdown(func() {
		if p.engine != nil {
			p.engine.Delete()
			p.engine = nil
		}
		if p.sessions != nil {
			p.sessions.Close()
		}
		clearProviders()
		log.Info().Msg("[autochat] 已停止")
	})

	log.Info().Int("models", len(c.LLM.Models)).Bool("vector", c.Vector.Enabled).Msg("[autochat] 已启动")
	return nil
}

func init() {
	plugin.Register(&pluginImpl{})
}
