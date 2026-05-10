package gallery

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type GalleryManager struct {
	db  *gorm.DB
	cfg *Config
}

func NewGalleryManager(db *gorm.DB, cfg *Config) *GalleryManager {
	return &GalleryManager{db: db, cfg: cfg}
}

// --- 画廊 CRUD ---

func (m *GalleryManager) ListGalleries() ([]GalleryInfo, error) {
	var galleries []GalleryInfo
	return galleries, m.db.Order("name").Find(&galleries).Error
}

func (m *GalleryManager) FindGallery(nameOrAlias string) (*GalleryInfo, error) {
	var g GalleryInfo
	if err := m.db.Where("name = ?", nameOrAlias).First(&g).Error; err == nil {
		return &g, nil
	}
	// 搜索别名
	var all []GalleryInfo
	if err := m.db.Find(&all).Error; err != nil {
		return nil, err
	}
	for _, g := range all {
		aliases := parseAliases(g.Aliases)
		for _, a := range aliases {
			if a == nameOrAlias {
				return &g, nil
			}
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *GalleryManager) CreateGallery(name string) error {
	if !isValidName(name) {
		return fmt.Errorf("画廊名称\"%s\"无效", name)
	}
	if _, err := m.FindGallery(name); err == nil {
		return fmt.Errorf("画廊\"%s\"已存在", name)
	}
	picsDir := filepath.Join(m.cfg.DataDir, name)
	if err := os.MkdirAll(picsDir, 0o755); err != nil {
		return err
	}
	g := GalleryInfo{
		Name:    name,
		Aliases: "[]",
		Mode:    "edit",
		PicsDir: picsDir,
	}
	return m.db.Create(&g).Error
}

func (m *GalleryManager) DeleteGallery(nameOrAlias string) error {
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}
	m.db.Where("gall_name = ?", g.Name).Delete(&GalleryPic{})
	return m.db.Delete(g).Error
}

func (m *GalleryManager) SetMode(nameOrAlias, mode string) (string, string, error) {
	if !isValidMode(mode) {
		return "", "", fmt.Errorf("无效模式\"%s\"", mode)
	}
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return "", "", fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}
	old := g.Mode
	g.Mode = mode
	return old, mode, m.db.Save(g).Error
}

// EffectiveMode 返回某画廊在指定群中的实际生效模式。
// 优先 GroupModes[gid]，缺省回落到 g.Mode。私聊或 gid==0 始终用 g.Mode。
func (m *GalleryManager) EffectiveMode(g *GalleryInfo, groupID int64) string {
	if g == nil {
		return ""
	}
	if groupID > 0 {
		gm := parseGroupModes(g.GroupModes)
		if v, ok := gm[groupID]; ok && v != "" {
			return v
		}
	}
	return g.Mode
}

// SetGroupMode 设置某画廊在某群中的覆盖模式。mode 为空字符串时等价于 UnsetGroupMode。
func (m *GalleryManager) SetGroupMode(nameOrAlias string, groupID int64, mode string) error {
	if groupID <= 0 {
		return errors.New("group id must be positive")
	}
	if mode != "" && !isValidMode(mode) {
		return fmt.Errorf("无效模式\"%s\"", mode)
	}
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}
	gm := parseGroupModes(g.GroupModes)
	if mode == "" {
		delete(gm, groupID)
	} else {
		gm[groupID] = mode
	}
	g.GroupModes = marshalGroupModes(gm)
	return m.db.Save(g).Error
}

// UnsetGroupMode 移除某画廊在某群的覆盖（回落到全局 Mode）。
func (m *GalleryManager) UnsetGroupMode(nameOrAlias string, groupID int64) error {
	return m.SetGroupMode(nameOrAlias, groupID, "")
}

// ReplaceGroupModes 整体替换 GroupModes（供 Web 表单一次保存全部覆盖）。
func (m *GalleryManager) ReplaceGroupModes(nameOrAlias string, modes map[int64]string) error {
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}
	clean := map[int64]string{}
	for gid, mode := range modes {
		if gid <= 0 || mode == "" {
			continue
		}
		if !isValidMode(mode) {
			return fmt.Errorf("群 %d 的模式\"%s\"无效", gid, mode)
		}
		clean[gid] = mode
	}
	g.GroupModes = marshalGroupModes(clean)
	return m.db.Save(g).Error
}

