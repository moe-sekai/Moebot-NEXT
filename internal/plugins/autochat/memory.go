package autochat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// MemoryItem 单用户画像。
type MemoryItem struct {
	Text string `json:"text"`
}

// SummaryItem 群对话总结条目。
type SummaryItem struct {
	Content string `json:"content"`
	Time    int64  `json:"time"`
}

// GroupMemory 一个群的本地记忆文件结构。
type GroupMemory struct {
	UMS       map[string]MemoryItem `json:"ums"`
	Summaries []SummaryItem         `json:"summaries,omitempty"`
}

// MemoryManager 文件系统的本地记忆管理器。
// 记忆按 (groupID, templateName) 分桶存储：
//
//	<data_dir>/memory/<groupID>/<templateName>.json
//
// templateName 为空时使用特殊文件名 "__default__"。
// 旧格式 memory/<groupID>.json 在首次读取时自动迁移到 memory/<groupID>/__default__.json。
type MemoryManager struct {
	mu      sync.RWMutex
	rootDir string
}

const defaultMemoryFile = "__default__"

func newMemoryManager(rootDir string) *MemoryManager {
	return &MemoryManager{rootDir: rootDir}
}

// memoryPath 返回指定群+模板的记忆文件路径。
// templateName 为空时使用 __default__。
func (m *MemoryManager) memoryPath(groupID int64, templateName string) string {
	dir := filepath.Join(m.rootDir, "memory", fmt.Sprintf("%d", groupID))
	file := templateName
	if file == "" {
		file = defaultMemoryFile
	}
	return filepath.Join(dir, file+".json")
}

// legacyMemoryPath 返回旧版记忆文件路径（兼容迁移用）。
func (m *MemoryManager) legacyMemoryPath(groupID int64) string {
	return filepath.Join(m.rootDir, "memory", fmt.Sprintf("%d.json", groupID))
}

// migrateLegacyFile 首次读取时自动将旧文件 memory/<gid>.json
// 迁移到 memory/<gid>/__default__.json。
func (m *MemoryManager) migrateLegacyFile(groupID int64) {
	legacyPath := m.legacyMemoryPath(groupID)
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return
	}
	newPath := m.memoryPath(groupID, "")
	if _, err := os.Stat(newPath); err == nil {
		// 新文件已存在，旧文件可能是遗留垃圾，直接删掉
		os.Remove(legacyPath)
		return
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return
	}
	_ = os.Rename(legacyPath, newPath)
}

func (m *MemoryManager) load(groupID int64, templateName string) (GroupMemory, error) {
	var gm GroupMemory
	// 仅对默认模板执行旧文件迁移
	if templateName == "" {
		m.migrateLegacyFile(groupID)
	}
	path := m.memoryPath(groupID, templateName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return gm, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return gm, err
	}
	if err := json.Unmarshal(data, &gm); err != nil {
		return gm, err
	}
	return gm, nil
}

func (m *MemoryManager) save(groupID int64, templateName string, gm GroupMemory) error {
	path := m.memoryPath(groupID, templateName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(gm, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (m *MemoryManager) GetUserMemory(groupID, userID int64, templateName string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	gm, err := m.load(groupID, templateName)
	if err != nil {
		return "", err
	}
	if item, ok := gm.UMS[fmt.Sprintf("%d", userID)]; ok {
		return item.Text, nil
	}
	// 兼容旧数据：如果当前模板没有用户画像，回退查默认模板（旧数据）
	if templateName != "" {
		gm, err := m.load(groupID, "")
		if err != nil {
			return "", err
		}
		if item, ok := gm.UMS[fmt.Sprintf("%d", userID)]; ok {
			return item.Text, nil
		}
	}
	return "", nil
}

func (m *MemoryManager) GetRecentSummaries(groupID int64, templateName string, limit int) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	gm, err := m.load(groupID, templateName)
	if err != nil {
		return nil, err
	}
	// 兼容旧数据：如果当前模板没有总结，回退查默认模板（旧数据）
	if len(gm.Summaries) == 0 && templateName != "" {
		gm, err = m.load(groupID, "")
		if err != nil {
			return nil, err
		}
	}
	if len(gm.Summaries) == 0 {
		return []string{}, nil
	}
	if limit > len(gm.Summaries) {
		limit = len(gm.Summaries)
	}
	out := make([]string, limit)
	start := len(gm.Summaries) - limit
	for i := 0; i < limit; i++ {
		out[i] = gm.Summaries[start+i].Content
	}
	return out, nil
}

func (m *MemoryManager) AddSummary(groupID int64, templateName string, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	gm, _ := m.load(groupID, templateName)
	if gm.UMS == nil {
		gm.UMS = map[string]MemoryItem{}
	}
	gm.Summaries = append(gm.Summaries, SummaryItem{Content: content, Time: currentTimestamp()})
	if len(gm.Summaries) > 20 {
		gm.Summaries = gm.Summaries[len(gm.Summaries)-20:]
	}
	if err := m.save(groupID, templateName, gm); err != nil {
		return err
	}
	if vc := GetVectorClient(); vc != nil && vc.IsEnabled() {
		go func() { _ = vc.UpsertSummary(groupID, templateName, content) }()
	}
	return nil
}

func (m *MemoryManager) UpdateUserMemory(groupID, userID int64, templateName, text string) error {
	return m.UpdateUserMemoryWithName(groupID, userID, templateName, "", text)
}

func (m *MemoryManager) UpdateUserMemoryWithName(groupID, userID int64, templateName, userName, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	gm, _ := m.load(groupID, templateName)
	if gm.UMS == nil {
		gm.UMS = map[string]MemoryItem{}
	}
	gm.UMS[fmt.Sprintf("%d", userID)] = MemoryItem{Text: text}
	if err := m.save(groupID, templateName, gm); err != nil {
		return err
	}
	if vc := GetVectorClient(); vc != nil && vc.IsEnabled() {
		go func() { _ = vc.UpsertUserMemory(groupID, userID, userName, templateName, text) }()
	}
	return nil
}

// DeleteTemplateMemories 删除指定群+模板的所有本地记忆文件。
func (m *MemoryManager) DeleteTemplateMemories(groupID int64, templateName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := m.memoryPath(groupID, templateName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// DeleteAllTemplateMemories 删除指定模板在所有群的本地记忆文件。
func (m *MemoryManager) DeleteAllTemplateMemories(templateName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	memoryDir := filepath.Join(m.rootDir, "memory")
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		target := filepath.Join(memoryDir, e.Name(), templateName+".json")
		_ = os.Remove(target)
	}
	return nil
}
