package autochat

import (
	"sync"
)

// Config 是 autochat 插件的子配置结构，落盘在 data/plugins/autochat.yml。
//
// 与原 autochat 项目相比：
//   - 删除 Gemini / Qdrant / 历史 SiliconFlow 段
//   - 新增 LLM.Providers 描述 OpenAI 兼容与 Anthropic 双 provider
//   - Embedding / Rerank 改为 OpenAI 兼容 endpoint
//   - 向量库改用 SQLite + sqlite-vec
type Config struct {
	LogLevel string `yaml:"log_level"`

	LLM struct {
		Providers struct {
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
		} `yaml:"providers"`
		Models    []string `yaml:"models"`
		MaxTokens int      `yaml:"max_tokens"`
		Reasoning bool     `yaml:"reasoning"`
		Timeout   int      `yaml:"timeout"`
	} `yaml:"llm"`

	Vector struct {
		Enabled    bool `yaml:"enabled"`
		Dimensions int  `yaml:"dimensions"`
		TopK       int  `yaml:"top_k"`
	} `yaml:"vector"`

	Embedding struct {
		Enabled    bool   `yaml:"enabled"`
		BaseURL    string `yaml:"base_url"`
		APIKey     string `yaml:"api_key"`
		Model      string `yaml:"model"`
		Dimensions int    `yaml:"dimensions"`
		Timeout    int    `yaml:"timeout"`
	} `yaml:"embedding"`

	Rerank struct {
		Enabled   bool    `yaml:"enabled"`
		BaseURL   string  `yaml:"base_url"`
		APIKey    string  `yaml:"api_key"`
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
	} `yaml:"chat"`
}

var (
	cfg     *Config
	cfgMu   sync.RWMutex
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

// applyDefaults 在 cfg 上填补未设置的字段。
func applyDefaults(c *Config) {
	if c.LLM.Providers.OpenAI.BaseURL == "" {
		c.LLM.Providers.OpenAI.BaseURL = "https://api.openai.com/v1"
	}
	if c.LLM.Providers.OpenAI.Timeout <= 0 {
		c.LLM.Providers.OpenAI.Timeout = 60
	}
	if c.LLM.Providers.Anthropic.BaseURL == "" {
		c.LLM.Providers.Anthropic.BaseURL = "https://api.anthropic.com"
	}
	if c.LLM.Providers.Anthropic.AnthropicVersion == "" {
		c.LLM.Providers.Anthropic.AnthropicVersion = "2023-06-01"
	}
	if c.LLM.Providers.Anthropic.Timeout <= 0 {
		c.LLM.Providers.Anthropic.Timeout = 60
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
	if c.Embedding.BaseURL == "" {
		c.Embedding.BaseURL = c.LLM.Providers.OpenAI.BaseURL
	}
	if c.Embedding.APIKey == "" {
		c.Embedding.APIKey = c.LLM.Providers.OpenAI.APIKey
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
		c.Chat.Prompt.Persona["default"] = "你是一个有用的AI助手。"
	}
	if c.Chat.Prompt.Framework == "" {
		c.Chat.Prompt.Framework = defaultFramework
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
