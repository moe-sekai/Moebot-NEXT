package autochat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// FileDB 一个简单的 JSON 持久化 KV，用于白名单 / 冷却 last-use / token 统计。
type FileDB struct {
	path string
	data map[string]any
	mu   sync.RWMutex
}

func NewFileDB(path string) *FileDB {
	db := &FileDB{path: path, data: map[string]any{}}
	db.load()
	return db
}

func (db *FileDB) load() {
	data, err := os.ReadFile(db.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &db.data)
}

func (db *FileDB) save() error {
	dir := filepath.Dir(db.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(db.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(db.path, data, 0o644)
}

func (db *FileDB) Get(key string) any {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.data[key]
}

func (db *FileDB) GetString(key string) string {
	if s, ok := db.Get(key).(string); ok {
		return s
	}
	return ""
}

func (db *FileDB) GetStringSlice(key string) []string {
	v := db.Get(key)
	if v == nil {
		return nil
	}
	if arr, ok := v.([]string); ok {
		return arr
	}
	if arr, ok := v.([]any); ok {
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func (db *FileDB) GetMap(key string) map[string]any {
	if m, ok := db.Get(key).(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func (db *FileDB) Set(key string, value any) error {
	db.mu.Lock()
	db.data[key] = value
	err := db.save()
	db.mu.Unlock()
	return err
}

func (db *FileDB) Delete(key string) {
	db.mu.Lock()
	delete(db.data, key)
	_ = db.save()
	db.mu.Unlock()
}

// GroupWhiteList 群组白名单。
type GroupWhiteList struct {
	db  *FileDB
	key string
}

func NewGroupWhiteList(db *FileDB, key string) *GroupWhiteList {
	return &GroupWhiteList{db: db, key: key}
}

func (wl *GroupWhiteList) listKey() string { return wl.key + "_whitelist" }

func (wl *GroupWhiteList) Check(groupID int64) bool {
	gidStr := formatInt64(groupID)
	for _, id := range wl.db.GetStringSlice(wl.listKey()) {
		if id == gidStr {
			return true
		}
	}
	return false
}

func (wl *GroupWhiteList) Add(groupID int64) error {
	list := wl.db.GetStringSlice(wl.listKey())
	gidStr := formatInt64(groupID)
	for _, id := range list {
		if id == gidStr {
			return nil
		}
	}
	list = append(list, gidStr)
	return wl.db.Set(wl.listKey(), list)
}

func (wl *GroupWhiteList) Remove(groupID int64) error {
	list := wl.db.GetStringSlice(wl.listKey())
	gidStr := formatInt64(groupID)
	out := make([]string, 0, len(list))
	for _, id := range list {
		if id != gidStr {
			out = append(out, id)
		}
	}
	return wl.db.Set(wl.listKey(), out)
}

// ColdDown 冷却控制（内存态，不持久化每次调用，但 lastUse 在进程内共享）。
type ColdDown struct {
	duration int
	lastUse  map[string]int64
	mu       sync.RWMutex
}

func NewColdDown(seconds int) *ColdDown {
	return &ColdDown{duration: seconds, lastUse: map[string]int64{}}
}

func (cd *ColdDown) Check(id string) bool {
	cd.mu.RLock()
	last, ok := cd.lastUse[id]
	cd.mu.RUnlock()
	now := currentTimestamp()
	if !ok || now-last >= int64(cd.duration) {
		cd.mu.Lock()
		cd.lastUse[id] = now
		cd.mu.Unlock()
		return true
	}
	return false
}

func (cd *ColdDown) Remaining(id string) int64 {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	last, ok := cd.lastUse[id]
	if !ok {
		return 0
	}
	r := int64(cd.duration) - (currentTimestamp() - last)
	if r < 0 {
		return 0
	}
	return r
}

// TokenStats 按日累计 token 用量。
type TokenStats struct {
	db *FileDB
	mu sync.Mutex
}

func NewTokenStats(db *FileDB) *TokenStats {
	return &TokenStats{db: db}
}

func (ts *TokenStats) Record(prompt, completion int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	today := currentDate()
	stats := ts.db.GetMap("token_stats")
	if stats == nil {
		stats = map[string]any{}
	}
	cur, _ := stats[today].(map[string]any)
	if cur == nil {
		cur = map[string]any{"prompt_tokens": 0.0, "completion_tokens": 0.0, "request_count": 0.0}
	}
	pt, _ := cur["prompt_tokens"].(float64)
	ct, _ := cur["completion_tokens"].(float64)
	rc, _ := cur["request_count"].(float64)
	cur["prompt_tokens"] = pt + float64(prompt)
	cur["completion_tokens"] = ct + float64(completion)
	cur["request_count"] = rc + 1
	stats[today] = cur
	if err := ts.db.Set("token_stats", stats); err != nil {
		log.Warn().Err(err).Msg("[autochat] 写入 token_stats 失败")
	}
}

func (ts *TokenStats) GetStats(days int) (prompt, completion, requests int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	stats := ts.db.GetMap("token_stats")
	now := currentTimestamp()
	for i := 0; i < days; i++ {
		date := formatDate(now - int64(i*86400))
		m, _ := stats[date].(map[string]any)
		if m == nil {
			continue
		}
		pt, _ := m["prompt_tokens"].(float64)
		ct, _ := m["completion_tokens"].(float64)
		rc, _ := m["request_count"].(float64)
		prompt += int(pt)
		completion += int(ct)
		requests += int(rc)
	}
	return
}

// BufferMessage / MessageBuffer 用于群上下文滚动窗口。
type BufferMessage struct {
	SenderName string
	SenderID   int64
	Content    string
	Time       int64
	GroupID    int64
	ImageDescs []string
}

type MessageBuffer struct {
	mu    sync.RWMutex
	data  map[int64][]BufferMessage
	limit int
}

func NewMessageBuffer(limit int) *MessageBuffer {
	if limit <= 0 {
		limit = 20
	}
	return &MessageBuffer{data: map[int64][]BufferMessage{}, limit: limit}
}

func (mb *MessageBuffer) Add(groupID int64, senderName string, senderID int64, msg string, msgTime int64) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	q := mb.data[groupID]
	q = append(q, BufferMessage{SenderName: senderName, SenderID: senderID, Content: msg, Time: msgTime, GroupID: groupID})
	if len(q) > mb.limit {
		q = q[len(q)-mb.limit:]
	}
	mb.data[groupID] = q
}

func (mb *MessageBuffer) UpdateImageDescs(groupID, senderID, msgTime int64, descs []string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	q, ok := mb.data[groupID]
	if !ok {
		return
	}
	for i := len(q) - 1; i >= 0; i-- {
		if q[i].SenderID == senderID && q[i].Time == msgTime {
			q[i].ImageDescs = descs
			mb.data[groupID] = q
			return
		}
	}
}

// Get 返回最近 limit 条消息的格式化字符串列表。
func (mb *MessageBuffer) Get(groupID int64, limit int) []string {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	q, ok := mb.data[groupID]
	if !ok {
		return []string{}
	}
	if limit > len(q) {
		limit = len(q)
	}
	loc := time.FixedZone("CST", 8*3600)
	now := time.Now().Unix()
	out := make([]string, limit)
	start := len(q) - limit
	for i := 0; i < limit; i++ {
		m := q[start+i]
		content := m.Content
		if len([]rune(content)) > 150 {
			content = string([]rune(content)[:150]) + "..."
		}
		if len(m.ImageDescs) > 0 {
			content += "\n[图片: " + strings.Join(m.ImageDescs, " / ") + "]"
		}
		t := time.Unix(m.Time, 0).In(loc)
		var rel string
		diff := now - m.Time
		switch {
		case diff < 60:
			rel = "刚刚"
		case diff < 3600:
			rel = fmt.Sprintf("%d分钟前", diff/60)
		case diff < 86400:
			rel = fmt.Sprintf("%d小时前", diff/3600)
		default:
			rel = fmt.Sprintf("%d天前", diff/86400)
		}
		content = strings.ReplaceAll(content, "\n", "\\n")
		out[i] = fmt.Sprintf("%s (%s) [%d] %s(%d):\n%s",
			t.Format("2006-01-02 15:04:05"), rel, m.GroupID, m.SenderName, m.SenderID, content)
	}
	return out
}

func (mb *MessageBuffer) GetContext(groupID int64, limit int) []BufferMessage {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	q, ok := mb.data[groupID]
	if !ok {
		return []BufferMessage{}
	}
	if limit > len(q) {
		limit = len(q)
	}
	out := make([]BufferMessage, limit)
	copy(out, q[len(q)-limit:])
	return out
}
