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

// MemoryManager 文件系统的本地记忆管理器（位于 <data_dir>/memory/<group>.json）。
// 同时会在向量库可用时把更新同步到 RAG。
type MemoryManager struct {
	mu      sync.RWMutex
	rootDir string
}

func newMemoryManager(rootDir string) *MemoryManager {
	return &MemoryManager{rootDir: rootDir}
}

func (m *MemoryManager) memoryPath(groupID int64) string {
	return filepath.Join(m.rootDir, "memory", fmt.Sprintf("%d.json", groupID))
}

func (m *MemoryManager) load(groupID int64) (GroupMemory, error) {
	var gm GroupMemory
	path := m.memoryPath(groupID)
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

func (m *MemoryManager) save(groupID int64, gm GroupMemory) error {
	path := m.memoryPath(groupID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(gm, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (m *MemoryManager) GetUserMemory(groupID, userID int64) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	gm, err := m.load(groupID)
	if err != nil {
		return "", err
	}
	if item, ok := gm.UMS[fmt.Sprintf("%d", userID)]; ok {
		return item.Text, nil
	}
	return "", nil
}

func (m *MemoryManager) GetRecentSummaries(groupID int64, limit int) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	gm, err := m.load(groupID)
	if err != nil {
		return nil, err
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

func (m *MemoryManager) AddSummary(groupID int64, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	gm, _ := m.load(groupID)
	if gm.UMS == nil {
		gm.UMS = map[string]MemoryItem{}
	}
	gm.Summaries = append(gm.Summaries, SummaryItem{Content: content, Time: currentTimestamp()})
	if len(gm.Summaries) > 20 {
		gm.Summaries = gm.Summaries[len(gm.Summaries)-20:]
	}
	if err := m.save(groupID, gm); err != nil {
		return err
	}
	if vc := GetVectorClient(); vc != nil && vc.IsEnabled() {
		go func() { _ = vc.UpsertSummary(groupID, content) }()
	}
	return nil
}

func (m *MemoryManager) UpdateUserMemory(groupID, userID int64, text string) error {
	return m.UpdateUserMemoryWithName(groupID, userID, "", text)
}

func (m *MemoryManager) UpdateUserMemoryWithName(groupID, userID int64, userName, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	gm, _ := m.load(groupID)
	if gm.UMS == nil {
		gm.UMS = map[string]MemoryItem{}
	}
	gm.UMS[fmt.Sprintf("%d", userID)] = MemoryItem{Text: text}
	if err := m.save(groupID, gm); err != nil {
		return err
	}
	if vc := GetVectorClient(); vc != nil && vc.IsEnabled() {
		go func() { _ = vc.UpsertUserMemory(groupID, userID, userName, text) }()
	}
	return nil
}
