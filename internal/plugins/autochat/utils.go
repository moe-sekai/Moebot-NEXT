package autochat

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wdvxdr1123/ZeroBot/message"
)

func currentTimestamp() int64 { return time.Now().Unix() }
func currentDate() string     { return time.Now().Format("2006-01-02") }
func formatDate(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02")
}
func formatInt64(n int64) string { return strconv.FormatInt(n, 10) }
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Truncate 按 rune 截断
func Truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}

// ExtractText 取消息纯文本
func ExtractText(msg message.Message) string {
	var b strings.Builder
	for _, seg := range msg {
		if seg.Type == "text" {
			b.WriteString(seg.Data["text"])
		}
	}
	return b.String()
}

// ExtractImageURLs 取消息中的图片 URL
func ExtractImageURLs(msg message.Message) []string {
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

// ExtractAtQQ 提取 @ 的 QQ 号列表
func ExtractAtQQ(msg message.Message) []int64 {
	var qqs []int64
	for _, seg := range msg {
		if seg.Type == "at" {
			if qq, ok := seg.Data["qq"]; ok {
				if id, err := strconv.ParseInt(qq, 10, 64); err == nil {
					qqs = append(qqs, id)
				}
			}
		}
	}
	if len(qqs) == 0 {
		re := regexp.MustCompile(`\[CQ:at,qq=(\d+)[^\]]*\]`)
		for _, m := range re.FindAllStringSubmatch(msg.String(), -1) {
			if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
				qqs = append(qqs, id)
			}
		}
	}
	return qqs
}

// ExtractReplyID 提取回复消息 ID
func ExtractReplyID(msg message.Message) int64 {
	for _, seg := range msg {
		if seg.Type == "reply" {
			if id, ok := seg.Data["id"]; ok {
				if v, err := parseInt64(id); err == nil {
					return v
				}
			}
		}
	}
	return 0
}

// HasAt 是否 @ 了指定 QQ
func HasAt(msg message.Message, qq int64) bool {
	for _, id := range ExtractAtQQ(msg) {
		if id == qq {
			return true
		}
	}
	return false
}

// 工具函数
var CleanChatTriggerWords = []string{"cleanchat", "clean_chat", "cleanmode", "clean_mode"}

func ContainsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func RemoveAll(s string, subs []string) string {
	for _, sub := range subs {
		s = strings.ReplaceAll(s, sub, "")
	}
	return s
}

func ExtractModelName(text string) (string, string) {
	re := regexp.MustCompile(`model:(\S+)`)
	m := re.FindStringSubmatch(text)
	if len(m) > 1 {
		return m[1], strings.TrimSpace(strings.Replace(text, "model:"+m[1], "", 1))
	}
	return "", text
}

func loadsJSON(s string) (map[string]any, error) {
	var r map[string]any
	err := json.Unmarshal([]byte(s), &r)
	return r, err
}

func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "未知时间"
	}
	return time.Unix(ts, 0).Format("01-02 15:04")
}

// extractPureText 拼接 text 段，跳过 at-bot 段，得到“真正的用户输入文本”，
// 用于判断是否是其它插件的命令。其它 @人 的 at 段保留为 " @qq "（与 ZeroBot 命令注册处的写法一致）。
func extractPureText(msg message.Message, selfID int64) string {
	selfStr := strconv.FormatInt(selfID, 10)
	var b strings.Builder
	for _, seg := range msg {
		switch seg.Type {
		case "text":
			b.WriteString(seg.Data["text"])
		case "at":
			if seg.Data["qq"] == selfStr {
				continue
			}
			if qq, ok := seg.Data["qq"]; ok {
				b.WriteString(" @" + qq + " ")
			}
		}
	}
	return b.String()
}

// ignorePatternCache 缓存编译后的正则，避免每条消息重复编译。
var (
	ignorePatternMu    sync.RWMutex
	ignorePatternCache = map[string]*regexp.Regexp{}
)

func compileIgnorePattern(p string) *regexp.Regexp {
	ignorePatternMu.RLock()
	re, ok := ignorePatternCache[p]
	ignorePatternMu.RUnlock()
	if ok {
		return re
	}
	compiled, err := regexp.Compile(p)
	if err != nil {
		return nil
	}
	ignorePatternMu.Lock()
	ignorePatternCache[p] = compiled
	ignorePatternMu.Unlock()
	return compiled
}

// isIgnoredCommand 判断 pureText 是否应被视作其它插件的命令而跳过 autochat。
// 规则：去掉首尾空白后，若以任一 IgnorePrefixes 开头，或匹配任一 IgnorePatterns 正则，则返回 true。
func isIgnoredCommand(pureText string) bool {
	t := strings.TrimSpace(pureText)
	if t == "" {
		return false
	}
	c := GetConfig()
	if c == nil {
		// 与默认值保持一致
		return strings.ContainsAny(string([]rune(t)[0:1]), "/#!！.。>&")
	}
	for _, p := range c.Chat.IgnorePrefixes {
		if p == "" {
			continue
		}
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	for _, p := range c.Chat.IgnorePatterns {
		if p == "" {
			continue
		}
		if re := compileIgnorePattern(p); re != nil && re.MatchString(t) {
			return true
		}
	}
	return false
}

// fmtUserName 兼容空昵称
func fmtUserName(name string, id int64) string {
	if name == "" {
		return fmt.Sprintf("%d", id)
	}
	return name
}
