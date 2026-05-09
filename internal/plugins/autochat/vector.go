package autochat

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// MemoryVector 对应一条向量化记忆（user_memory 或 summary）。
type MemoryVector struct {
	ID        int64
	GroupID   int64
	UserID    int64
	UserName  string
	Type      string // user_memory | summary
	Text      string
	Timestamp int64
	Score     float64
}

// VectorClient 基于 SQLite + sqlite-vec(vec0) 的向量存储与检索。
//
// 主表 autochat_memories：保存原文/元数据
// 虚拟表 autochat_vec：存储 float[N] 向量，N 由首次创建时的 dimensions 决定
//
// 当 cfg.Vector.Dimensions 与已存在的虚拟表维度不一致时，会重建虚拟表
// 并清空向量（原文保留），并打印一条 warn 日志。
type VectorClient struct {
	db         *gorm.DB
	dimensions int
	topK       int
	enabled    bool
	mu         sync.Mutex
}

var (
	vectorClient *VectorClient
	vectorMu     sync.RWMutex
)

func initVectorClient(c *Config, db *gorm.DB) error {
	vectorMu.Lock()
	defer vectorMu.Unlock()
	if !c.Vector.Enabled || c.Vector.Dimensions <= 0 {
		vectorClient = &VectorClient{enabled: false}
		return nil
	}
	vc := &VectorClient{
		db:         db,
		dimensions: c.Vector.Dimensions,
		topK:       c.Vector.TopK,
		enabled:    true,
	}
	if err := vc.ensureSchema(); err != nil {
		return fmt.Errorf("vector: 初始化 schema 失败: %w", err)
	}
	vectorClient = vc
	log.Info().Int("dim", vc.dimensions).Msg("[autochat][vector] sqlite-vec 已初始化")
	return nil
}

func GetVectorClient() *VectorClient {
	vectorMu.RLock()
	defer vectorMu.RUnlock()
	return vectorClient
}

func (v *VectorClient) IsEnabled() bool { return v != nil && v.enabled }