func isValidMode(mode string) bool {
	switch mode {
	case "edit", "view", "off":
		return true
	}
	return false
}

func (m *GalleryManager) AddAlias(nameOrAlias, alias string) error {
	if !isValidName(alias) {
		return fmt.Errorf("别名\"%s\"无效", alias)
	}
	if _, err := m.FindGallery(alias); err == nil {
		return fmt.Errorf("别名\"%s\"已被占用", alias)
	}
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}
	aliases := parseAliases(g.Aliases)
	aliases = append(aliases, alias)
	g.Aliases = marshalAliases(aliases)
	return m.db.Save(g).Error
}

func (m *GalleryManager) DelAlias(nameOrAlias, alias string) error {
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}
	aliases := parseAliases(g.Aliases)
	found := false
	var newAliases []string
	for _, a := range aliases {
		if a == alias {
			found = true
		} else {
			newAliases = append(newAliases, a)
		}
	}
	if !found {
		return fmt.Errorf("别名\"%s\"不存在", alias)
	}
	g.Aliases = marshalAliases(newAliases)
	return m.db.Save(g).Error
}

func (m *GalleryManager) SetCover(nameOrAlias string, pid uint) error {
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}
	var pic GalleryPic
	if err := m.db.Where("pid = ? AND gall_name = ?", pid, g.Name).First(&pic).Error; err != nil {
		return fmt.Errorf("图片pid=%d不属于画廊\"%s\"", pid, g.Name)
	}
	g.CoverPID = pid
	return m.db.Save(g).Error
}

// --- 图片操作 ---

func (m *GalleryManager) PicCount(gallName string) int64 {
	var count int64
	m.db.Model(&GalleryPic{}).Where("gall_name = ?", gallName).Count(&count)
	return count
}

func (m *GalleryManager) ListPics(gallName string, offset, limit int) ([]GalleryPic, error) {
	var pics []GalleryPic
	q := m.db.Where("gall_name = ?", gallName).Order("pid")
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	return pics, q.Find(&pics).Error
}

func (m *GalleryManager) FindPic(pid uint) (*GalleryPic, error) {
	var pic GalleryPic
	return &pic, m.db.Where("pid = ?", pid).First(&pic).Error
}

func (m *GalleryManager) RandomPics(gallName string, n int) ([]GalleryPic, error) {
	var all []GalleryPic
	if err := m.db.Where("gall_name = ?", gallName).Find(&all).Error; err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, nil
	}
	pics := make([]GalleryPic, 0, n)
	for i := 0; i < n; i++ {
		pics = append(pics, all[rand.Intn(len(all))])
	}
	return pics, nil
}

// FindPicByNegativeIndex 支持 -1 表示倒数第一张等。
func (m *GalleryManager) FindPicByNegativeIndex(gallName string, negIdx int) (*GalleryPic, error) {
	var pics []GalleryPic
	if err := m.db.Where("gall_name = ?", gallName).Order("pid").Find(&pics).Error; err != nil {
		return nil, err
	}
	idx := len(pics) + negIdx
	if idx < 0 || idx >= len(pics) {
		return nil, fmt.Errorf("画廊仅有%d张图片", len(pics))
	}
	return &pics[idx], nil
}

