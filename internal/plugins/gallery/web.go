package gallery

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

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
}

type galleryDTO struct {
	Name     string   `json:"name"`
	Mode     string   `json:"mode"`
	Aliases  []string `json:"aliases"`
	CoverPID uint     `json:"cover_pid"`
	PicCount int64    `json:"pic_count"`
}

func (p *pluginImpl) webListGalleries(c *fiber.Ctx) error {
	galleries, err := p.mgr.ListGalleries()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out := make([]galleryDTO, 0, len(galleries))
	for _, g := range galleries {
		out = append(out, galleryDTO{
			Name:     g.Name,
			Mode:     g.Mode,
			Aliases:  parseAliases(g.Aliases),
			CoverPID: g.CoverPID,
			PicCount: p.mgr.PicCount(g.Name),
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
		Mode     *string `json:"mode,omitempty"`
		AddAlias *string `json:"add_alias,omitempty"`
		DelAlias *string `json:"del_alias,omitempty"`
		CoverPID *uint   `json:"cover_pid,omitempty"`
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
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "pid": pid})
}
