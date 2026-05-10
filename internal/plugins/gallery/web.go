package gallery

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// paramName 从路径参数中取出名称并做 URL 解码。
// fiber 的 c.Params 不会自动 URL 解码，因此中文等非 ASCII 名称会
// 以百分号编码形式传到 handler，需要显式解码。
func paramName(c *fiber.Ctx) string {
	raw := c.Params("name")
	if raw == "" {
		return raw
	}
	if decoded, err := url.PathUnescape(raw); err == nil {
		return decoded
	}
	return raw
}

func (p *pluginImpl) registerWebRoutes(api fiber.Router) {
	g := api.Group("/plugins/" + PluginName)

	g.Get("/galleries", p.webListGalleries)
	g.Post("/galleries", p.webCreateGallery)
	g.Delete("/galleries/:name", p.webDeleteGallery)
	g.Put("/galleries/:name", p.webUpdateGallery)
	g.Get("/galleries/:name/pics", p.webListPics)

	g.Get("/pics/:pid/image", p.webGetPicImage)
	g.Get("/pics/:pid/thumb", p.webGetPicThumb)
	g.Delete("/pics/:pid", p.webDeletePic)
	g.Post("/upload", p.webUploadPic)

	g.Get("/upload-records", p.webListUploadRecords)
	g.Post("/upload-records/:id/revert", p.webRevertUploadRecord)
}

type galleryDTO struct {
	Name       string            `json:"name"`
	Mode       string            `json:"mode"`
	GroupModes map[string]string `json:"group_modes"` // 前端用 string key 方便 JSON 处理
	Aliases    []string          `json:"aliases"`
	CoverPID   uint              `json:"cover_pid"`
	// CoverThumbPID 是用于在控制台显示缩略图的真实 PID。
	// 当用户已用 /gall cover 设置过 → 等于 CoverPID；
	// 否则后端回退到该画廊"最新一张图"的 PID，让卡片立即有缩略图。
	CoverThumbPID uint  `json:"cover_thumb_pid"`
	PicCount      int64 `json:"pic_count"`
}

func toGroupModesDTO(s string) map[string]string {
	parsed := parseGroupModes(s)
	out := make(map[string]string, len(parsed))
	for k, v := range parsed {
		out[strconv.FormatInt(k, 10)] = v
	}
	return out
}

func (p *pluginImpl) webListGalleries(c *fiber.Ctx) error {
	galleries, err := p.mgr.ListGalleries()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out := make([]galleryDTO, 0, len(galleries))
	for _, g := range galleries {
		thumbPID := g.CoverPID
		if thumbPID == 0 {
			// 用户没设过封面：回退到最新一张图，直接给卡片一个能显示的缩略图
			var latest GalleryPic
			if err := p.mgr.db.Where("gall_name = ?", g.Name).Order("pid DESC").Limit(1).Take(&latest).Error; err == nil {
				thumbPID = latest.PID
			}
		}
		out = append(out, galleryDTO{
			Name:          g.Name,
			Mode:          g.Mode,
			GroupModes:    toGroupModesDTO(g.GroupModes),
			Aliases:       parseAliases(g.Aliases),
			CoverPID:      g.CoverPID,
			CoverThumbPID: thumbPID,
			PicCount:      p.mgr.PicCount(g.Name),
		})
	}
	return c.JSON(fiber.Map{"galleries": out})
}

func (p *pluginImpl) webCreateGallery(c *fiber.Ctx) error {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := p.mgr.CreateGallery(body.Name); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "name": body.Name})
}