func (m *GalleryManager) AddPic(gallName, srcPath string, checkDup bool) (uint, error) {
	g, err := m.FindGallery(gallName)
	if err != nil {
		return 0, fmt.Errorf("画廊\"%s\"不存在", gallName)
	}
	hash1, hash2, err := calcHashes(srcPath)
	if err != nil {
		return 0, fmt.Errorf("计算图片哈希失败: %w", err)
	}

	if checkDup {
		if dupPID, err := m.checkDuplicate(g.Name, hash1, hash2, 0); err == nil {
			return 0, fmt.Errorf("画廊中已存在相似图片(pid=%d)", dupPID)
		}
	}

	ext := filepath.Ext(srcPath)
	timeStr := time.Now().Format("2006-01-02_15-04-05")
	if err := os.MkdirAll(g.PicsDir, 0o755); err != nil {
		return 0, err
	}

	pic := GalleryPic{
		GallName: g.Name,
		Hash1:    hash1,
		Hash2:    hash2,
	}
	if err := m.db.Create(&pic).Error; err != nil {
		return 0, err
	}

	dstPath := filepath.Join(g.PicsDir, fmt.Sprintf("%s_%d%s", timeStr, pic.PID, ext))
	data, err := os.ReadFile(srcPath)
	if err != nil {
		m.db.Delete(&pic)
		return 0, err
	}
	if err := os.WriteFile(dstPath, data, 0o644); err != nil {
		m.db.Delete(&pic)
		return 0, err
	}

	pic.Path = dstPath
	thumbPath, thumbErr := ensureThumb(dstPath)
	if thumbErr == nil {
		pic.ThumbPath = thumbPath
	}
	m.db.Save(&pic)

	return pic.PID, nil
}

func (m *GalleryManager) DelPic(pid uint) error {
	var pic GalleryPic
	if err := m.db.Where("pid = ?", pid).First(&pic).Error; err != nil {
		return fmt.Errorf("图片ID %d 不存在", pid)
	}
	removeFileQuiet(pic.Path)
	removeFileQuiet(pic.ThumbPath)
	return m.db.Delete(&pic).Error
}

func (m *GalleryManager) ReplacePic(pid uint, srcPath string, checkDup bool) error {
	var pic GalleryPic
	if err := m.db.Where("pid = ?", pid).First(&pic).Error; err != nil {
		return fmt.Errorf("图片ID %d 不存在", pid)
	}

	hash1, hash2, err := calcHashes(srcPath)
	if err != nil {
		return fmt.Errorf("计算图片哈希失败: %w", err)
	}

	if checkDup {
		if dupPID, err := m.checkDuplicate(pic.GallName, hash1, hash2, pid); err == nil {
			return fmt.Errorf("画廊中已存在相似图片(pid=%d)", dupPID)
		}
	}

	removeFileQuiet(pic.Path)
	removeFileQuiet(pic.ThumbPath)

	ext := filepath.Ext(srcPath)
	g, _ := m.FindGallery(pic.GallName)
	dir := m.cfg.DataDir
	if g != nil {
		dir = g.PicsDir
	}
	timeStr := time.Now().Format("2006-01-02_15-04-05")
	dstPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", timeStr, pid, ext))
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dstPath, data, 0o644); err != nil {
		return err
	}

	pic.Path = dstPath
	pic.Hash1 = hash1
	pic.Hash2 = hash2
	pic.ThumbPath = ""
	if thumbPath, err := ensureThumb(dstPath); err == nil {
		pic.ThumbPath = thumbPath
	}
	return m.db.Save(&pic).Error
}

func (m *GalleryManager) checkDuplicate(gallName, hash1, hash2 string, excludePID uint) (uint, error) {
	var pics []GalleryPic
	m.db.Where("gall_name = ?", gallName).Find(&pics)
	cfg := m.cfg
	for _, p := range pics {
		if p.PID == excludePID {
			continue
		}
		if p.Hash1 == "" || p.Hash2 == "" {
			continue
		}
		if isSame(hash1, hash2, p.Hash1, p.Hash2, cfg) {
			return p.PID, nil
		}
	}
	return 0, errors.New("no duplicate")
}

