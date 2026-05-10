package gallery

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"moebot-next/internal/renderer"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// textRegexRule 仅对消息中的文本段（type == "text"）做正则匹配，
// 忽略消息开头的 [CQ:image]、[CQ:reply] 等非文本段。
// 这样即使用户把图片或 reply 段放在指令前面，也能正确解析指令。
func textRegexRule(pattern string) zero.Rule {
	re := regexp.MustCompile(pattern)
	return func(ctx *zero.Ctx) bool {
		var sb strings.Builder
		for _, seg := range ctx.Event.Message {
			if seg.Type == "text" {
				sb.WriteString(seg.Data["text"])
			}
		}
		text := strings.TrimSpace(sb.String())
		if text == "" {
			return false
		}
		matched := re.FindStringSubmatch(text)
		if matched == nil {
			return false
		}
		ctx.State["regex_matched"] = matched
		return true
	}
}

// readPicBytes 读取图片字节，供 message.ImageBytes 使用。
// 与 moesekai 等其它插件保持一致的发图方式：bot 与 OneBot 客户端常常
// 不在同一文件系统（不同容器/不同机器），file:// 不可达；用字节流由
// zero-bot 自动编为 base64 是最稳的做法。
func readPicBytes(p string) ([]byte, error) {
	if p == "" {
		return nil, fmt.Errorf("path empty")
	}
	return os.ReadFile(p)
}

// gate 包装一个 handler：在调用前先咨询 filter 网关；未放行则静默丢弃，
// 让控制台 /filter 页面对画廊插件的群/用户/正则规则真正生效。
func (p *pluginImpl) gate(h func(*zero.Ctx)) func(*zero.Ctx) {
	return func(ctx *zero.Ctx) {
		ev := ctx.Event
		if ev == nil {
			h(ctx)
			return
		}
		isPrivate := ev.MessageType == "private" || ev.DetailType == "private"
		if !p.allowedByFilter(ev.GroupID, ev.UserID, isPrivate, ev.RawMessage) {
			return
		}
		h(ctx)
	}
}

