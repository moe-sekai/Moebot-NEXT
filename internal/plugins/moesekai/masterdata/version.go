package masterdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// version.go — current_version.json / versions.json fetching & local persistence
//
// 大多数 masterdata 镜像在仓库根/特定路径提供版本 JSON：
//   - MoeSekai / Haruki: <root>/versions/current_version.json (camelCase)
//   - 8823 / Sekai-World: <root>/versions.json (snake_case 或 camelCase)
//
// 具体 URL 由 config.ResolveMasterdata 写入 ResolvedMasterdata.VersionURLs，
// 这里只负责拉取与解析，并在本地 <localPath>/data_version.json 持久化最近一次
// 成功更新的版本，用于决定下次刷新时是否跳过全量拉取。
// ---------------------------------------------------------------------------

// RemoteVersionInfo 同时支持 camelCase（current_version.json）与 snake_case
// （versions.json）两种字段命名。解析时会先按 camelCase 反序列化，再用
// snake_case 的别名结构补全空字段，覆盖两类源。
type RemoteVersionInfo struct {
	AppHash          string `json:"appHash,omitempty"`
	SystemProfile    string `json:"systemProfile,omitempty"`
	AppVersion       string `json:"appVersion,omitempty"`
	MultiPlayVersion string `json:"multiPlayVersion,omitempty"`
	AssetVersion     string `json:"assetVersion,omitempty"`
	AppVersionStatus string `json:"appVersionStatus,omitempty"`
	DataVersion      string `json:"dataVersion,omitempty"`
	AssetHash        string `json:"assetHash,omitempty"`
}

// snakeVersion 对应 8823 / Sekai-World 风格 versions.json 的 snake_case 字段。
type snakeVersion struct {
	AppHash      string `json:"app_hash"`
	AppVersion   string `json:"app_version"`
	DataVersion  string `json:"data_version"`
	AssetVersion string `json:"asset_version"`
	AssetHash    string `json:"asset_hash"`
}

// storedVersion 是 <localPath>/data_version.json 的磁盘格式，包含一份远端
// version 信息的拷贝以及最近一次成功更新时间，便于排查与展示。
type storedVersion struct {
	RemoteVersionInfo
	UpdatedAt string `json:"updatedAt,omitempty"`
}

const versionFilename = "data_version.json"

// parseRemoteVersion 解析 raw 字节为 RemoteVersionInfo，兼容 camelCase 与
// snake_case 两种命名。dataVersion 缺失时返回错误。
func parseRemoteVersion(raw []byte) (*RemoteVersionInfo, error) {
	var v RemoteVersionInfo
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	if strings.TrimSpace(v.DataVersion) == "" {
		var alt snakeVersion
		if err := json.Unmarshal(raw, &alt); err == nil {
			if v.DataVersion == "" {
				v.DataVersion = alt.DataVersion
			}
			if v.AppVersion == "" {
				v.AppVersion = alt.AppVersion
			}
			if v.AssetVersion == "" {
				v.AssetVersion = alt.AssetVersion
			}
			if v.AppHash == "" {
				v.AppHash = alt.AppHash
			}
			if v.AssetHash == "" {
				v.AssetHash = alt.AssetHash
			}
		}
	}
	if strings.TrimSpace(v.DataVersion) == "" {
		return nil, fmt.Errorf("dataVersion missing")
	}
	return &v, nil
}

// fetchRemoteVersion 依次尝试 versionURLs，命中即返回。全部失败时返回最后一次错误。
// 当 versionURLs 为空时返回 nil + 描述性错误，由调用方决定如何处理（一般是回退到
// 不带版本检查的全量拉取）。
func (l *Loader) fetchRemoteVersion(versionURLs []string) (*RemoteVersionInfo, string, error) {
	if len(versionURLs) == 0 {
		return nil, "", fmt.Errorf("no version endpoint configured for source")
	}
	var lastErr error
	for _, url := range versionURLs {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		raw, err := l.fetchRemote(url)
		if err != nil {
			lastErr = err
			continue
		}
		v, err := parseRemoteVersion(raw)
		if err != nil {
			lastErr = fmt.Errorf("parse %s: %w", url, err)
			continue
		}
		return v, url, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no usable version endpoint")
	}
	return nil, "", lastErr
}

// loadStoredVersion 读取本地 data_version.json，失败/不存在时返回 nil。
func loadStoredVersion(localPath string) *storedVersion {
	if localPath == "" {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(localPath, versionFilename))
	if err != nil {
		return nil
	}
	var v storedVersion
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return &v
}

// saveStoredVersion 把最新的远端版本信息持久化到 localPath/data_version.json。
func saveStoredVersion(localPath string, info RemoteVersionInfo) error {
	if localPath == "" {
		return fmt.Errorf("local path not configured")
	}
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", localPath, err)
	}
	sv := storedVersion{
		RemoteVersionInfo: info,
		UpdatedAt:         time.Now().Format(time.RFC3339),
	}
	raw, err := json.MarshalIndent(sv, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(localPath, versionFilename), raw, 0644)
}

// CurrentDataVersion 返回本地缓存的 dataVersion，无缓存时返回空字符串。
// 主要给 /update 命令展示本地版本用。
func (l *Loader) CurrentDataVersion() string {
	cfg, _ := l.configSnapshot()
	sv := loadStoredVersion(strings.TrimSpace(cfg.LocalPath))
	if sv == nil {
		return ""
	}
	return sv.DataVersion
}
