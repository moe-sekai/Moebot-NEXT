package autochat

import (
	"fmt"
	"sync"
)

// Config 是 autochat 插件的子配置结构，落盘在 data/plugins/autochat.yml。
//
// 与原 autochat 项目相比：
//   - 删除 Gemini / Qdrant / 历史 SiliconFlow 段
//   - 新增 LLM.Providers 描述 OpenAI 兼容与 Anthropic 双 provider
//   - Embedding / Rerank 改为 OpenAI 兼容 endpoint
//   - 向量库改用 SQLite + sqlite-vec
//
// ProviderConfig 描述一个用户接入的 LLM 提供商实例。
//
// Name 是用户自定义的标识（例如 "openai-main"、"siliconflow"、"claude"）；
// 在 LLM.Models 中以 "<Name>:<modelId>" 形式被引用。
// Type 取值 "openai" / "anthropic"，分别走 OpenAI 兼容协议或 Anthropic Messages API。
type ProviderConfig struct {
	Name             string `yaml:"name"`
	Type             string `yaml:"type"` // openai | anthropic
	BaseURL          string `yaml:"base_url"`
	APIKey           string `yaml:"api_key"`
	Timeout          int    `yaml:"timeout"`
	AnthropicVersion string `yaml:"anthropic_version,omitempty"`
}

