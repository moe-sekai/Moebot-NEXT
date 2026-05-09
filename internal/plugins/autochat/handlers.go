package autochat

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// registerHandlers 在独立的 ZeroBot Engine 上注册所有处理器。
// 返回 engine 句柄，便于插件禁用时调用 engine.Delete() 注销。
func (p *pluginImpl) registerHandlers() *zero.Engine {
	engine := zero.New()

	engine.OnRegex(`^/chat\s*(.*)`, zero.OnlyGroup).SetBlock(true).Handle(p.handleChat)
	engine.OnMessage(zero.OnlyGroup).SetBlock(false).Handle(p.handleAutoReply)
	engine.OnRegex(`^/(模型|chatmodel)(\s+.*)?$`, zero.OnlyGroup).SetBlock(true).Handle(p.handleModel)
	engine.OnFullMatchGroup([]string{"/模型列表", "/modellist"}).SetBlock(true).Handle(p.handleModelList)
	engine.OnRegex(`^/(开启|关闭)(聊天|autochat)\s*$`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(p.handleWhiteList)
	engine.OnFullMatchGroup([]string{"/消耗统计", "/token统计"}).SetBlock(true).Handle(p.handleTokenStats)
	engine.OnRegex(`^/查询记忆(\s+.*)?$`, zero.OnlyGroup).SetBlock(true).Handle(p.handleQueryMemory)

	return engine
}

func (p *pluginImpl) handleChat(ctx *zero.Ctx) {
	groupID := ctx.Event.GroupID
	userID := ctx.Event.UserID
	// Filter 网关：控制台「Filter」页面给本插件分配的 internal app 模板，
	// 决定该 group/user/消息是否被允许触发对话。失配时静默忽略。
	if !p.allowedByFilter(groupID, userID, false, ctx.Event.RawMessage) {
		return
	}
	// 兼容旧 /开启聊天 /关闭聊天 命令的本地白名单。
	if !p.chatWhiteList.Check(groupID) {
		return
	}
	cdKey := fmt.Sprintf("%d_%d", groupID, userID)
	if !p.chatCD.Check(cdKey) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("冷却中，%d秒", p.chatCD.Remaining(cdKey))))
		return
	}
	matches := ctx.State["regex_matched"].([]string)
	queryText := strings.TrimSpace(matches[1])
	if queryText == "" {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请输入要对话的内容"))
		return
	}
	nick := ctx.CardOrNickName(userID)
	msgTime := time.Now().Unix()
	p.messageBuffer.Add(groupID, nick, userID, queryText, msgTime)
	images := ExtractImageURLs(ctx.Event.Message)
	if len(images) > 0 {
		go p.processImageDescription(groupID, userID, msgTime, images, GetConfig())
	}
	p.processChat(ctx, groupID, userID, queryText, false)
}

