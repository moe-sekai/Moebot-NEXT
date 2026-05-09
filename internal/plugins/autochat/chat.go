package autochat

import (
	"context"
	"encoding/xml"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// AutoChatXMLResponse 是 LLM 必须输出的 XML 结构。
// 选择 XML 而非 JSON 是因为 LLM 输出 XML 时格式更稳定（不易因引号/逗号丢格式）。
type AutoChatXMLResponse struct {
	XMLName         xml.Name     `xml:"response"`
	Replies         []string     `xml:"replies>reply"`
	DialogueSummary string       `xml:"dialogue_summary"`
	UpdateProfiles  []xmlEntryKV `xml:"update_profiles>profile"` // 用户画像（覆盖式）
	AddMemories     []xmlEntryKV `xml:"add_memories>memory"`     // 历史记忆（增量 RAG）
	ReplyToQQs      []int64      `xml:"reply_to_qqs>qq"`
}

// xmlEntryKV 表示一条带 qq 属性的键值条目，例如：
//
//	<profile qq="123456">覆盖式用户画像</profile>
type xmlEntryKV struct {
	QQ    string `xml:"qq,attr"`
	Value string `xml:",chardata"`
}

// extractXMLBlock 从 LLM 原文中提取 <response>...</response> 片段。
// 兼容代码块包裹（```xml ... ```）以及前后多余文本。
func extractXMLBlock(raw string) string {
	s := strings.TrimSpace(raw)
	start := strings.Index(s, "<response>")
	end := strings.LastIndex(s, "</response>")
	if start >= 0 && end > start {
		return s[start : end+len("</response>")]
	}
	return s
}

// processChat 处理一次对话：拼 system prompt → 调 LLM → 解析 XML → 发送回复 + 更新记忆。
func (p *pluginImpl) processChat(ctx *zero.Ctx, groupID, userID int64, queryText string, allowTargetSelection bool) {
	cfg := GetConfig()
	msg := ctx.Event.Message
	images := ExtractImageURLs(msg)
	imageB64s := make([]string, 0)
	for _, u := range images {
		if b64, err := DownloadImageToBase64(u); err == nil {
			imageB64s = append(imageB64s, b64)
		}
	}

	// 引用消息→续会话
	replyID := ExtractReplyID(msg)
	var sess *ChatSession
	if replyID != 0 {
		key := fmt.Sprintf("%d", replyID)
		if existing, ok := p.sessions.Get(key); ok {
			sess = existing
			p.sessions.Delete(key)
		} else {
			refMsg := ctx.GetMessage(replyID)
			if len(refMsg.Elements) > 0 {
				refText := ExtractText(refMsg.Elements)
				for _, u := range ExtractImageURLs(refMsg.Elements) {
					if b64, err := DownloadImageToBase64(u); err == nil {
						imageB64s = append(imageB64s, b64)
					}
				}
				if refText != "" {
					queryText = fmt.Sprintf("Wait! I am replying to this:\n%s\n\n%s", refText, queryText)
				}
			}
		}
	}

	if sess == nil {
		sess = NewChatSession(p.buildSystemPrompt(ctx, cfg, groupID, userID, queryText, allowTargetSelection))
	}

	// 模型选择：群级覆盖 > 文本 model: 前缀 > cfg 列表
	modelList := cfg.LLM.Models
	if mn := p.fileDB.GetString(fmt.Sprintf("model_%d", groupID)); mn != "" {
		modelList = append([]string{mn}, modelList...)
	}
	if mn, newText := ExtractModelName(queryText); mn != "" {
		modelList = append([]string{mn}, modelList...)
		queryText = newText
	}

	// CleanChat 模式
	if ContainsAny(queryText, CleanChatTriggerWords) {
		queryText = RemoveAll(queryText, CleanChatTriggerWords)
		sess.SystemPrompt = ""
	}

	senderName := ctx.CardOrNickName(userID)
	if senderName == "" {
		senderName = fmt.Sprintf("%d", userID)
	}
	contentWithIdentity := fmt.Sprintf("%s (%d): %s", senderName, userID, queryText)
	sess.AppendUserContent(contentWithIdentity, imageB64s)

	timeStart := time.Now()
	cctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.LLM.Timeout)*time.Second)
	defer cancel()
	resp, err := UniversalChatWithFallback(cctx, sess, modelList, cfg.LLM.MaxTokens)
	if err != nil {
		log.Warn().Err(err).Msg("[autochat] 对话失败")
		return
	}
	elapsed := time.Since(timeStart)

	// 解析 XML 响应
	clean := extractXMLBlock(resp.Result)

	var llmResp AutoChatXMLResponse
	if err := xml.Unmarshal([]byte(clean), &llmResp); err != nil {
		log.Warn().Err(err).Str("raw", truncate(clean, 200)).Msg("[autochat] XML 解析失败，丢弃")
		return
	}

	finalReplies := make([]string, 0, len(llmResp.Replies))
	for _, r := range llmResp.Replies {
		r = strings.TrimSpace(r)
		if r != "" {
			finalReplies = append(finalReplies, r)
		}
	}
	if len(finalReplies) > 10 {
		finalReplies = finalReplies[:10]
	}

	if s := strings.TrimSpace(llmResp.DialogueSummary); s != "" {
		_ = p.memory.AddSummary(groupID, s)
	}
	for _, e := range llmResp.UpdateProfiles {
		text := strings.TrimSpace(e.Value)
		if text == "" {
			continue
		}
		if uid, err := strconv.ParseInt(strings.TrimSpace(e.QQ), 10, 64); err == nil {
			_ = p.memory.UpdateUserMemory(groupID, uid, text)
		}
	}
	if vc := GetVectorClient(); vc != nil && vc.IsEnabled() {
		for _, e := range llmResp.AddMemories {
			text := strings.TrimSpace(e.Value)
			if text == "" {
				continue
			}
			if uid, err := strconv.ParseInt(strings.TrimSpace(e.QQ), 10, 64); err == nil {
				userName := ctx.CardOrNickName(uid)
				go func(g, u int64, n, t string) { _ = vc.UpsertUserMemory(g, u, n, t) }(groupID, uid, userName, text)
			}
		}
	}

	if len(finalReplies) == 0 {
		return
	}
	sess.AppendBotContent(strings.Join(finalReplies, " "))
	p.messageBuffer.Add(groupID, "Bot", ctx.Event.SelfID, strings.Join(finalReplies, " "), time.Now().Unix())

	for i, replyText := range finalReplies {
		if replyText == "" {
			continue
		}
		if i > 0 {
			delay := time.Duration(rand.Float64()*2000) * time.Millisecond
			if i >= 3 {
				delay = time.Duration(rand.Float64()*5000) * time.Millisecond
			}
			time.Sleep(delay)
		}

		var retMsg message.ID
		targetQQs := llmResp.ReplyToQQs
		if allowTargetSelection {
			if len(targetQQs) > 0 {
				elements := make(message.Message, 0)
				if i == 0 {
					seen := map[int64]bool{}
					for _, q := range targetQQs {
						if q != 0 && q != ctx.Event.SelfID && !seen[q] {
							elements = append(elements, message.At(q))
							seen[q] = true
						}
					}
				}
				elements = append(elements, message.Text(" "+replyText))
				retMsg = ctx.SendChain(elements...)
			} else {
				retMsg = ctx.SendChain(message.Text(replyText))
			}
		} else {
			if i == 0 {
				retMsg = ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(replyText))
			} else {
				retMsg = ctx.SendChain(message.Text(replyText))
			}
		}
		if retMsg.ID() != 0 {
			p.sessions.Set(fmt.Sprintf("%d", retMsg.ID()), sess)
		}
	}

	p.tokenStats.Record(resp.PromptTokens, resp.CompletionTokens)
	log.Info().Str("model", resp.Model).Float64("sec", elapsed.Seconds()).
		Int("prompt", resp.PromptTokens).Int("completion", resp.CompletionTokens).
		Msg("[autochat] 回复完成")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// buildSystemPrompt 拼装人设 / 上下文 / RAG / 输出格式约束。
