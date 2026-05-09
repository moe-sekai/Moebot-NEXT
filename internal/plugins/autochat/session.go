package autochat

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// MessageRole 消息角色
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "model"
	RoleSystem    MessageRole = "system"
)

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
	Images  []string    `json:"images,omitempty"` // base64 编码的图片(data URI 或纯 base64)
}

// ChatSession 聊天会话
type ChatSession struct {
	ID           string        `json:"id"`
	Messages     []ChatMessage `json:"messages"`
	SystemPrompt string        `json:"system_prompt"`
	UpdateTime   time.Time     `json:"update_time"`
	mu           sync.RWMutex
}

func NewChatSession(systemPrompt string) *ChatSession {
	return &ChatSession{
		ID:           uuid.New().String(),
		Messages:     make([]ChatMessage, 0),
		SystemPrompt: systemPrompt,
		UpdateTime:   time.Now(),
	}
}

func (s *ChatSession) AppendUserContent(content string, images []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, ChatMessage{Role: RoleUser, Content: content, Images: images})
	s.UpdateTime = time.Now()
}

func (s *ChatSession) AppendBotContent(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, ChatMessage{Role: RoleAssistant, Content: content})
	s.UpdateTime = time.Now()
}

func (s *ChatSession) Snapshot() []ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ChatMessage, len(s.Messages))
	copy(out, s.Messages)
	return out
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions   map[string]*ChatSession
	expireTime time.Duration
	mu         sync.RWMutex
	stop       chan struct{}
}

func NewSessionManager(expireTime time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions:   make(map[string]*ChatSession),
		expireTime: expireTime,
		stop:       make(chan struct{}),
	}
	go sm.cleanupLoop()
	return sm
}

func (sm *SessionManager) Get(id string) (*ChatSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	return s, ok
}

func (sm *SessionManager) Set(id string, session *ChatSession) {
	sm.mu.Lock()
	sm.sessions[id] = session
	sm.mu.Unlock()
}

func (sm *SessionManager) Delete(id string) {
	sm.mu.Lock()
	delete(sm.sessions, id)
	sm.mu.Unlock()
}

func (sm *SessionManager) Close() {
	select {
	case <-sm.stop:
	default:
		close(sm.stop)
	}
}

func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	for {
		select {
		case <-sm.stop:
			return
		case <-ticker.C:
			sm.cleanup()
		}
	}
}

func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	now := time.Now()
	for id, session := range sm.sessions {
		if now.Sub(session.UpdateTime) > sm.expireTime {
			delete(sm.sessions, id)
		}
	}
}
