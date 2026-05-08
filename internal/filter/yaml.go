package filter

import (
	"fmt"

	"moebot-next/internal/models"

	regexp "github.com/dlclark/regexp2"
	"gopkg.in/yaml.v3"
)

// YAMLConfig matches the OneBotFilter project's config.yaml shape so users can
// import/export between the two without manual transformation.
type YAMLConfig struct {
	Server  YAMLServer   `yaml:"server"`
	BotApps []YAMLBotApp `yaml:"bot-apps"`
}

type YAMLServer struct {
	Host       string       `yaml:"host"`
	Port       int          `yaml:"port"`
	Suffix     string       `yaml:"suffix"`
	BotID      string       `yaml:"bot-id"`
	UserAgent  string       `yaml:"user-agent"`
	BufferSize int          `yaml:"buffer-size"`
	SleepTime  float32      `yaml:"sleep-time"`
	Debug      bool         `yaml:"debug"`
	Default    YAMLDefaults `yaml:"default,omitempty"`
}

type YAMLDefaults struct {
	UserID  YAMLIDRule `yaml:"user-id,omitempty"`
	GroupID YAMLIDRule `yaml:"group-id,omitempty"`
}

type YAMLIDRule struct {
	Mode string  `yaml:"mode,omitempty"`
	IDs  []int64 `yaml:"ids,omitempty"`
}

type YAMLMessageRule struct {
	Mode          string   `yaml:"mode,omitempty"`
	Filters       []string `yaml:"filters,omitempty"`
	Prefix        []string `yaml:"prefix,omitempty"`
	PrefixReplace string   `yaml:"prefix-replace,omitempty"`
}

type YAMLBotApp struct {
	Name           string          `yaml:"name"`
	URI            string          `yaml:"uri"`
	AccessToken    string          `yaml:"access-token,omitempty"`
	UserID         YAMLIDRule      `yaml:"user-id,omitempty"`
	GroupID        YAMLIDRule      `yaml:"group-id,omitempty"`
	Message        YAMLMessageRule `yaml:"message,omitempty"`
	PrivateMessage YAMLMessageRule `yaml:"private-message,omitempty"`
	GroupMessage   YAMLMessageRule `yaml:"group-message,omitempty"`
}

// ExportYAML produces a YAMLConfig from the current DB state. defaultTemplate
// supplies server.default.user-id / group-id (replacing the legacy gateway
// fields). templates is used to expand per-app rules when an app references a
// template, so the exported YAML stays compatible with the original
// OneBotFilter project (which has no concept of templates).
func ExportYAML(gateway *models.FilterGateway, defaultTemplate *models.FilterTemplate, templates []models.FilterTemplate, apps []models.FilterApp) ([]byte, error) {
	cfg := YAMLConfig{}
	cfg.Server = YAMLServer{
		Host:       gateway.Host,
		Port:       gateway.Port,
		Suffix:     gateway.Suffix,
		BotID:      gateway.BotID,
		UserAgent:  gateway.UserAgent,
		BufferSize: gateway.BufferSize,
		SleepTime:  gateway.SleepTime,
		Debug:      gateway.Debug,
	}
	if defaultTemplate != nil {
		cfg.Server.Default.UserID = idRuleToYAML(DecodeIDRule(defaultTemplate.UserIDRules))
		cfg.Server.Default.GroupID = idRuleToYAML(DecodeIDRule(defaultTemplate.GroupIDRules))
	}
	tplByID := map[uint]*models.FilterTemplate{}
	for i := range templates {
		t := &templates[i]
		tplByID[t.ID] = t
	}
	cfg.BotApps = make([]YAMLBotApp, 0, len(apps))
	for _, a := range apps {
		ur, gr := a.UserIDRules, a.GroupIDRules
		mr, pr, grm := a.MessageRules, a.PrivateMessageRules, a.GroupMessageRules
		if a.TemplateID != nil {
			if t, ok := tplByID[*a.TemplateID]; ok {
				ur, gr = t.UserIDRules, t.GroupIDRules
				mr, pr, grm = t.MessageRules, t.PrivateMessageRules, t.GroupMessageRules
			}
		}
		cfg.BotApps = append(cfg.BotApps, YAMLBotApp{
			Name:           a.Name,
			URI:            a.URI,
			AccessToken:    a.AccessToken,
			UserID:         idRuleToYAML(DecodeIDRule(ur)),
			GroupID:        idRuleToYAML(DecodeIDRule(gr)),
			Message:        msgRuleToYAML(DecodeMessageRule(mr)),
			PrivateMessage: msgRuleToYAML(DecodeMessageRule(pr)),
			GroupMessage:   msgRuleToYAML(DecodeMessageRule(grm)),
		})
	}
	return yaml.Marshal(cfg)
}

