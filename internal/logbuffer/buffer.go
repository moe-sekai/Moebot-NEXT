// Package logbuffer provides a thread-safe ring buffer that captures
// zerolog JSON log lines for in-memory inspection via the admin web UI.
package logbuffer

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

const (
	defaultCapacity = 2000
	minCapacity     = 200
	maxCapacity     = 50000
)

// Entry represents a single captured log line.
type Entry struct {
	Seq     uint64                 `json:"seq"`
	Time    time.Time              `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// FilterOpts controls Snapshot filtering.
type FilterOpts struct {
	Levels   []string // lower-case; empty = all
	Query    string   // case-insensitive substring on message + serialised fields
	SinceSeq uint64   // only return entries with Seq > SinceSeq (0 = all)
	Limit    int      // max entries returned (0 = no extra cap; safety cap = capacity)
}

// Buffer is a thread-safe ring buffer of log entries that implements io.Writer.
type Buffer struct {
	mu      sync.RWMutex
	entries []Entry
	head    int    // next write position
	size    int    // current count
	cap     int    // capacity
	nextSeq uint64 // monotonically increasing sequence
	dropped uint64 // entries that have been overwritten since boot
}

// New creates a new ring buffer. Capacity is clamped to a sane range.
func New(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity < minCapacity {
		capacity = minCapacity
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}
	return &Buffer{
		entries: make([]Entry, capacity),
		cap:     capacity,
	}
}

// Capacity returns the maximum number of entries the buffer can hold.
func (b *Buffer) Capacity() int {
	if b == nil {
		return 0
	}
	return b.cap
}

// Stats returns aggregate counters.
func (b *Buffer) Stats() (total int, dropped uint64, nextSeq uint64) {
	if b == nil {
		return 0, 0, 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size, b.dropped, b.nextSeq
}

// Write parses a zerolog JSON line and appends an entry. Failed parses are
// stored as a raw message so nothing is lost.
func (b *Buffer) Write(p []byte) (int, error) {
	if b == nil {
		return len(p), nil
	}
	entry := parseLine(p)
	b.mu.Lock()
	b.nextSeq++
	entry.Seq = b.nextSeq
	if b.size == b.cap {
		b.dropped++
	} else {
		b.size++
	}
	b.entries[b.head] = entry
	b.head = (b.head + 1) % b.cap
	b.mu.Unlock()
	return len(p), nil
}

// Snapshot returns matching entries in newest-first order. The second return
// value is the total number of entries in the buffer (pre-filter).
func (b *Buffer) Snapshot(opts FilterOpts) ([]Entry, int) {
	if b == nil {
		return nil, 0
	}
	b.mu.RLock()
	defer b.mu.RUnlock()

	total := b.size
	if total == 0 {
		return []Entry{}, 0
	}

	limit := opts.Limit
	if limit <= 0 || limit > b.cap {
		limit = b.cap
	}

	levels := normalisedLevelSet(opts.Levels)
	query := strings.ToLower(strings.TrimSpace(opts.Query))

	out := make([]Entry, 0, limit)
	// iterate newest -> oldest
	for i := 0; i < b.size; i++ {
		idx := (b.head - 1 - i + b.cap) % b.cap
		entry := b.entries[idx]
		if entry.Seq <= opts.SinceSeq {
			continue
		}
		if levels != nil {
			if _, ok := levels[strings.ToLower(entry.Level)]; !ok {
				continue
			}
		}
		if query != "" && !matchQuery(entry, query) {
			continue
		}
		out = append(out, entry)
		if len(out) >= limit {
			break
		}
	}
	return out, total
}

func parseLine(p []byte) Entry {
	trim := strings.TrimRight(string(p), "\r\n")
	entry := Entry{Time: time.Now()}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trim), &raw); err != nil {
		entry.Level = "info"
		entry.Message = trim
		return entry
	}
	fields := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		switch key {
		case "level":
			var s string
			if err := json.Unmarshal(value, &s); err == nil {
				entry.Level = s
			}
		case "time":
			entry.Time = parseTime(value)
		case "message":
			var s string
			if err := json.Unmarshal(value, &s); err == nil {
				entry.Message = s
			}
		default:
			var v interface{}
			if err := json.Unmarshal(value, &v); err == nil {
				fields[key] = v
			} else {
				fields[key] = string(value)
			}
		}
	}
	if entry.Level == "" {
		entry.Level = "info"
	}
	if len(fields) > 0 {
		entry.Fields = fields
	}
	return entry
}

func parseTime(value json.RawMessage) time.Time {
	var s string
	if err := json.Unmarshal(value, &s); err == nil {
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, time.DateTime} {
			if t, err := time.Parse(layout, s); err == nil {
				return t
			}
		}
	}
	var n int64
	if err := json.Unmarshal(value, &n); err == nil {
		return time.Unix(n, 0)
	}
	return time.Now()
}

func normalisedLevelSet(levels []string) map[string]struct{} {
	if len(levels) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(levels))
	for _, level := range levels {
		level = strings.ToLower(strings.TrimSpace(level))
		if level == "" {
			continue
		}
		out[level] = struct{}{}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func matchQuery(entry Entry, query string) bool {
	if strings.Contains(strings.ToLower(entry.Message), query) {
		return true
	}
	if len(entry.Fields) == 0 {
		return false
	}
	if data, err := json.Marshal(entry.Fields); err == nil {
		if strings.Contains(strings.ToLower(string(data)), query) {
			return true
		}
	}
	return false
}
