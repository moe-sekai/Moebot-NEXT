package filter

import (
	"encoding/json"
	"slices"
	"strings"
	"sync"

	regexp "github.com/dlclark/regexp2"
	"github.com/rs/zerolog/log"
)

// Filter is the compiled, runtime view of a downstream FilterApp's rules.
type Filter struct {
	Name           string
	UserID         IDFilter
	GroupID        IDFilter
	PrivateMessage MessageFilter
	GroupMessage   MessageFilter

	publisher func(Event)
	mu        sync.RWMutex
}

// SetPublisher attaches an event publisher; safe to call before/after Compile.
func (f *Filter) SetPublisher(p func(Event)) {
	f.mu.Lock()
	f.publisher = p
	f.mu.Unlock()
}

func (f *Filter) emit(ev Event) {
	if f.publisher == nil {
		return
	}
	ev.Filter = f.Name
	f.publisher(ev)
}

// IDFilter wraps an IDRule with no extra state for now.
type IDFilter struct {
	IDRule
}

// MessageFilter holds compiled regexps along with the textual rule.
type MessageFilter struct {
	MessageRule
	regexps []*regexp.Regexp
}

// CompiledRules is what the manager passes when (re)building a Filter.
type CompiledRules struct {
	Name           string
	UserID         IDRule
	GroupID        IDRule
	Message        MessageRule
	PrivateMessage MessageRule
	GroupMessage   MessageRule
	DefaultUserID  IDRule
	DefaultGroupID IDRule
}

// Compile rebuilds the filter's runtime state from raw rules. It applies the
// "default" fallback semantics from the original OneBotFilter project.
func (f *Filter) Compile(c CompiledRules) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Name = c.Name

	userID := c.UserID
	if userID.Mode == "" || userID.Mode == ModeDefault {
		userID = c.DefaultUserID
	}
	groupID := c.GroupID
	if groupID.Mode == "" || groupID.Mode == ModeDefault {
		groupID = c.DefaultGroupID
	}
	f.UserID = IDFilter{IDRule: userID}
	f.GroupID = IDFilter{IDRule: groupID}

	private := c.PrivateMessage
	if private.Mode == "" || private.Mode == ModeDefault {
		private = c.Message
	}
	group := c.GroupMessage
	if group.Mode == "" || group.Mode == ModeDefault {
		group = c.Message
	}
	f.PrivateMessage = compileMessage(private)
	f.GroupMessage = compileMessage(group)
}

func compileMessage(rule MessageRule) MessageFilter {
	mf := MessageFilter{MessageRule: rule}
	for _, pat := range rule.Filters {
		if pat == "" {
			continue
		}
		re, err := regexp.Compile(pat, regexp.None)
		if err != nil {
			log.Warn().Str("pattern", pat).Err(err).Msg("filter: invalid regex")
			continue
		}
		mf.regexps = append(mf.regexps, re)
	}
	return mf
}