func (p *pluginImpl) buildSystemPrompt(ctx *zero.Ctx, cfg *Config, groupID, userID int64, queryText string, allowTargetSelection bool) string {
	persona := loadPersona(cfg, groupID)
	msg := ctx.Event.Message

	// User memories (本地画像 + 涉及到的相关用户)
	relatedUsers := map[int64]string{userID: ctx.CardOrNickName(userID)}
	for _, atID := range ExtractAtQQ(msg) {
		if atID != ctx.Event.SelfID {
			relatedUsers[atID] = ctx.CardOrNickName(atID)
		}
	}
	for _, m := range p.messageBuffer.GetContext(groupID, cfg.Chat.ContextSize) {
		if _, ok := relatedUsers[m.SenderID]; !ok {
			relatedUsers[m.SenderID] = m.SenderName
		}
	}
	var umText string
	for uid, name := range relatedUsers {
		t, err := p.memory.GetUserMemory(groupID, uid)
		if err == nil && t != "" {
			umText += fmt.Sprintf("- 用户 %s (%d): %s\n", fmtUserName(name, uid), uid, t)
		}
	}
	if umText != "" {
		umText = "[用户画像]\n" + umText
	}

	// RAG memories
	var ragMemText, ragSummaryText string
	if vc := GetVectorClient(); vc != nil && vc.IsEnabled() {
		if recents, err := vc.QueryRecentMemories(groupID, 5); err == nil && len(recents) > 0 {
			ragMemText = "[历史记忆片段]\n"
			for _, m := range recents {
				name := fmtUserName(m.UserName, m.UserID)
				text := m.Text
				if len([]rune(text)) > 50 {
					text = string([]rune(text)[:50]) + "..."
				}
				ragMemText += fmt.Sprintf("- [%s] %s: %s\n", formatTimestamp(m.Timestamp), name, text)
			}
		}
		ragQuery := p.generateRAGSummary(groupID, queryText, cfg)
		if sums, err := vc.QueryRelevantSummaries(groupID, ragQuery, 3); err == nil && len(sums) > 0 {
			ragSummaryText = "[相关历史]\n"
			for _, s := range sums {
				ragSummaryText += "- " + s.Text + "\n"
			}
		}
	}

	// Summary memory
	var smText string
	if sums, err := p.memory.GetRecentSummaries(groupID, 5); err == nil && len(sums) > 0 {
		smText = "[前情提要]\n"
		for i, s := range sums {
			smText += fmt.Sprintf("%d. %s\n", i+1, s)
		}
	}

	// Recent context
	recentMsgs := p.messageBuffer.Get(groupID, cfg.Chat.ContextSize)
	var recentText string
	if len(recentMsgs) > 0 {
		recentText = "[上下文]\n最近群聊（格式：时间 (相对时间) [群号] 昵称(QQ):\\n内容）:\n" + strings.Join(recentMsgs, "\n")
	}
	if !allowTargetSelection {
		curName := ctx.Event.Sender.Card
		if curName == "" {
			curName = ctx.Event.Sender.NickName
		}
		recentText += fmt.Sprintf("\n\n[当前消息]\n当前与你对话的用户是: %s (QQ: %d)\n", curName, userID)
	}

	framework := cfg.Chat.Prompt.Framework
	systemPrompt := framework
	systemPrompt = strings.ReplaceAll(systemPrompt, "{self_id}", fmt.Sprintf("%d", ctx.Event.SelfID))
	systemPrompt = strings.ReplaceAll(systemPrompt, "{self_name}", "bot")
	systemPrompt = strings.ReplaceAll(systemPrompt, "{persona}", persona)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{recent_text}", recentText)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{um_text}", umText)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{sm_text}", smText)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{em_text}", "")
	systemPrompt = strings.ReplaceAll(systemPrompt, "{rag_mem_text}", ragMemText)
	systemPrompt = strings.ReplaceAll(systemPrompt, "{rag_summary_text}", ragSummaryText)

	xmlFormat := `<response>
  <replies>
    <reply>回复1</reply>
    <reply>回复2</reply>
  </replies>
  <dialogue_summary>本次对话简短总结，用于未来的[前情提要]</dialogue_summary>
  <update_profiles>
    <profile qq="123456">覆盖式用户画像描述</profile>
  </update_profiles>
  <add_memories>
    <memory qq="123456">新增的具体事件/细节，将进入RAG</memory>
  </add_memories>`
	if allowTargetSelection {
		xmlFormat += `
  <reply_to_qqs>
    <qq>123456</qq>
  </reply_to_qqs>`
	}
	xmlFormat += `
</response>
说明：
1. 严格输出 XML（不要 Markdown 代码块、不要前后多余文字），根标签必须是 <response>。
2. 单条回复尽量 30 字以内，可拆为多条 <reply>；无回复时 <replies></replies> 留空。
3. 闲聊无新信息时 <update_profiles>/<add_memories> 留空（不要写 <profile>/<memory>）。
4. 文本中如出现 < > & 字符请使用 &lt; &gt; &amp; 转义；除此之外不要做其他转义。`
	systemPrompt += "\n\n# Output Format\n请严格按下列格式输出：\n" + xmlFormat
	return systemPrompt
}