// ParseYAML decodes a YAML payload into a YAMLConfig.
func ParseYAML(data []byte) (*YAMLConfig, error) {
	cfg := &YAMLConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return cfg, nil
}

// ApplyYAMLToModels converts a YAMLConfig into model values ready for DB save.
// It does not write to the DB; the caller decides which apps to overwrite.
//
// defaultTemplate, when non-nil, will be mutated to reflect server.default
// rules from the YAML payload (caller is responsible for persisting).
func ApplyYAMLToModels(cfg *YAMLConfig, gateway *models.FilterGateway, defaultTemplate *models.FilterTemplate) ([]models.FilterApp, *models.FilterGateway) {
	if cfg.Server.Host != "" {
		gateway.Host = cfg.Server.Host
	}
	if cfg.Server.Port != 0 {
		gateway.Port = cfg.Server.Port
	}
	if cfg.Server.Suffix != "" {
		gateway.Suffix = cfg.Server.Suffix
	}
	if cfg.Server.BotID != "" {
		gateway.BotID = cfg.Server.BotID
	}
	if cfg.Server.UserAgent != "" {
		gateway.UserAgent = cfg.Server.UserAgent
	}
	if cfg.Server.BufferSize != 0 {
		gateway.BufferSize = cfg.Server.BufferSize
	}
	if cfg.Server.SleepTime != 0 {
		gateway.SleepTime = cfg.Server.SleepTime
	}
	gateway.Debug = cfg.Server.Debug
	if defaultTemplate != nil {
		defaultTemplate.UserIDRules = EncodeIDRule(yamlToIDRule(cfg.Server.Default.UserID))
		defaultTemplate.GroupIDRules = EncodeIDRule(yamlToIDRule(cfg.Server.Default.GroupID))
	}

	apps := make([]models.FilterApp, 0, len(cfg.BotApps))
	for _, b := range cfg.BotApps {
		apps = append(apps, models.FilterApp{
			Name:                b.Name,
			URI:                 b.URI,
			AccessToken:         b.AccessToken,
			Enabled:             true,
			UserIDRules:         EncodeIDRule(yamlToIDRule(b.UserID)),
			GroupIDRules:        EncodeIDRule(yamlToIDRule(b.GroupID)),
			MessageRules:        EncodeMessageRule(yamlToMsgRule(b.Message)),
			PrivateMessageRules: EncodeMessageRule(yamlToMsgRule(b.PrivateMessage)),
			GroupMessageRules:   EncodeMessageRule(yamlToMsgRule(b.GroupMessage)),
		})
	}
	return apps, gateway
}

func idRuleToYAML(r IDRule) YAMLIDRule {
	if r.IDs == nil {
		r.IDs = []int64{}
	}
	return YAMLIDRule{Mode: r.Mode, IDs: r.IDs}
}

func yamlToIDRule(r YAMLIDRule) IDRule {
	ids := r.IDs
	if ids == nil {
		ids = []int64{}
	}
	return IDRule{Mode: r.Mode, IDs: ids}
}

func msgRuleToYAML(r MessageRule) YAMLMessageRule {
	return YAMLMessageRule{
		Mode:          r.Mode,
		Filters:       r.Filters,
		Prefix:        r.Prefix,
		PrefixReplace: r.PrefixReplace,
	}
}

func yamlToMsgRule(r YAMLMessageRule) MessageRule {
	return MessageRule{
		Mode:          r.Mode,
		Filters:       r.Filters,
		Prefix:        r.Prefix,
		PrefixReplace: r.PrefixReplace,
	}
}

// TestRegex compiles a single regexp2 pattern and tests it against a string.
// Returns (compiled-ok, matched, error-message).
func TestRegex(pattern, text string) (bool, bool, string) {
	re, err := regexp.Compile(pattern, regexp.None)
	if err != nil {
		return false, false, err.Error()
	}
	ok, err := re.MatchString(text)
	if err != nil {
		return true, false, err.Error()
	}
	return true, ok, ""
}
