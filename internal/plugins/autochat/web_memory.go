package autochat

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// memoryItemDTO 单条记忆条目。
type memoryItemDTO struct {
	ID        int64   `json:"id"`
	GroupID   int64   `json:"group_id"`
	UserID    int64   `json:"user_id"`
	UserName  string  `json:"user_name,omitempty"`
	Type      string  `json:"type"`
	Text      string  `json:"text"`
	Timestamp int64   `json:"timestamp"`
	Score     float64 `json:"score,omitempty"`
}

// handleListMemoryGroups 返回向量库中出现过的 group_id 列表 + 每群条数。
func (p *pluginImpl) handleListMemoryGroups(c *fiber.Ctx) error {
	vc := GetVectorClient()
	if vc == nil || !vc.IsEnabled() {
		return c.JSON(fiber.Map{"groups": []any{}, "vector_enabled": false})
	}
	type row struct {
		GroupID int64
		Count   int64
	}
	var rows []row
	if err := vc.db.Raw(
		`SELECT group_id AS group_id, COUNT(*) AS count FROM autochat_memories GROUP BY group_id ORDER BY group_id`,
	).Scan(&rows).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out := make([]fiber.Map, 0, len(rows))
	for _, r := range rows {
		out = append(out, fiber.Map{"group_id": r.GroupID, "count": r.Count})
	}
	return c.JSON(fiber.Map{"groups": out, "vector_enabled": true})
}

// handleQueryMemoryItems 检索记忆。
//
//	q     非空 -> 走 embedding+sqlite-vec KNN（要求 type 已指定，user_memory 或 summary）
//	q     为空 -> 按 group/user/type 过滤后按 timestamp DESC 列出
//	limit 缺省 20，最大 100
func (p *pluginImpl) handleQueryMemoryItems(c *fiber.Ctx) error {
	vc := GetVectorClient()
	if vc == nil || !vc.IsEnabled() {
		return c.JSON(fiber.Map{"items": []any{}, "total": 0, "vector_enabled": false})
	}

	q := strings.TrimSpace(c.Query("q"))
	gid, _ := strconv.ParseInt(c.Query("group_id"), 10, 64)
	uid, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	typ := strings.TrimSpace(c.Query("type"))
	if typ != "" && typ != "user_memory" && typ != "summary" {
		typ = ""
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if q != "" {
		// 语义检索：必须指定 group_id 和 type，否则返回 400
		if gid == 0 || typ == "" {
			return fiber.NewError(fiber.StatusBadRequest, "semantic search requires group_id and type")
		}
		mems, err := p.queryMemoryByVector(gid, typ, q, limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		out := make([]memoryItemDTO, 0, len(mems))
		for _, m := range mems {
			if uid != 0 && m.UserID != uid {
				continue
			}
			out = append(out, memoryItemDTO{
				ID: m.ID, GroupID: m.GroupID, UserID: m.UserID, UserName: m.UserName,
				Type: m.Type, Text: m.Text, Timestamp: m.Timestamp, Score: m.Score,
			})
		}
		return c.JSON(fiber.Map{"items": out, "total": len(out), "vector_enabled": true, "mode": "semantic"})
	}

	// 时间序列模式：直接 SQL
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
	conds := []string{"1=1"}
	args := []any{}
	if gid != 0 {
		conds = append(conds, "group_id=?")
		args = append(args, gid)
	}
	if uid != 0 {
		conds = append(conds, "user_id=?")
		args = append(args, uid)
	}
	if typ != "" {
		conds = append(conds, "type=?")
		args = append(args, typ)
	}
	args = append(args, limit)
	sql := `SELECT id, group_id, user_id, user_name, type, text, timestamp
	        FROM autochat_memories
	        WHERE ` + strings.Join(conds, " AND ") + ` ORDER BY timestamp DESC LIMIT ?`
	if err := vc.db.Raw(sql, args...).Scan(&rows).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out := make([]memoryItemDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, memoryItemDTO{
			ID: r.ID, GroupID: r.GroupID, UserID: r.UserID, UserName: r.UserName,
			Type: r.Type, Text: r.Text, Timestamp: r.Timestamp,
		})
	}
	return c.JSON(fiber.Map{"items": out, "total": len(out), "vector_enabled": true, "mode": "recent"})
}

// queryMemoryByVector 是 VectorClient.queryByVector 的薄包装（包内私有，
// 直接调用 GetEmbedding + queryByVector）。
func (p *pluginImpl) queryMemoryByVector(groupID int64, typ, queryText string, topK int) ([]MemoryVector, error) {
	vc := GetVectorClient()
	if vc == nil || !vc.IsEnabled() {
		return nil, nil
	}
	emb, err := vc.generateEmbedding(queryText)
	if err != nil {
		return nil, err
	}
	return vc.queryByVector(groupID, typ, emb, topK)
}

// handleDeleteMemoryItem 删除单条记忆（同时清理 vec 表）。
func (p *pluginImpl) handleDeleteMemoryItem(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	vc := GetVectorClient()
	if vc == nil || !vc.IsEnabled() {
		return fiber.NewError(fiber.StatusServiceUnavailable, "vector client not available")
	}
	tx := vc.db.Begin()
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, tx.Error.Error())
	}
	if err := tx.Exec(`DELETE FROM autochat_memories WHERE id=?`, id).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if err := tx.Exec(`DELETE FROM autochat_vec WHERE id=?`, id).Error; err != nil {
		tx.Rollback()
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if err := tx.Commit().Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "id": id})
}