// Allow checks whether an OneBot message should be forwarded. It may rewrite
// the message in-place when prefix-replace is configured.
func (f *Filter) Allow(msg *OneBotMessage, debug bool) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	base := Event{
		UserID:  msg.Partial.UserID,
		GroupID: msg.Partial.GroupID,
		MsgType: msg.Partial.MessageType,
		Raw:     msg.Partial.RawMessage,
	}
	var rule *MessageFilter
	switch msg.Partial.MessageType {
	case MessageTypeGroup:
		if !f.GroupID.match(msg.Partial.GroupID) {
			if debug {
				log.Debug().Str("filter", f.Name).Int64("group_id", msg.Partial.GroupID).Msg("filter: group blocked")
			}
			b := base
			b.Kind = EventBlock
			b.Reason = "group_id"
			f.emit(b)
			return false
		}
		rule = &f.GroupMessage
		fallthrough
	case MessageTypePrivate:
		if !f.UserID.match(msg.Partial.UserID) {
			if debug {
				log.Debug().Str("filter", f.Name).Int64("user_id", msg.Partial.UserID).Msg("filter: user blocked")
			}
			b := base
			b.Kind = EventBlock
			b.Reason = "user_id"
			f.emit(b)
			return false
		}
		if rule == nil {
			rule = &f.PrivateMessage
		}
	default:
		return true
	}

	if rule == nil || rule.Mode == "" || rule.Mode == ModeOn {
		a := base
		a.Kind = EventAllow
		a.Reason = "on"
		f.emit(a)
		return true
	}
	if rule.Mode == ModeOff {
		if debug {
			log.Debug().Str("filter", f.Name).Str("raw", msg.Partial.RawMessage).Msg("filter: dropped by off")
		}
		b := base
		b.Kind = EventBlock
		b.Reason = "off"
		f.emit(b)
		return false
	}
	if rule.prefixPass(msg) {
		log.Info().Str("filter", f.Name).Str("raw", msg.Partial.RawMessage).Msg("filter: prefix passthrough")
		a := base
		a.Kind = EventPrefixPass
		a.Reason = "prefix"
		a.Raw = msg.Partial.RawMessage
		f.emit(a)
		return true
	}

	matched := false
	if msg.Partial.MessageFormat == MessageFormatArray {
		for _, seg := range msg.Partial.MessageArray {
			if seg.Type != MessageContentText {
				continue
			}
			text, _ := seg.Data["text"].(string)
			if rule.matchesText(strings.TrimSpace(text)) {
				matched = true
				break
			}
		}
	} else {
		matched = rule.matchesText(strings.TrimSpace(msg.Partial.MessageString))
	}

	switch rule.Mode {
	case ModeWhitelist:
		if matched {
			a := base
			a.Kind = EventAllow
			a.Reason = "whitelist_hit"
			f.emit(a)
			return true
		}
		if debug {
			log.Debug().Str("filter", f.Name).Str("raw", msg.Partial.RawMessage).Msg("filter: not in whitelist")
		}
		b := base
		b.Kind = EventBlock
		b.Reason = "whitelist_miss"
		f.emit(b)
		return false
	case ModeBlacklist:
		if matched {
			if debug {
				log.Debug().Str("filter", f.Name).Str("raw", msg.Partial.RawMessage).Msg("filter: hit blacklist")
			}
			b := base
			b.Kind = EventBlock
			b.Reason = "blacklist_hit"
			f.emit(b)
			return false
		}
		a := base
		a.Kind = EventAllow
		a.Reason = "blacklist_miss"
		f.emit(a)
		return true
	}
	log.Warn().Str("filter", f.Name).Str("mode", rule.Mode).Msg("filter: invalid message mode, blocking")
	return false
}

func (idf *IDFilter) match(id int64) bool {
	if id == 0 {
		return true
	}
	switch idf.Mode {
	case "", ModeOn:
		return true
	case ModeOff:
		return false
	case ModeWhitelist:
		return slices.Contains(idf.IDs, id)
	case ModeBlacklist:
		return !slices.Contains(idf.IDs, id)
	}
	return true
}

func (mf *MessageFilter) matchesText(text string) bool {
	for _, re := range mf.regexps {
		ok, err := re.MatchString(text)
		if err != nil {
			log.Warn().Str("pattern", re.String()).Err(err).Msg("filter: regex match error")
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

func (mf *MessageFilter) prefixPass(msg *OneBotMessage) bool {
	if mf == nil || len(mf.Prefix) == 0 {
		return false
	}
	var textOld string
	var index int
	switch msg.Partial.MessageFormat {
	case MessageFormatArray:
		for i, seg := range msg.Partial.MessageArray {
			if seg.Type != MessageContentText {
				continue
			}
			t, _ := seg.Data["text"].(string)
			textOld = strings.TrimSpace(t)
			index = i
			break
		}
	case MessageFormatString:
		textOld = strings.TrimSpace(msg.Partial.MessageString)
	default:
		return false
	}
	if textOld == "" {
		return false
	}
	prefix := ""
	for _, p := range mf.Prefix {
		if p == "" {
			continue
		}
		if strings.HasPrefix(textOld, p) {
			prefix = p
			break
		}
	}
	if prefix == "" {
		return false
	}
	text := mf.PrefixReplace + textOld[len(prefix):]
	switch msg.Partial.MessageFormat {
	case MessageFormatArray:
		msg.Partial.MessageArray[index].Data["text"] = text
		if strings.TrimSpace(text) == "" {
			msg.Partial.MessageArray = append(msg.Partial.MessageArray[:index], msg.Partial.MessageArray[index+1:]...)
		}
		b, err := json.Marshal(msg.Partial.MessageArray)
		if err != nil {
			log.Warn().Err(err).Msg("filter: re-marshal message array failed")
			return false
		}
		msg.Intact["message"] = b
	case MessageFormatString:
		msg.Partial.MessageString = text
		b, err := json.Marshal(msg.Partial.MessageString)
		if err != nil {
			log.Warn().Err(err).Msg("filter: re-marshal message string failed")
			return false
		}
		msg.Intact["message"] = b
	}
	msg.Partial.RawMessage = strings.Replace(msg.Partial.RawMessage, textOld, text, 1)
	if b, err := json.Marshal(msg.Partial.RawMessage); err == nil {
		msg.Intact["raw_message"] = b
	}
	return true
}