func loadPersona(cfg *Config, groupID int64) string {
	gid := fmt.Sprintf("%d", groupID)
	if v, ok := cfg.Chat.Prompt.Persona[gid]; ok && v != "" {
		return v
	}
	if v, ok := cfg.Chat.Prompt.Persona["default"]; ok {
		return v
	}
	return ""
}

// generateRAGSummary 用便宜模型把当前消息+上下文压缩成检索用 query。
// 未配置时直接降级为 queryText。
func (p *pluginImpl) generateRAGSummary(groupID int64, currentText string, cfg *Config) string {
	if !cfg.RAGSummary.Enabled || cfg.RAGSummary.Model == "" || cfg.RAGSummary.Prompt == "" {
		return currentText
	}
	recents := p.messageBuffer.Get(groupID, 5)
	if len(recents) == 0 {
		return currentText
	}
	prompt := strings.ReplaceAll(cfg.RAGSummary.Prompt, "{text}", strings.Join(recents, "\n"))
	sess := NewChatSession("")
	sess.AppendUserContent(prompt, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.RAGSummary.Timeout)*time.Second)
	defer cancel()
	resp, err := UniversalChat(ctx, sess, cfg.RAGSummary.Model, cfg.RAGSummary.MaxTokens)
	if err != nil {
		return currentText
	}
	out := strings.TrimSpace(resp.Result)
	if out == "" {
		return currentText
	}
	return out
}

// processImageDescription 异步生成图片描述写回 MessageBuffer。
func (p *pluginImpl) processImageDescription(groupID, userID, msgTime int64, images []string, cfg *Config) {
	if !cfg.ImageCaption.Enabled || cfg.ImageCaption.Model == "" {
		return
	}
	prompt := strings.ReplaceAll(cfg.ImageCaption.Prompt, "{sub_type}", "图片")
	timeout := cfg.ImageCaption.Timeout
	if timeout <= 0 {
		timeout = 20
	}
	maxTokens := cfg.ImageCaption.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 80
	}
	var descs []string
	for _, u := range images {
		b64, err := DownloadImageToBase64(u)
		if err != nil {
			continue
		}
		if strings.Contains(b64, "image/gif") {
			continue
		}
		sess := NewChatSession("")
		sess.AppendUserContent(prompt, []string{b64})
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		resp, err := UniversalChat(ctx, sess, cfg.ImageCaption.Model, maxTokens)
		cancel()
		if err != nil {
			continue
		}
		if d := strings.TrimSpace(resp.Result); d != "" {
			descs = append(descs, d)
		}
	}
	if len(descs) > 0 {
		p.messageBuffer.UpdateImageDescs(groupID, userID, msgTime, descs)
	}
}
