package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ReadYAMLInto 将插件子配置文件解码到 out 指针指向的结构。
// 若文件不存在则保持 out 为零值并返回 nil（视为"使用默认值"）。
//
// 调用者通常在 Init 内使用，例如：
//
//	var cfg moesekai.Config
//	if err := plugin.ReadYAMLInto(ctx.PluginConfigPath, &cfg); err != nil { ... }
func ReadYAMLInto(path string, out any) error {
	if path == "" {
		return fmt.Errorf("plugin: empty config path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// WriteYAMLFrom 把 in 序列化为 YAML 写入 path（必要时创建父目录）。
func WriteYAMLFrom(path string, in any) error {
	if path == "" {
		return fmt.Errorf("plugin: empty config path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