func (p *pluginImpl) webDeleteGallery(c *fiber.Ctx) error {
	name := paramName(c)
	if err := p.mgr.DeleteGallery(name); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (p *pluginImpl) webUpdateGallery(c *fiber.Ctx) error {
	name := paramName(c)
	var body struct {
		Mode       *string            `json:"mode,omitempty"`
		AddAlias   *string            `json:"add_alias,omitempty"`
		DelAlias   *string            `json:"del_alias,omitempty"`
		CoverPID   *uint              `json:"cover_pid,omitempty"`
		GroupModes *map[string]string `json:"group_modes,omitempty"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.Mode != nil {
		if _, _, err := p.mgr.SetMode(name, *body.Mode); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	if body.AddAlias != nil {
		if err := p.mgr.AddAlias(name, *body.AddAlias); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	if body.DelAlias != nil {
		if err := p.mgr.DelAlias(name, *body.DelAlias); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	if body.CoverPID != nil {
		if err := p.mgr.SetCover(name, *body.CoverPID); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	if body.GroupModes != nil {
		modes := map[int64]string{}
		for k, v := range *body.GroupModes {
			gid, err := strconv.ParseInt(k, 10, 64)
			if err != nil || gid <= 0 {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid group id: %s", k))
			}
			modes[gid] = v
		}
		if err := p.mgr.ReplaceGroupModes(name, modes); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (p *pluginImpl) webListPics(c *fiber.Ctx) error {
	name := paramName(c)
	g, err := p.mgr.FindGallery(name)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("画廊\"%s\"不存在", name))
	}
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	pics, err := p.mgr.ListPics(g.Name, offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	total := p.mgr.PicCount(g.Name)
	return c.JSON(fiber.Map{"pics": pics, "total": total})
}

func (p *pluginImpl) webGetPicImage(c *fiber.Ctx) error {
	pid, err := strconv.Atoi(c.Params("pid"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid pid")
	}
	pic, err := p.mgr.FindPic(uint(pid))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "pic not found")
	}
	return c.SendFile(pic.Path)
}

func (p *pluginImpl) webGetPicThumb(c *fiber.Ctx) error {
	pid, err := strconv.Atoi(c.Params("pid"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid pid")
	}
	pic, err := p.mgr.FindPic(uint(pid))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "pic not found")
	}
	if pic.ThumbPath == "" {
		thumbPath, tErr := ensureThumb(pic.Path)
		if tErr != nil {
			return c.SendFile(pic.Path)
		}
		pic.ThumbPath = thumbPath
		p.mgr.db.Save(pic)
	}
	return c.SendFile(pic.ThumbPath)
}

func (p *pluginImpl) webDeletePic(c *fiber.Ctx) error {
	pid, err := strconv.Atoi(c.Params("pid"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid pid")
	}
	if err := p.mgr.DelPic(uint(pid)); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (p *pluginImpl) webUploadPic(c *fiber.Ctx) error {
	gallName := c.FormValue("gallery")
	if gallName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing gallery name")
	}
	g, err := p.mgr.FindGallery(gallName)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("画廊\"%s\"不存在", gallName))
	}
	checkDup := c.FormValue("check_dup", "true") == "true"

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "missing file")
	}

	src, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "open file: "+err.Error())
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	tmpFile, err := os.CreateTemp("", "gallery_upload_*"+ext)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "create temp: "+err.Error())
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, src); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "save: "+err.Error())
	}
	tmpFile.Close()

	pid, err := p.mgr.AddPic(g.Name, tmpFile.Name(), checkDup)
	if err != nil {
		// 去重冲突：附带 hint，前端可以提示用户"取消勾选去重后重试"。
		msg := err.Error()
		if checkDup && strings.Contains(msg, "相似图片") {
			msg += `（可在上传时取消勾选"去重"或使用 force 强制上传）`
		}
		return fiber.NewError(fiber.StatusBadRequest, msg)
	}
	return c.JSON(fiber.Map{"ok": true, "pid": pid})
}

type uploadRecordDTO struct {
	ID        uint   `json:"id"`
	UserID    int64  `json:"user_id"`
	GroupID   int64  `json:"group_id"`
	GallName  string `json:"gall_name"`
	PIDs      []uint `json:"pids"`
	Reverted  bool   `json:"reverted"`
	CreatedAt string `json:"created_at"`
}

// webListUploadRecords 返回上传记录，可按 user_id / group_id / gallery 过滤。
func (p *pluginImpl) webListUploadRecords(c *fiber.Ctx) error {
	uid, _ := strconv.ParseInt(c.Query("user_id", "0"), 10, 64)
	gid, _ := strconv.ParseInt(c.Query("group_id", "0"), 10, 64)
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	gall := c.Query("gallery", "")
	if gall != "" {
		if decoded, err := url.QueryUnescape(gall); err == nil {
			gall = decoded
		}
	}
	rows, total, err := p.mgr.ListUploadRecords(UploadRecordFilter{
		UserID: uid, GroupID: gid, GallName: gall, Offset: offset, Limit: limit,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out := make([]uploadRecordDTO, 0, len(rows))
	for _, r := range rows {
		var pids []uint
		_ = parseJSON(r.PIDs, &pids)
		out = append(out, uploadRecordDTO{
			ID: r.ID, UserID: r.UserID, GroupID: r.GroupID, GallName: r.GallName,
			PIDs: pids, Reverted: r.Reverted,
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return c.JSON(fiber.Map{"records": out, "total": total})
}

// webRevertUploadRecord 撤销指定上传记录（删除其所有 PID 对应的图片）。
func (p *pluginImpl) webRevertUploadRecord(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	okList, errList, err := p.mgr.RevertUploadRecord(uint(id))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "deleted": okList, "failed": errList})
}