// handleAutoReply 负责自动回复（@bot / 关键词 / 阈值随机触发）。
func (p *pluginImpl) handleAutoReply(ctx *zero.Ctx) {
	groupID := ctx.Event.GroupID
	userID := ctx.Event.UserID
	msg := ctx.Event.Message
	if !p.allowedByFilter(groupID, userID, false, ctx.Event.RawMessage) {
		return
	}
	if !p.chatWhiteList.Check(groupID) {
		return
	}

	var b strings.Builder
	for _, seg := range msg {
		if seg.Type == "text" {
			b.WriteString(seg.Data["text"])
		} else if seg.Type == "at" {
			if qq, ok := seg.Data["qq"]; ok {
				b.WriteString(fmt.Sprintf(" @%s ", qq))
			}
		}
	}
	text := strings.TrimSpace(b.String())
	// 提取剔除 @bot 之后的“纯文本”，用于判定是否是其它插件的命令。
	pureText := strings.TrimSpace(extractPureText(msg, ctx.Event.SelfID))
	if isIgnoredCommand(pureText) {
		return
	}

	if text != "" || len(ExtractImageURLs(msg)) > 0 {
		nick := ctx.CardOrNickName(userID)
		msgTime := time.Now().Unix()
		logText := text
		if logText == "" {
			logText = "[图片消息]"
		}
		p.messageBuffer.Add(groupID, nick, userID, logText, msgTime)
		images := ExtractImageURLs(msg)
		if len(images) > 0 {
			go p.processImageDescription(groupID, userID, msgTime, images, GetConfig())
		}
	}

	cfg := GetConfig()
	delta := 0.0
	isDirect := false
	if HasAt(msg, ctx.Event.SelfID) {
		delta = 2.5
		isDirect = true
	}
	if !isDirect {
		if !p.autoWhiteList.Check(groupID) {
			return
		}
		for _, kw := range cfg.Chat.Keywords {
			if strings.Contains(text, kw) {
				delta = 1.0
				isDirect = true
				break
			}
		}
		if !isDirect && text != "" {
			delta = randFloat() * 0.2
		}
	}

	cur := p.threshold(groupID)
	target := cfg.Chat.Willing.Threshold
	if g, ok := cfg.Chat.Willing.GroupThresholds[fmt.Sprintf("%d", groupID)]; ok {
		target = g
	}
	newVal := cur + delta
	if newVal >= target {
		p.setThreshold(groupID, 0)
		cdKey := fmt.Sprintf("%d_%d", groupID, userID)
		if !p.chatCD.Check(cdKey) {
			return
		}
		// 去掉对自己的 @
		text = strings.ReplaceAll(text, fmt.Sprintf("@%d", ctx.Event.SelfID), "")
		text = strings.TrimSpace(text)
		log.Debug().Float64("threshold", newVal).Float64("target", target).Bool("direct", isDirect).Msg("[autochat] 阈值触发")
		p.processChat(ctx, groupID, userID, text, !isDirect)
	} else {
		p.setThreshold(groupID, newVal)
	}
}

func (p *pluginImpl) handleModel(ctx *zero.Ctx) {
	groupID := ctx.Event.GroupID
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[2])
	key := fmt.Sprintf("model_%d", groupID)
	if args == "" {
		cur := p.fileDB.GetString(key)
		if cur == "" {
			cfg := GetConfig()
			if len(cfg.LLM.Models) > 0 {
				cur = cfg.LLM.Models[0]
			}
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("当前模型: %s", cur)))
		return
	}
	if !zero.SuperUserPermission(ctx) && !zero.AdminPermission(ctx) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("仅管理员可切换模型"))
		return
	}
	_ = p.fileDB.Set(key, args)
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("已切换模型为: "+args))
}

func (p *pluginImpl) handleModelList(ctx *zero.Ctx) {
	cfg := GetConfig()
	if len(cfg.LLM.Models) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("尚未配置任何模型"))
		return
	}
	out := "可用模型列表:\n"
	for _, m := range cfg.LLM.Models {
		out += "- " + m + "\n"
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(out))
}

func (p *pluginImpl) handleWhiteList(ctx *zero.Ctx) {
	matches := ctx.State["regex_matched"].([]string)
	action, target := matches[1], matches[2]
	groupID := ctx.Event.GroupID
	wl := p.chatWhiteList
	if target == "autochat" {
		wl = p.autoWhiteList
	}
	if action == "开启" {
		_ = wl.Add(groupID)
	} else {
		_ = wl.Remove(groupID)
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("已%s%s功能", action, target)))
}

func (p *pluginImpl) handleTokenStats(ctx *zero.Ctx) {
	pt1, ct1, rc1 := p.tokenStats.GetStats(1)
	pt7, ct7, rc7 := p.tokenStats.GetStats(7)
	text := fmt.Sprintf(
		"📊 Token 消耗统计\n\n📅 今日: 请求 %d 次 | 输入 %d | 输出 %d | 合计 %d\n📅 近7天: 请求 %d 次 | 输入 %d | 输出 %d | 合计 %d",
		rc1, pt1, ct1, pt1+ct1, rc7, pt7, ct7, pt7+ct7,
	)
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
}

