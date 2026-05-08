package b30

import (
	"context"
	"strings"
	"sync"
	"time"

	"moebot-next/internal/config"
)

// Client caches the community constants table in memory.
type Client struct {
	url             string
	timeout         time.Duration
	refreshInterval time.Duration

	mu       sync.RWMutex
	table    ConstantsTable
	loadedAt time.Time
}

func NewClient(cfg config.B30Config) *Client {
	url := strings.TrimSpace(cfg.ConstantsURL)
	if url == "" {
		url = config.DefaultB30ConstantsURL
	}
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	refresh := time.Duration(cfg.RefreshInterval) * time.Second
	if refresh <= 0 {
		refresh = 6 * time.Hour
	}
	return &Client{url: url, timeout: timeout, refreshInterval: refresh}
}

func DefaultClient() *Client {
	return NewClient(config.B30Config{ConstantsURL: config.DefaultB30ConstantsURL, Timeout: int(DefaultTimeout / time.Second), RefreshInterval: int((6 * time.Hour) / time.Second)})
}

func (c *Client) URL() string {
	if c == nil || strings.TrimSpace(c.url) == "" {
		return config.DefaultB30ConstantsURL
	}
	return c.url
}

func (c *Client) Get(ctx context.Context) (ConstantsTable, error) {
	if c == nil {
		c = DefaultClient()
	}
	if table, ok := c.cached(); ok {
		return table, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if table, ok := c.cachedLocked(); ok {
		return table, nil
	}
	table, err := LoadConstants(ctx, c.URL(), c.timeout)
	if err != nil {
		if len(c.table.Entries) > 0 {
			return c.table, nil
		}
		return ConstantsTable{}, err
	}
	c.table = table
	c.loadedAt = time.Now()
	return table, nil
}

func (c *Client) cached() (ConstantsTable, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cachedLocked()
}

func (c *Client) cachedLocked() (ConstantsTable, bool) {
	if len(c.table.Entries) == 0 || c.loadedAt.IsZero() {
		return ConstantsTable{}, false
	}
	if c.refreshInterval <= 0 || time.Since(c.loadedAt) < c.refreshInterval {
		return c.table, true
	}
	return ConstantsTable{}, false
}