func (p *pluginImpl) registerHandlers() *zero.Engine {
	engine := zero.New()

	// 普通命令
	// 注意：`/看所有|看全部|gall list` 必须比 `/看` 先注册，否则
	// `/看全部冬雪` 会被 `/看` 先命中 (args="全部冬雪") 而不进入 list 处理。
	engine.OnMessage(textRegexRule(`^/(看所有|看全部|gall list)\s*(.*)`), zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleList))
	engine.OnMessage(textRegexRule(`^/(看|gall pick)\s*(.*)`), zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handlePick))
	// 同理：更具体的子命令 (上传记录 / 取消上传) 必须先注册，否则前缀更短的
	// `/上传` 会先匹配 `/上传记录 1`、把 args 解析为"记录 1"。
	engine.OnMessage(textRegexRule(`^/(上传记录|gall record)\s*(.*)`), zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleRecord))
	engine.OnMessage(textRegexRule(`^/(取消上传|撤销上传|回退上传|gall cancel|gall revert)\s*(.*)`), zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleCancel))
	engine.OnMessage(textRegexRule(`^/(上传|添加|gall add|gall upload)\s*(.*)`), zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleAdd))

	// 管理命令
	engine.OnMessage(textRegexRule(`^/gall open\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleOpen))
	engine.OnMessage(textRegexRule(`^/gall close\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleClose))
	engine.OnMessage(textRegexRule(`^/gall mode\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleMode))
	engine.OnMessage(textRegexRule(`^/gall (del|remove)\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleDel))
	engine.OnMessage(textRegexRule(`^/gall replace\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleReplace))
	engine.OnMessage(textRegexRule(`^/gall (reload|update)\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleReload))
	engine.OnMessage(textRegexRule(`^/gall alias add\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleAliasAdd))
	engine.OnMessage(textRegexRule(`^/gall alias del\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleAliasDel))
	engine.OnMessage(textRegexRule(`^/gall cover\s+(.+)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleCover))
	engine.OnMessage(textRegexRule(`^/gall check\s*(.*)`), zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.gate(p.handleCheck))

	return engine
}

// --- 普通命令 ---

func (p *pluginImpl) handlePick(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[2])
	if args == "" {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /看 画廊名"))
		return
	}
	cfg := getConfig()
	limit := cfg.PickLimit

	// 尝试解析为 PID 列表
	parts := strings.Fields(args)
	allInt := true
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			allInt = false
			break
		}
	}

	if allInt && len(parts) > 0 {
		if len(parts) > limit {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("一次最多查看%d张图片", limit)))
			return
		}
		var msgs []message.Segment
		for _, part := range parts {
			pid, _ := strconv.Atoi(part)
			var pic *GalleryPic
			var err error
			if pid < 0 {
				// 全局倒数查找：先获取所有 PID 排序
				var allPics []GalleryPic
				p.mgr.db.Order("pid").Find(&allPics)
				idx := len(allPics) + pid
				if idx < 0 || idx >= len(allPics) {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊仅有%d张图片", len(allPics))))
					return
				}
				pic = &allPics[idx]
			} else {
				pic, err = p.mgr.FindPic(uint(pid))
			}
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("图片pid=%d不存在", pid)))
				return
			}
			data, rErr := readPicBytes(pic.Path)
			if rErr != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("图片pid=%d文件读取失败(%s): %s", pic.PID, pic.Path, rErr.Error())))
				return
			}
			msgs = append(msgs, message.ImageBytes(data))
		}
		ctx.SendChain(msgs...)
		return
	}

	// 解析 画廊名 [xN] [-N]
	name := args
	num := 1
	name = strings.ReplaceAll(name, "*", "x")
	name = strings.ReplaceAll(name, "×", "x")

	if strings.Contains(name, "-") {
		lastDash := strings.LastIndex(name, "-")
		nStr := name[lastDash+1:]
		if n, err := strconv.Atoi(nStr); err == nil && n > 0 {
			gallName := strings.TrimSpace(name[:lastDash])
			g, gErr := p.mgr.FindGallery(gallName)
			if gErr != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"不存在", gallName)))
				return
			}
			if p.mgr.EffectiveMode(g, ctx.Event.GroupID) == "off" && !isSuperUser(ctx) {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"在本群已关闭", gallName)))
				return
			}
			pic, err := p.mgr.FindPicByNegativeIndex(g.Name, -n)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
				return
			}
			data, rErr := readPicBytes(pic.Path)
			if rErr != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("图片pid=%d文件读取失败(%s): %s", pic.PID, pic.Path, rErr.Error())))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
			return
		}
	}

	if strings.Contains(name, "x") {
		lastX := strings.LastIndex(name, "x")
		nStr := name[lastX+1:]
		if n, err := strconv.Atoi(nStr); err == nil && n > 0 {
			name = strings.TrimSpace(name[:lastX])
			num = n
		}
	}

	if num > limit {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("一次最多查看%d张图片", limit)))
		return
	}

	g, err := p.mgr.FindGallery(name)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"不存在", name)))
		return
	}
	if p.mgr.EffectiveMode(g, ctx.Event.GroupID) == "off" && !isSuperUser(ctx) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"在本群已关闭", name)))
		return
	}

	pics, err := p.mgr.RandomPics(g.Name, num)
	if err != nil || len(pics) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"没有图片", name)))
		return
	}
	var msgs []message.Segment
	var missing []string
	for _, pic := range pics {
		data, rErr := readPicBytes(pic.Path)
		if rErr != nil {
			missing = append(missing, fmt.Sprintf("pid=%d (%s): %s", pic.PID, pic.Path, rErr.Error()))
			continue
		}
		msgs = append(msgs, message.ImageBytes(data))
	}
	if len(msgs) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("画廊\""+g.Name+"\"中所有图片读取失败:\n"+strings.Join(missing, "\n")))
		return
	}
	if len(missing) > 0 {
		msgs = append(msgs, message.Text(fmt.Sprintf("\n[警告] %d 张图片读取失败", len(missing))))
	}
	ctx.SendChain(msgs...)
}

// pageRe 解析 "画廊名 @N" / "画廊名@N" 形式的页码后缀。
var pageRe = regexp.MustCompile(`\s*@\s*(\d+)\s*$`)

func (p *pluginImpl) handleList(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	name := strings.TrimSpace(matches[2])

	// 解析尾部 @N 当页码（默认 1）
	page := 1
	if m := pageRe.FindStringSubmatch(name); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			page = n
		}
		name = strings.TrimSpace(pageRe.ReplaceAllString(name, ""))
	}

	if name == "" {
		galleries, err := p.mgr.ListGalleries()
		if err != nil || len(galleries) == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("当前没有任何画廊"))
			return
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("共%d个画廊:", len(galleries)))
		for _, g := range galleries {
			count := p.mgr.PicCount(g.Name)
			aliases := parseAliases(g.Aliases)
			aliasStr := ""
			if len(aliases) > 0 {
				aliasStr = " 别名:" + strings.Join(aliases, ",")
			}
			lines = append(lines, fmt.Sprintf("· %s [%s] %d张%s", g.Name, g.Mode, count, aliasStr))
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(strings.Join(lines, "\n")))
		return
	}

	g, err := p.mgr.FindGallery(name)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"不存在", name)))
		return
	}
	if p.mgr.EffectiveMode(g, ctx.Event.GroupID) == "off" && !isSuperUser(ctx) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"在本群已关闭", name)))
		return
	}
	pics, _ := p.mgr.ListPics(g.Name, 0, 0)
	if len(pics) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"没有图片", name)))
		return
	}

	p.sendGalleryGrid(ctx, g.Name, pics, page)
}

// sendGalleryGrid 调用 satori 渲染服务把当前页缩略图拼成一张大图发送。
// 渲染失败时回退为文本 PID 列表。
func (p *pluginImpl) sendGalleryGrid(ctx *zero.Ctx, gallName string, pics []GalleryPic, page int) {
	const perPage = 100
	total := len(pics)
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}
	if page <= 0 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}
	pageSlice := pics[start:end]

	if p.rendererCl == nil || !p.rendererCl.Health() {
		// 渲染服务不可用：发文本列表
		var pids []string
		for _, pic := range pageSlice {
			pids = append(pids, strconv.Itoa(int(pic.PID)))
		}
		text := fmt.Sprintf("画廊\"%s\"第%d/%d页 共%d张:\n%s", gallName, page, totalPages, total, strings.Join(pids, " "))
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
		return
	}

	// 把每张缩略图字节转成 data URI
	type dtoPic struct {
		PID     uint   `json:"pid"`
		DataURI string `json:"dataUri"`
	}
	var dtoPics []dtoPic
	for _, pic := range pageSlice {
		path := pic.ThumbPath
		if path == "" {
			path = pic.Path
		}
		data, err := readPicBytes(path)
		if err != nil {
			continue
		}
		mime := "image/jpeg"
		if strings.HasSuffix(strings.ToLower(path), ".png") {
			mime = "image/png"
		}
		dtoPics = append(dtoPics, dtoPic{
			PID:     pic.PID,
			DataURI: "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data),
		})
	}
	if len(dtoPics) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"第%d页所有缩略图读取失败", gallName, page)))
		return
	}

	payload := map[string]interface{}{
		"title":      "画廊 " + gallName,
		"subtitle":   fmt.Sprintf("第 %d/%d 页 · 共 %d 张 · 本页 %d 张", page, totalPages, total, len(dtoPics)),
		"pics":       dtoPics,
		"page":       page,
		"totalPages": totalPages,
		"total":      total,
	}
	png, err := p.rendererCl.Render(renderer.RenderRequest{Template: "gallery_grid", Data: payload})
	if err != nil {
		log.Warn().Err(err).Msg("[gallery] satori 渲染失败，回退文本列表")
		var pids []string
		for _, pic := range pageSlice {
			pids = append(pids, strconv.Itoa(int(pic.PID)))
		}
		text := fmt.Sprintf("画廊\"%s\"第%d/%d页 共%d张:\n%s\n（渲染失败: %s）", gallName, page, totalPages, total, strings.Join(pids, " "), err.Error())
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
		return
	}
	hint := fmt.Sprintf("画廊\"%s\" 第 %d/%d 页", gallName, page, totalPages)
	if totalPages > 1 {
		hint += "，使用 /看所有 " + gallName + " @N 切换页码"
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(hint), message.ImageBytes(png))
}

func (p *pluginImpl) handleAdd(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[2])

	checkDup := true
	if strings.Contains(args, "force") {
		checkDup = false
		args = strings.ReplaceAll(args, "force", "")
		args = strings.TrimSpace(args)
	}

	name := args
	g, err := p.mgr.FindGallery(name)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"不存在", name)))
		return
	}
	effMode := p.mgr.EffectiveMode(g, ctx.Event.GroupID)
	if effMode != "edit" && !isSuperUser(ctx) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"在本群不允许上传图片（当前模式: %s）", name, effMode)))
		return
	}

	imageURLs := collectImageURLs(ctx)
	if len(imageURLs) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请附加要上传的图片（可回复包含图片的消息）"))
		return
	}

	var okList []uint
	var errMsgs []string
	for i, url := range imageURLs {
		tmpPath, dlErr := downloadImage(url)
		if dlErr != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("第%d张下载失败: %s", i+1, dlErr.Error()))
			continue
		}
		pid, addErr := p.mgr.AddPic(g.Name, tmpPath, checkDup)
		os.Remove(tmpPath)
		if addErr != nil {
			errMsgs = append(errMsgs, fmt.Sprintf("第%d张: %s", i+1, addErr.Error()))
			continue
		}
		okList = append(okList, pid)
	}

	var hid uint
	if len(okList) > 0 {
		hid, _ = p.mgr.AddUploadRecord(ctx.Event.UserID, ctx.Event.GroupID, g.Name, okList)
	}

	msg := ""
	if hid > 0 {
		msg += fmt.Sprintf("[#%d] ", hid)
	}
	msg += fmt.Sprintf("成功上传%d/%d张图片到\"%s\"", len(okList), len(imageURLs), g.Name)
	if len(errMsgs) > 0 {
		msg += "\n" + strings.Join(errMsgs, "\n")
	}
	if len(okList) > 0 {
		msg += fmt.Sprintf("\n⚠ 请勿上传违规内容，你的QQ号(%d)已被记录。如需撤回请发送 /取消上传", ctx.Event.UserID)
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(msg))
}

func (p *pluginImpl) handleCancel(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[2])

	if args != "" {
		if !isSuperUser(ctx) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("仅管理员可撤销指定上传记录"))
			return
		}
		hid, err := strconv.Atoi(args)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /取消上传 [记录ID]"))
			return
		}
		okList, errList, err := p.mgr.RevertUploadRecord(uint(hid))
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(formatRevertResult(uint(hid), okList, errList)))
		return
	}

	r, okList, errList, err := p.mgr.RevertUserLastUpload(ctx.Event.UserID)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(formatRevertResult(r.ID, okList, errList)))
}

func (p *pluginImpl) handleRecord(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[2])
	hid, err := strconv.Atoi(args)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /上传记录 记录ID"))
		return
	}
	r, err := p.mgr.GetUploadRecord(uint(hid))
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("上传#%d不存在", hid)))
		return
	}
	var pids []uint
	_ = parseJSON(r.PIDs, &pids)
	reverted := ""
	if r.Reverted {
		reverted = " (已撤销)"
	}
	text := fmt.Sprintf("上传记录#%d%s\n用户: %d\n时间: %s\n图片: %s",
		r.ID, reverted, r.UserID, r.CreatedAt.Format("2006-01-02 15:04:05"), formatPIDs(pids))
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
}

// --- 管理命令 ---

func (p *pluginImpl) handleOpen(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	name := strings.TrimSpace(matches[1])
	if err := p.mgr.CreateGallery(name); err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"创建成功", name)))
}

func (p *pluginImpl) handleClose(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	name := strings.TrimSpace(matches[1])
	if err := p.mgr.DeleteGallery(name); err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"删除成功", name)))
}

func (p *pluginImpl) handleMode(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.Fields(strings.TrimSpace(matches[1]))
	if len(args) == 1 {
		g, err := p.mgr.FindGallery(args[0])
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"不存在", args[0])))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"当前模式: %s", g.Name, g.Mode)))
		return
	}
	if len(args) != 2 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /gall mode 画廊名称 模式(edit/view/off)"))
		return
	}
	old, new, err := p.mgr.SetMode(args[0], args[1])
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊模式修改: %s -> %s", old, new)))
}

func (p *pluginImpl) handleDel(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[2])

	var pids []int
	if strings.Contains(args, "-") && !strings.HasPrefix(args, "-") {
		parts := strings.SplitN(args, "-", 2)
		l, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		r, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil || r-l >= 20 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /gall del 123 456 或 /gall del 100-119 (最多20张)"))
			return
		}
		for i := l; i <= r; i++ {
			pids = append(pids, i)
		}
	} else {
		for _, s := range strings.Fields(args) {
			pid, err := strconv.Atoi(s)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /gall del 123 456 或 /gall del 100-119"))
				return
			}
			pids = append(pids, pid)
		}
	}

	var okList, errList []uint
	for _, pid := range pids {
		if err := p.mgr.DelPic(uint(pid)); err != nil {
			errList = append(errList, uint(pid))
		} else {
			okList = append(okList, uint(pid))
		}
	}
	msg := ""
	if len(okList) > 0 {
		msg += fmt.Sprintf("%d张图片删除成功: %s", len(okList), formatPIDs(okList))
	}
	if len(errList) > 0 {
		if msg != "" {
			msg += "\n"
		}
		msg += fmt.Sprintf("%d张图片删除失败: %s", len(errList), formatPIDs(errList))
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(msg))
}

func (p *pluginImpl) handleReplace(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[1])
	checkDup := true
	if strings.Contains(args, "force") {
		checkDup = false
		args = strings.ReplaceAll(args, "force", "")
		args = strings.TrimSpace(args)
	}
	pid, err := strconv.Atoi(args)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /gall replace pid"))
		return
	}
	imageURLs := collectImageURLs(ctx)
	if len(imageURLs) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请附加要替换的图片"))
		return
	}
	tmpPath, dlErr := downloadImage(imageURLs[0])
	if dlErr != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("下载图片失败: "+dlErr.Error()))
		return
	}
	defer os.Remove(tmpPath)

	if err := p.mgr.ReplacePic(uint(pid), tmpPath, checkDup); err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("成功替换图片pid=%d", pid)))
}

func (p *pluginImpl) handleReload(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	name := strings.TrimSpace(matches[2])
	newPIDs, delPIDs, err := p.mgr.ReloadGallery(name)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"重新加载完成\n新增: %d张 失效: %d张", name, len(newPIDs), len(delPIDs))))
}

func (p *pluginImpl) handleAliasAdd(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.Fields(strings.TrimSpace(matches[1]))
	if len(args) != 2 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /gall alias add 画廊名称 别名"))
		return
	}
	if err := p.mgr.AddAlias(args[0], args[1]); err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"添加别名\"%s\"成功", args[0], args[1])))
}

func (p *pluginImpl) handleAliasDel(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.Fields(strings.TrimSpace(matches[1]))
	if len(args) != 2 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /gall alias del 画廊名称 别名"))
		return
	}
	if err := p.mgr.DelAlias(args[0], args[1]); err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"删除别名\"%s\"成功", args[0], args[1])))
}

func (p *pluginImpl) handleCover(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.Fields(strings.TrimSpace(matches[1]))
	if len(args) != 2 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("使用方式: /gall cover 画廊名称 图片ID"))
		return
	}
	pid, err := strconv.Atoi(args[1])
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("图片ID必须是整数"))
		return
	}
	if err := p.mgr.SetCover(args[0], uint(pid)); err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
		return
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"封面设置为pid=%d", args[0], pid)))
}

func (p *pluginImpl) handleCheck(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[1])
	rehash := strings.Contains(args, "rehash")
	args = strings.ReplaceAll(args, "rehash", "")
	args = strings.TrimSpace(args)

	doCheck := func(name string) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("正在检查画廊\"%s\"...", name)))
		result, err := p.mgr.CheckDuplicates(name, rehash)
		if err != nil {
			ctx.SendChain(message.Text(err.Error()))
			return
		}
		if len(result) == 0 {
			ctx.SendChain(message.Text(fmt.Sprintf("画廊\"%s\"未发现重复图片", name)))
			return
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("画廊\"%s\"发现%d组重复图片:", name, len(result)))
		for firstPID, dups := range result {
			lines = append(lines, fmt.Sprintf("  %d ↔ %s", firstPID, formatPIDs(dups)))
		}
		ctx.SendChain(message.Text(strings.Join(lines, "\n")))
	}

	if args == "" || args == "all" {
		galleries, _ := p.mgr.ListGalleries()
		for _, g := range galleries {
			doCheck(g.Name)
		}
	} else {
		doCheck(args)
	}
}

// --- 工具函数 ---

func extractImageURLs(msg message.Message) []string {
	var urls []string
	for _, seg := range msg {
		if seg.Type == "image" {
			if u, ok := seg.Data["url"]; ok {
				urls = append(urls, u)
			}
		}
	}
	return urls
}

// collectImageURLs 从当前消息提取图片 url；若当前消息无图片但存在 reply 段，
// 则调用 OneBot get_msg API 取被回复消息中的图片。
func collectImageURLs(ctx *zero.Ctx) []string {
	if urls := extractImageURLs(ctx.Event.Message); len(urls) > 0 {
		return urls
	}
	// 找 reply 段
	for _, seg := range ctx.Event.Message {
		if seg.Type != "reply" {
			continue
		}
		idStr, ok := seg.Data["id"]
		if !ok || idStr == "" {
			continue
		}
		mid, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		replied := ctx.GetMessage(mid)
		if replied.Elements == nil {
			continue
		}
		if urls := extractImageURLs(replied.Elements); len(urls) > 0 {
			return urls
		}
	}
	return nil
}

func downloadImage(url string) (string, error) {
	reqCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ext := ".jpg"
	ct := resp.Header.Get("Content-Type")
	switch {
	case strings.Contains(ct, "png"):
		ext = ".png"
	case strings.Contains(ct, "gif"):
		ext = ".gif"
	case strings.Contains(ct, "webp"):
		ext = ".webp"
	}

	tmpFile, err := os.CreateTemp("", "gallery_*"+ext)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}

func isSuperUser(ctx *zero.Ctx) bool {
	for _, su := range zero.BotConfig.SuperUsers {
		if su == ctx.Event.UserID {
			return true
		}
	}
	return false
}

func formatPIDs[T ~uint | ~int](pids []T) string {
	parts := make([]string, len(pids))
	for i, pid := range pids {
		parts[i] = strconv.Itoa(int(pid))
	}
	return strings.Join(parts, " ")
}

func formatRevertResult(hid uint, okList, errList []uint) string {
	msg := fmt.Sprintf("撤销上传记录#%d\n", hid)
	if len(okList) > 0 {
		msg += fmt.Sprintf("%d张图片删除成功: %s\n", len(okList), formatPIDs(okList))
	}
	if len(errList) > 0 {
		msg += fmt.Sprintf("%d张图片删除失败: %s\n", len(errList), formatPIDs(errList))
	}
	return strings.TrimSpace(msg)
}

func parseJSON(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}