func (p *pluginImpl) handleQueryMemory(ctx *zero.Ctx) {
	groupID := ctx.Event.GroupID
	senderID := ctx.Event.UserID
	matches := ctx.State["regex_matched"].([]string)
	args := strings.TrimSpace(matches[1])
	atList := ExtractAtQQ(ctx.Event.Message)

	if len(atList) > 0 {
		p.queryUserMemory(ctx, groupID, atList[0], fmt.Sprintf("%d", atList[0]))
		return
	}
	if args == "" {
		p.queryUserMemory(ctx, groupID, senderID, ctx.CardOrNickName(senderID))
		return
	}
	if id, err := strconv.ParseInt(args, 10, 64); err == nil {
		p.queryUserMemory(ctx, groupID, id, args)
		return
	}
	p.queryMemoryByKeyword(ctx, groupID, args)
}

func (p *pluginImpl) queryUserMemory(ctx *zero.Ctx, groupID, userID int64, name string) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[记忆查询] %s (%d)\n", name, userID))
	if local, err := p.memory.GetUserMemory(groupID, userID); err == nil && local != "" {
		b.WriteString("\n📌 用户画像:\n")
		b.WriteString(local)
	} else {
		b.WriteString("\n📌 用户画像: (无)")
	}
	if vc := GetVectorClient(); vc != nil && vc.IsEnabled() {
		if mems, err := vc.QueryUserMemories(groupID, userID, 5); err == nil && len(mems) > 0 {
			b.WriteString("\n\n📜 历史记忆片段:\n")
			for i, m := range mems {
				text := m.Text
				if len([]rune(text)) > 100 {
					text = string([]rune(text)[:100]) + "..."
				}
				b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, formatTimestamp(m.Timestamp), text))
			}
		}
	}
	b.WriteString("\n\n💡 输入 /查询记忆 <关键词> 可语义检索")
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(b.String()))
}

func (p *pluginImpl) queryMemoryByKeyword(ctx *zero.Ctx, groupID int64, keyword string) {
	vc := GetVectorClient()
	if vc == nil || !vc.IsEnabled() {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("向量搜索功能未启用"))
		return
	}
	mems, err := vc.QueryMemoriesByKeyword(groupID, keyword, 20)
	if err != nil {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("搜索失败: "+err.Error()))
		return
	}
	if len(mems) == 0 {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("未找到相关记忆"))
		return
	}
	if rc := GetRerankClient(); rc != nil && rc.IsEnabled() {
		rctx, rcancel := context.WithTimeout(context.Background(), 15*time.Second)
		if r, err := rc.RerankMemories(rctx, keyword, mems); err == nil && len(r) > 0 {
			mems = r
		}
		rcancel()
	}
	if len(mems) > 5 {
		mems = mems[:5]
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("[vector] 关键词 '%s' 命中 %d 条:\n\n", keyword, len(mems)))
	for i, m := range mems {
		name := fmtUserName(m.UserName, m.UserID)
		text := m.Text
		if len([]rune(text)) > 60 {
			text = string([]rune(text)[:60]) + "..."
		}
		b.WriteString(fmt.Sprintf("%d. %s (%s)\n   %s\n", i+1, name, formatTimestamp(m.Timestamp), text))
	}
	ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(b.String()))
}

// 阈值状态（仅内存）。
var (
	thresholdMu sync.RWMutex
)

func (p *pluginImpl) threshold(groupID int64) float64 {
	thresholdMu.RLock()
	defer thresholdMu.RUnlock()
	return p.thresholds[groupID]
}

func (p *pluginImpl) setThreshold(groupID int64, v float64) {
	thresholdMu.Lock()
	p.thresholds[groupID] = v
	thresholdMu.Unlock()
}

// randFloat 返回 [0,1) 伪随机数，避免在多个文件重复 import math/rand。
func randFloat() float64 {
	return float64(time.Now().UnixNano()%10000) / 10000.0
}