// legacyProviders 是 v0 配置中的固定双 provider 段，仅用于向后兼容自动迁移。
type legacyProviders struct {
	OpenAI struct {
		BaseURL string `yaml:"base_url"`
		APIKey  string `yaml:"api_key"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"openai"`
	Anthropic struct {
		BaseURL          string `yaml:"base_url"`
		APIKey           string `yaml:"api_key"`
		AnthropicVersion string `yaml:"anthropic_version"`
		Timeout          int    `yaml:"timeout"`
	} `yaml:"anthropic"`
}

type Config struct {
	LogLevel string `yaml:"log_level"`

	LLM struct {
		// ProviderList 是 v1 字段：用户可自由增删提供商实例。
		ProviderList []ProviderConfig `yaml:"provider_list"`
		// Providers 是 v0 兼容字段：仅在 ProviderList 为空时被读取并迁移。
		Providers legacyProviders `yaml:"providers,omitempty"`
		Models    []string        `yaml:"models"`
		MaxTokens int             `yaml:"max_tokens"`
		Reasoning bool            `yaml:"reasoning"`
		Timeout   int             `yaml:"timeout"`
	} `yaml:"llm"`

	Vector struct {
		Enabled    bool `yaml:"enabled"`
		Dimensions int  `yaml:"dimensions"`
		TopK       int  `yaml:"top_k"`
	} `yaml:"vector"`

	Embedding struct {
		Enabled    bool   `yaml:"enabled"`
		Provider   string `yaml:"provider"` // 引用 ProviderList[].Name；为空则使用下方独立 base_url/api_key
		BaseURL    string `yaml:"base_url,omitempty"`
		APIKey     string `yaml:"api_key,omitempty"`
		Model      string `yaml:"model"`
		Dimensions int    `yaml:"dimensions"`
		Timeout    int    `yaml:"timeout"`
	} `yaml:"embedding"`

	Rerank struct {
		Enabled   bool    `yaml:"enabled"`
		Provider  string  `yaml:"provider"`
		BaseURL   string  `yaml:"base_url,omitempty"`
		APIKey    string  `yaml:"api_key,omitempty"`
		Model     string  `yaml:"model"`
		Timeout   int     `yaml:"timeout"`
		Threshold float64 `yaml:"threshold"`
	} `yaml:"rerank"`

	ImageCaption struct {
		Enabled   bool   `yaml:"enabled"`
		Model     string `yaml:"model"` // 形如 openai:gpt-4o-mini
		Timeout   int    `yaml:"timeout"`
		MaxTokens int    `yaml:"max_tokens"`
		Prompt    string `yaml:"prompt"`
	} `yaml:"image_caption"`

	RAGSummary struct {
		Enabled   bool   `yaml:"enabled"`
		Model     string `yaml:"model"`
		Timeout   int    `yaml:"timeout"`
		MaxTokens int    `yaml:"max_tokens"`
		Prompt    string `yaml:"prompt"`
	} `yaml:"rag_summary"`

	Chat struct {
		Willing struct {
			Threshold       float64            `yaml:"threshold"`
			GroupThresholds map[string]float64 `yaml:"group_thresholds"`
		} `yaml:"willing"`

		ChatCDSeconds int `yaml:"chat_cd_seconds"`
		TTSCDSeconds  int `yaml:"tts_cd_seconds"`

		ContextSize    int `yaml:"context_size"`     // 进 LLM 的最近消息条数
		BufferLimit    int `yaml:"buffer_limit"`     // MessageBuffer 容量
		ReplyMaxLength int `yaml:"reply_max_length"` // 单条回复长度上限

		Prompt struct {
			Persona   map[string]string `yaml:"persona"`
			Framework string            `yaml:"framework"`
		} `yaml:"prompt"`

		Keywords []string `yaml:"keywords"`

		// IgnorePrefixes：以这些字符/字串开头的“纯文本”消息将不会触发 autochat，
		// 避免其它插件命令（/查歌、#help 等）误触自动对话。
		IgnorePrefixes []string `yaml:"ignore_prefixes"`
		// IgnorePatterns：额外的正则表达式列表，匹配到的纯文本会被忽略，
		// 用于覆盖那些不以固定前缀开头的命令（例如纯中文指令名）。
		IgnorePatterns []string `yaml:"ignore_patterns"`
	} `yaml:"chat"`
}

var (
	cfg   *Config
	cfgMu sync.RWMutex
)

// GetConfig 返回当前生效配置（仅在 Init 后非 nil）。
func GetConfig() *Config {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return cfg
}

func setConfig(c *Config) {
	cfgMu.Lock()
	cfg = c
	cfgMu.Unlock()
}

// applyDefaults 在 cfg 上填补未设置的字段；同时把 v0 的 Providers 双段迁移到 ProviderList。
func applyDefaults(c *Config) {
	// ---- v0 -> v1 迁移 ----
	if len(c.LLM.ProviderList) == 0 {
		lp := c.LLM.Providers
		if lp.OpenAI.BaseURL != "" || lp.OpenAI.APIKey != "" {
			c.LLM.ProviderList = append(c.LLM.ProviderList, ProviderConfig{
				Name: "openai", Type: "openai",
				BaseURL: lp.OpenAI.BaseURL, APIKey: lp.OpenAI.APIKey, Timeout: lp.OpenAI.Timeout,
			})
		}
		if lp.Anthropic.BaseURL != "" || lp.Anthropic.APIKey != "" {
			c.LLM.ProviderList = append(c.LLM.ProviderList, ProviderConfig{
				Name: "anthropic", Type: "anthropic",
				BaseURL: lp.Anthropic.BaseURL, APIKey: lp.Anthropic.APIKey,
				AnthropicVersion: lp.Anthropic.AnthropicVersion, Timeout: lp.Anthropic.Timeout,
			})
		}
		// 清空 v0 字段，避免下次写盘还重复
		c.LLM.Providers = legacyProviders{}
	}
	// 为 ProviderList 内每条填默认值并校正 Type
	for i := range c.LLM.ProviderList {
		pc := &c.LLM.ProviderList[i]
		if pc.Name == "" {
			pc.Name = fmt.Sprintf("provider-%d", i+1)
		}
		if pc.Type != "anthropic" {
			pc.Type = "openai"
		}
		if pc.Timeout <= 0 {
			pc.Timeout = 60
		}
		if pc.Type == "anthropic" && pc.AnthropicVersion == "" {
			pc.AnthropicVersion = "2023-06-01"
		}
		if pc.BaseURL == "" {
			if pc.Type == "anthropic" {
				pc.BaseURL = "https://api.anthropic.com"
			} else {
				pc.BaseURL = "https://api.openai.com/v1"
			}
		}
	}
	if c.LLM.MaxTokens <= 0 {
		c.LLM.MaxTokens = 2048
	}
	if c.LLM.Timeout <= 0 {
		c.LLM.Timeout = 120
	}
	if len(c.LLM.Models) == 0 {
		c.LLM.Models = []string{"openai:gpt-4o-mini"}
	}
	if c.Vector.Dimensions <= 0 {
		c.Vector.Dimensions = 1536
	}
	if c.Vector.TopK <= 0 {
		c.Vector.TopK = 5
	}
	if c.Embedding.Model == "" {
		c.Embedding.Model = "text-embedding-3-small"
	}
	if c.Embedding.Dimensions <= 0 {
		c.Embedding.Dimensions = c.Vector.Dimensions
	}
	if c.Embedding.Timeout <= 0 {
		c.Embedding.Timeout = 30
	}
	if c.Rerank.Threshold <= 0 {
		c.Rerank.Threshold = 0.3
	}
	if c.Rerank.Timeout <= 0 {
		c.Rerank.Timeout = 15
	}
	if c.ImageCaption.Timeout <= 0 {
		c.ImageCaption.Timeout = 20
	}
	if c.ImageCaption.MaxTokens <= 0 {
		c.ImageCaption.MaxTokens = 80
	}
	if c.ImageCaption.Prompt == "" {
		c.ImageCaption.Prompt = "请简短描述这张{sub_type}的内容用于无障碍辅助理解，不超过50字。"
	}
	if c.RAGSummary.Timeout <= 0 {
		c.RAGSummary.Timeout = 30
	}
	if c.RAGSummary.MaxTokens <= 0 {
		c.RAGSummary.MaxTokens = 256
	}
	if c.RAGSummary.Prompt == "" {
		c.RAGSummary.Prompt = "请总结以下对话核心话题，用于检索相关历史，控制在80字以内：\n```\n{text}\n```"
	}
	if c.Chat.Willing.Threshold <= 0 {
		c.Chat.Willing.Threshold = 2.5
	}
	if c.Chat.ChatCDSeconds <= 0 {
		c.Chat.ChatCDSeconds = 3
	}
	if c.Chat.TTSCDSeconds <= 0 {
		c.Chat.TTSCDSeconds = 10
	}
	if c.Chat.ContextSize <= 0 {
		c.Chat.ContextSize = 10
	}
	if c.Chat.BufferLimit <= 0 {
		c.Chat.BufferLimit = 20
	}
	if c.Chat.ReplyMaxLength <= 0 {
		c.Chat.ReplyMaxLength = 512
	}
	if c.Chat.Prompt.Persona == nil {
		c.Chat.Prompt.Persona = map[string]string{}
	}
	if _, ok := c.Chat.Prompt.Persona["default"]; !ok {
		c.Chat.Prompt.Persona["default"] = `你现在扮演的是《世界计划》中的花里实乃理，正在使用社交软件（QQ）和用户们聊天，请按照以下设定进行角色扮演：
        基本信息：
        年龄：17岁（高中二年级）
        所属组合：MORE MORE JUMP!
        身份：新人偶像，MMJ的精神支柱
        性格特征：
        极度乐观开朗，充满正能量
        即使遇到挫折也绝不放弃，总是积极向前
        天然呆，经常会做出笨拙的举动，不喜欢吃西兰花
        对朋友非常真诚，总是为他人着想
        对偶像工作充满热情和憧憬
        极度崇拜桐谷遥（遥前辈），经常提到她
        虽然唱歌跳舞技巧还不够完美，但努力程度无人能及
        喜欢鼓励他人，传播希望，避免长篇大论 输出尽量限制在40字以内`
	}
	if c.Chat.Prompt.Framework == "" {
		c.Chat.Prompt.Framework = defaultFramework
	}
	if c.Chat.IgnorePrefixes == nil {
		c.Chat.IgnorePrefixes = []string{"/", "#", "!", "！", ".", "。", ">", "&"}
	}
}

const defaultFramework = `# Role
你是一个正在群聊中的真实用户，请基于人设进行沉浸式聊天。
- ID: {self_id}
- 昵称: "{self_name}"
- 人设:
` + "```\n{persona}\n```" + `

# Memory
{recent_text}
{um_text}
{sm_text}
{rag_mem_text}
{rag_summary_text}

# Rules
1. 真实感：说话简短口语化，控制在一到两句话；避免长篇大论。
2. 仅在必要时 [@qqid] 提及他人。
3. 遇到敏感/冒犯话题选择忽略（回复空字符串）。
`