func (m *GalleryManager) ReloadGallery(nameOrAlias string) (newPIDs []uint, delPIDs []uint, err error) {
	g, err := m.FindGallery(nameOrAlias)
	if err != nil {
		return nil, nil, fmt.Errorf("画廊\"%s\"不存在", nameOrAlias)
	}

	entries, err := os.ReadDir(g.PicsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var existingPics []GalleryPic
	m.db.Where("gall_name = ?", g.Name).Find(&existingPics)

	existingPaths := map[string]struct{}{}
	for _, p := range existingPics {
		abs, _ := filepath.Abs(p.Path)
		existingPaths[abs] = struct{}{}
	}

	picExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	for _, entry := range entries {
		if entry.IsDir() || strings.Contains(entry.Name(), "_thumb") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !picExts[ext] {
			continue
		}
		absPath, _ := filepath.Abs(filepath.Join(g.PicsDir, entry.Name()))
		if _, ok := existingPaths[absPath]; ok {
			continue
		}
		hash1, hash2, hErr := calcHashes(absPath)
		if hErr != nil {
			log.Warn().Err(hErr).Str("path", absPath).Msg("[gallery] 计算哈希失败，跳过")
			continue
		}
		pic := GalleryPic{GallName: g.Name, Path: absPath, Hash1: hash1, Hash2: hash2}
		if err := m.db.Create(&pic).Error; err != nil {
			continue
		}
		if thumbPath, tErr := ensureThumb(absPath); tErr == nil {
			pic.ThumbPath = thumbPath
			m.db.Save(&pic)
		}
		newPIDs = append(newPIDs, pic.PID)
	}

	for _, pic := range existingPics {
		if _, err := os.Stat(pic.Path); os.IsNotExist(err) {
			m.db.Delete(&pic)
			delPIDs = append(delPIDs, pic.PID)
		}
	}
	return
}

func (m *GalleryManager) CheckDuplicates(gallName string, rehash bool) (map[uint][]uint, error) {
	g, err := m.FindGallery(gallName)
	if err != nil {
		return nil, fmt.Errorf("画廊\"%s\"不存在", gallName)
	}
	var pics []GalleryPic
	m.db.Where("gall_name = ?", g.Name).Find(&pics)

	if rehash {
		for i := range pics {
			h1, h2, hErr := calcHashes(pics[i].Path)
			if hErr != nil {
				continue
			}
			pics[i].Hash1 = h1
			pics[i].Hash2 = h2
			m.db.Save(&pics[i])
		}
	}

	type group struct {
		first *GalleryPic
		dups  []uint
	}
	groups := make(map[uint]*group)
	cfg := m.cfg

	for i := range pics {
		pic := &pics[i]
		if pic.Hash1 == "" || pic.Hash2 == "" {
			continue
		}
		var matchedPID uint
		for pid, g := range groups {
			if isSame(pic.Hash1, pic.Hash2, g.first.Hash1, g.first.Hash2, cfg) {
				matchedPID = pid
				break
			}
		}
		if matchedPID != 0 {
			groups[matchedPID].dups = append(groups[matchedPID].dups, pic.PID)
		} else {
			groups[pic.PID] = &group{first: pic}
		}
	}

	result := map[uint][]uint{}
	for pid, g := range groups {
		if len(g.dups) > 0 {
			result[pid] = g.dups
		}
	}
	return result, nil
}

// --- 上传记录 ---

func (m *GalleryManager) AddUploadRecord(userID, groupID int64, gallName string, pids []uint) (uint, error) {
	data, _ := json.Marshal(pids)
	r := GalleryUploadRecord{
		UserID:    userID,
		GroupID:   groupID,
		GallName:  gallName,
		PIDs:      string(data),
		CreatedAt: time.Now(),
	}
	if err := m.db.Create(&r).Error; err != nil {
		return 0, err
	}
	return r.ID, nil
}

// UploadRecordFilter 用于 Web/管理页面的筛选条件。空字段表示不过滤。
type UploadRecordFilter struct {
	UserID   int64
	GroupID  int64
	GallName string
	Offset   int
	Limit    int
}

// ListUploadRecords 返回符合筛选条件的上传记录及总数（按 ID 倒序）。
func (m *GalleryManager) ListUploadRecords(f UploadRecordFilter) ([]GalleryUploadRecord, int64, error) {
	q := m.db.Model(&GalleryUploadRecord{})
	if f.UserID != 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.GroupID != 0 {
		q = q.Where("group_id = ?", f.GroupID)
	}
	if f.GallName != "" {
		q = q.Where("gall_name = ?", f.GallName)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var rows []GalleryUploadRecord
	if err := q.Order("id DESC").Offset(f.Offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, total, err
	}
	return rows, total, nil
}

func (m *GalleryManager) GetUploadRecord(hid uint) (*GalleryUploadRecord, error) {
	var r GalleryUploadRecord
	return &r, m.db.Where("id = ?", hid).First(&r).Error
}

func (m *GalleryManager) RevertUploadRecord(hid uint) ([]uint, []uint, error) {
	r, err := m.GetUploadRecord(hid)
	if err != nil {
		return nil, nil, fmt.Errorf("上传#%d不存在", hid)
	}
	if r.Reverted {
		return nil, nil, fmt.Errorf("上传#%d已被撤销", hid)
	}
	var pids []uint
	json.Unmarshal([]byte(r.PIDs), &pids)

	var okList, errList []uint
	for _, pid := range pids {
		if err := m.DelPic(pid); err != nil {
			errList = append(errList, pid)
		} else {
			okList = append(okList, pid)
		}
	}
	r.Reverted = true
	m.db.Save(r)
	return okList, errList, nil
}

func (m *GalleryManager) RevertUserLastUpload(userID int64) (*GalleryUploadRecord, []uint, []uint, error) {
	var records []GalleryUploadRecord
	m.db.Where("user_id = ? AND reverted = ?", userID, false).Order("id DESC").Find(&records)
	if len(records) == 0 {
		return nil, nil, nil, errors.New("你没有可撤销的上传记录")
	}
	r := &records[0]
	expiredHours := m.cfg.RevertExpiredHours
	if time.Since(r.CreatedAt) > time.Duration(expiredHours)*time.Hour {
		return nil, nil, nil, fmt.Errorf("最近一次上传记录已超过%d小时，无法撤销", expiredHours)
	}

	var pids []uint
	json.Unmarshal([]byte(r.PIDs), &pids)

	var okList, errList []uint
	for _, pid := range pids {
		if err := m.DelPic(pid); err != nil {
			errList = append(errList, pid)
		} else {
			okList = append(okList, pid)
		}
	}
	r.Reverted = true
	m.db.Save(r)
	return r, okList, errList, nil
}

// --- 工具函数 ---

func isValidName(name string) bool {
	if name == "" || len(name) > 32 {
		return false
	}
	for _, c := range `\/:*?"<>| ` {
		if strings.ContainsRune(name, c) {
			return false
		}
	}
	// 纯数字不允许作为名称
	if _, err := fmt.Sscanf(name, "%d", new(int)); err == nil {
		return false
	}
	return true
}

func parseAliases(s string) []string {
	if s == "" || s == "[]" {
		return nil
	}
	var aliases []string
	json.Unmarshal([]byte(s), &aliases)
	return aliases
}

func marshalAliases(aliases []string) string {
	if len(aliases) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(aliases)
	return string(data)
}

// parseGroupModes 把 JSON {"123":"edit"} 形式的字符串解析回 map[int64]string。
// 健壮：空串/解析失败/非法 key 都返回空 map。
func parseGroupModes(s string) map[int64]string {
	out := map[int64]string{}
	if s == "" || s == "{}" {
		return out
	}
	raw := map[string]string{}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return out
	}
	for k, v := range raw {
		var gid int64
		if _, err := fmt.Sscanf(k, "%d", &gid); err != nil || gid <= 0 || v == "" {
			continue
		}
		out[gid] = v
	}
	return out
}

func marshalGroupModes(modes map[int64]string) string {
	if len(modes) == 0 {
		return "{}"
	}
	raw := make(map[string]string, len(modes))
	for k, v := range modes {
		raw[fmt.Sprintf("%d", k)] = v
	}
	data, _ := json.Marshal(raw)
	return string(data)
}

func removeFileQuiet(path string) {
	if path == "" {
		return
	}
	os.Remove(path)
}
