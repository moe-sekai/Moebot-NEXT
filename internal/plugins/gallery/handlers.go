package gallery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func (p *pluginImpl) registerHandlers() *zero.Engine {
	engine := zero.New()

	// 普通命令
	engine.OnRegex(`^/(看|gall pick)\s*(.*)`, zero.OnlyGroup).SetBlock(true).Handle(p.handlePick)
	engine.OnRegex(`^/(看所有|看全部|gall list)\s*(.*)`, zero.OnlyGroup).SetBlock(true).Handle(p.handleList)
	engine.OnRegex(`^/(上传|添加|gall add|gall upload)\s*(.*)`, zero.OnlyGroup).SetBlock(true).Handle(p.handleAdd)
	engine.OnRegex(`^/(取消上传|撤销上传|回退上传|gall cancel|gall revert)\s*(.*)`, zero.OnlyGroup).SetBlock(true).Handle(p.handleCancel)
	engine.OnRegex(`^/(上传记录|gall record)\s*(.*)`, zero.OnlyGroup).SetBlock(true).Handle(p.handleRecord)

	// 管理命令
	engine.OnRegex(`^/gall open\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleOpen)
	engine.OnRegex(`^/gall close\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleClose)
	engine.OnRegex(`^/gall mode\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleMode)
	engine.OnRegex(`^/gall (del|remove)\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleDel)
	engine.OnRegex(`^/gall replace\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleReplace)
	engine.OnRegex(`^/gall (reload|update)\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleReload)
	engine.OnRegex(`^/gall alias add\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleAliasAdd)
	engine.OnRegex(`^/gall alias del\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleAliasDel)
	engine.OnRegex(`^/gall cover\s+(.+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleCover)
	engine.OnRegex(`^/gall check\s*(.*)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleCheck)

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
			absPath, _ := filepath.Abs(pic.Path)
			msgs = append(msgs, message.Image("file:///"+absPath))
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
			if g.Mode == "off" && !isSuperUser(ctx) {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"已关闭", gallName)))
				return
			}
			pic, err := p.mgr.FindPicByNegativeIndex(g.Name, -n)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(err.Error()))
				return
			}
			absPath, _ := filepath.Abs(pic.Path)
			ctx.SendChain(message.Image("file:///" + absPath))
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
	if g.Mode == "off" && !isSuperUser(ctx) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"已关闭", name)))
		return
	}

	pics, err := p.mgr.RandomPics(g.Name, num)
	if err != nil || len(pics) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"没有图片", name)))
		return
	}
	var msgs []message.Segment
	for _, pic := range pics {
		absPath, _ := filepath.Abs(pic.Path)
		msgs = append(msgs, message.Image("file:///"+absPath))
	}
	ctx.SendChain(msgs...)
}

func (p *pluginImpl) handleList(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	name := strings.TrimSpace(matches[2])

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
	if g.Mode == "off" && !isSuperUser(ctx) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"已关闭", name)))
		return
	}
	pics, _ := p.mgr.ListPics(g.Name, 0, 0)
	if len(pics) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"没有图片", name)))
		return
	}
	var pids []string
	for _, pic := range pics {
		pids = append(pids, strconv.Itoa(int(pic.PID)))
	}
	text := fmt.Sprintf("画廊\"%s\"共%d张图片:\n%s", g.Name, len(pics), strings.Join(pids, " "))
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
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
	if g.Mode != "edit" && !isSuperUser(ctx) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("画廊\"%s\"不允许上传图片", name)))
		return
	}

	imageURLs := extractImageURLs(ctx.Event.Message)
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
		hid, _ = p.mgr.AddUploadRecord(ctx.Event.UserID, okList)
	}

	msg := ""
	if hid > 0 {
		msg += fmt.Sprintf("[#%d] ", hid)
	}
	msg += fmt.Sprintf("成功上传%d/%d张图片到\"%s\"", len(okList), len(imageURLs), g.Name)
	if len(errMsgs) > 0 {
		msg += "\n" + strings.Join(errMsgs, "\n")
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
	imageURLs := extractImageURLs(ctx.Event.Message)
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