// ensureSchema 创建 autochat_memories 主表与 autochat_vec 虚拟表（vec0）。
// vec0 维度固定，若已有的维度与 cfg 不一致则重建虚拟表。
func (v *VectorClient) ensureSchema() error {
	if err := v.db.Exec(`CREATE TABLE IF NOT EXISTS autochat_memories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		user_name TEXT,
		type TEXT NOT NULL,
		text TEXT NOT NULL,
		timestamp INTEGER NOT NULL
	)`).Error; err != nil {
		return err
	}
	if err := v.db.Exec(`CREATE INDEX IF NOT EXISTS idx_autochat_mem_g ON autochat_memories(group_id, type, timestamp DESC)`).Error; err != nil {
		return err
	}

	// 检查 vec 虚拟表是否存在以及维度是否匹配
	var existsCount int64
	v.db.Raw(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='autochat_vec'`).Scan(&existsCount)

	needRecreate := false
	if existsCount > 0 {
		// 探测当前维度：vec_length 不存在时通过 sql_create_statement 判断
		var sqlText string
		v.db.Raw(`SELECT sql FROM sqlite_master WHERE name='autochat_vec'`).Scan(&sqlText)
		expected := fmt.Sprintf("float[%d]", v.dimensions)
		if !contains(sqlText, expected) {
			log.Warn().Str("expected", expected).Str("sql", sqlText).Msg("[autochat][vector] 维度变化，重建 autochat_vec")
			needRecreate = true
		}
	} else {
		needRecreate = true
	}

	if needRecreate {
		if existsCount > 0 {
			if err := v.db.Exec(`DROP TABLE autochat_vec`).Error; err != nil {
				return err
			}
		}
		// vec0 是 sqlite-vec 提供的虚拟表
		ddl := fmt.Sprintf(`CREATE VIRTUAL TABLE autochat_vec USING vec0(
			id INTEGER PRIMARY KEY,
			embedding float[%d]
		)`, v.dimensions)
		if err := v.db.Exec(ddl).Error; err != nil {
			return err
		}
	}
	return nil
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && bytes.Contains([]byte(s), []byte(sub))
}

func serializeFloat32(vec []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, vec); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *VectorClient) generateEmbedding(text string) ([]float32, error) {
	ec := GetEmbeddingClient()
	if ec == nil || !ec.IsEnabled() {
		return nil, fmt.Errorf("embedding 未启用")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return ec.GetEmbedding(ctx, text)
}

// upsertWithVector 插入主记录 + 向量，自动获取自增 id。
func (v *VectorClient) upsertWithVector(groupID, userID int64, userName, typ, text string, timestamp int64, vec []float32) error {
	if len(vec) != v.dimensions {
		return fmt.Errorf("vector: 维度不匹配 expect=%d got=%d", v.dimensions, len(vec))
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	tx := v.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	res := tx.Exec(
		`INSERT INTO autochat_memories(group_id, user_id, user_name, type, text, timestamp) VALUES(?,?,?,?,?,?)`,
		groupID, userID, userName, typ, text, timestamp,
	)
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}
	var id int64
	tx.Raw(`SELECT last_insert_rowid()`).Scan(&id)
	blob, err := serializeFloat32(vec)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec(`INSERT INTO autochat_vec(id, embedding) VALUES(?, ?)`, id, blob).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (v *VectorClient) UpsertUserMemory(groupID, userID int64, userName, text string) error {
	if !v.IsEnabled() {
		return nil
	}
	emb, err := v.generateEmbedding(text)
	if err != nil {
		return err
	}
	return v.upsertWithVector(groupID, userID, userName, "user_memory", text, time.Now().Unix(), emb)
}

func (v *VectorClient) UpsertSummary(groupID int64, content string) error {
	if !v.IsEnabled() {
		return nil
	}
	emb, err := v.generateEmbedding(content)
	if err != nil {
		return err
	}
	return v.upsertWithVector(groupID, 0, "", "summary", content, time.Now().Unix(), emb)
}

// queryByVector 用 sqlite-vec KNN 检索 + 按 type/group 过滤。
func (v *VectorClient) queryByVector(groupID int64, typ string, queryVec []float32, topK int) ([]MemoryVector, error) {
	if topK <= 0 {
		topK = v.topK
	}
	if topK <= 0 {
		topK = 5
	}
	blob, err := serializeFloat32(queryVec)
	if err != nil {
		return nil, err
	}
	type row struct {
		ID        int64
		GroupID   int64
		UserID    int64
		UserName  string
		Type      string
		Text      string
		Timestamp int64
		Distance  float64
	}
	// 先用 vec MATCH 召回较多，然后按主表过滤；vec0 不支持 join 中的非 KNN 过滤，
	// 因此此处采用子查询：拿足够多 KNN，再按 group/type 过滤、按距离排序。
	limit := topK * 4
	if limit < 20 {
		limit = 20
	}
	var rows []row
	q := `
WITH knn AS (
  SELECT id, distance FROM autochat_vec
  WHERE embedding MATCH ? AND k = ?
)
SELECT m.id, m.group_id, m.user_id, m.user_name, m.type, m.text, m.timestamp, knn.distance
FROM knn JOIN autochat_memories m ON m.id = knn.id
WHERE m.group_id = ? AND m.type = ?
ORDER BY knn.distance ASC
LIMIT ?
`
	if err := v.db.Raw(q, blob, limit, groupID, typ, topK).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]MemoryVector, 0, len(rows))
	for _, r := range rows {
		out = append(out, MemoryVector{
			ID:        r.ID,
			GroupID:   r.GroupID,
			UserID:    r.UserID,
			UserName:  r.UserName,
			Type:      r.Type,
			Text:      r.Text,
			Timestamp: r.Timestamp,
			Score:     1 - r.Distance, // 余弦相似度近似
		})
	}
	return out, nil
}

func (v *VectorClient) QueryRelevantSummaries(groupID int64, queryText string, topK int) ([]MemoryVector, error) {
	if !v.IsEnabled() {
		return nil, nil
	}
	emb, err := v.generateEmbedding(queryText)
	if err != nil {
		return nil, err
	}
	return v.queryByVector(groupID, "summary", emb, topK)
}

func (v *VectorClient) QueryMemoriesByKeyword(groupID int64, keyword string, topK int) ([]MemoryVector, error) {
	if !v.IsEnabled() {
		return nil, nil
	}
	emb, err := v.generateEmbedding(keyword)
	if err != nil {
		return nil, err
	}
	return v.queryByVector(groupID, "user_memory", emb, topK)
}

// QueryUserMemories 列出指定用户最近 topK 条记忆（按时间倒序）。
func (v *VectorClient) QueryUserMemories(groupID, userID int64, topK int) ([]MemoryVector, error) {
	if !v.IsEnabled() {
		return nil, nil
	}
	if topK <= 0 {
		topK = 10
	}
	type row struct {
		ID        int64
		GroupID   int64
		UserID    int64
		UserName  string
		Type      string
		Text      string
		Timestamp int64
	}
	var rows []row
	if err := v.db.Raw(
		`SELECT id, group_id, user_id, user_name, type, text, timestamp FROM autochat_memories
		 WHERE group_id=? AND user_id=? AND type='user_memory' ORDER BY timestamp DESC LIMIT ?`,
		groupID, userID, topK,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]MemoryVector, 0, len(rows))
	for _, r := range rows {
		out = append(out, MemoryVector{
			ID: r.ID, GroupID: r.GroupID, UserID: r.UserID, UserName: r.UserName,
			Type: r.Type, Text: r.Text, Timestamp: r.Timestamp,
		})
	}
	return out, nil
}

// QueryRecentMemories 群组内最近 topK 条 user_memory（不做语义检索）。
func (v *VectorClient) QueryRecentMemories(groupID int64, topK int) ([]MemoryVector, error) {
	if !v.IsEnabled() {
		return nil, nil
	}
	if topK <= 0 {
		topK = 5
	}
	type row struct {
		ID        int64
		GroupID   int64
		UserID    int64
		UserName  string
		Type      string
		Text      string
		Timestamp int64
	}
	var rows []row
	if err := v.db.Raw(
		`SELECT id, group_id, user_id, user_name, type, text, timestamp FROM autochat_memories
		 WHERE group_id=? AND type='user_memory' ORDER BY timestamp DESC LIMIT ?`,
		groupID, topK,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]MemoryVector, 0, len(rows))
	for _, r := range rows {
		out = append(out, MemoryVector{
			ID: r.ID, GroupID: r.GroupID, UserID: r.UserID, UserName: r.UserName,
			Type: r.Type, Text: r.Text, Timestamp: r.Timestamp,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Timestamp > out[j].Timestamp })
	return out, nil
}
