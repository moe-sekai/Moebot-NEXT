package commands

import (
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"moebot-next/internal/config"
	"moebot-next/internal/plugins/moesekai/servers"
)

func runtimeFixture() *servers.Runtime {
	return &servers.Runtime{Region: config.RegionCN, Label: "国服", Enabled: true}
}

type fakeNetError struct{ timeout bool }

func (f fakeNetError) Error() string   { return "fake net error" }
func (f fakeNetError) Timeout() bool   { return f.timeout }
func (f fakeNetError) Temporary() bool { return false }

var _ net.Error = fakeNetError{}

func TestBindCommandHint(t *testing.T) {
	cases := map[string]string{
		"":              "/绑定",
		config.RegionJP: "/绑定",
		config.RegionCN: "/cn绑定",
		config.RegionTW: "/tw绑定",
	}
	for region, want := range cases {
		if got := bindCommandHint(region); got != want {
			t.Errorf("bindCommandHint(%q) = %q, want %q", region, got, want)
		}
	}
}

func TestFriendlySuiteError(t *testing.T) {
	rt := runtimeFixture()
	cases := []struct {
		name    string
		err     error
		wantHas string
	}{
		{"disabled", errors.New("suite api is disabled"), "暂未配置"},
		{"empty uid", errors.New("uid is empty"), "请先绑定"},
		{"404 not uploaded", errors.New("suite request returned 404"), "尚未在 Haruki Suite 上传"},
		{"403 forbidden", errors.New("suite request returned 403"), "鉴权失败"},
		{"429 rate limit", errors.New("suite request returned 429"), "请求过于频繁"},
		{"500 server", errors.New("suite request returned 500"), "服务异常"},
		{"503 server", errors.New("suite request returned 503"), "服务异常"},
		{"timeout text", errors.New("Get \"http://x\": context deadline exceeded"), "超时"},
		{"net timeout", fakeNetError{timeout: true}, "无法连接"},
		{"connection refused", errors.New("dial tcp: connection refused"), "无法连接"},
		{"decode", errors.New("decode suite response: unexpected EOF"), "解析"},
		{"parse url", errors.New("parse suite url: bad"), "配置异常"},
		{"unknown", errors.New("totally unknown failure"), "请稍后重试"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := friendlySuiteError(rt, "测试", tc.err)
			if !strings.Contains(got, tc.wantHas) {
				t.Errorf("got %q, want substring %q", got, tc.wantHas)
			}
			// Must never leak raw err string verbatim.
			if strings.Contains(got, "returned ") || strings.Contains(got, "tcp:") {
				t.Errorf("raw error leaked: %q", got)
			}
		})
	}
	if got := friendlySuiteError(rt, "测试", nil); got != "" {
		t.Errorf("nil err should yield empty message, got %q", got)
	}
}

func TestFriendlySekaiError(t *testing.T) {
	rt := runtimeFixture()
	cases := []struct {
		name    string
		err     error
		wantHas string
	}{
		{"disabled", errors.New("sekai api is disabled"), "暂未配置"},
		{"empty uid", errors.New("user id is empty"), "请先绑定"},
		{"404", errors.New("profile request returned 404"), "未在"},
		{"401", errors.New("profile request returned 401"), "鉴权失败"},
		{"500", errors.New("profile request returned 500"), "服务异常"},
		{"timeout", errors.New("Get: context deadline exceeded"), "超时"},
		{"network", errors.New("profile request failed: dial tcp"), "无法连接"},
		{"unknown", errors.New("???"), "请稍后重试"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := friendlySekaiError(rt, "个人信息", tc.err)
			if !strings.Contains(got, tc.wantHas) {
				t.Errorf("got %q, want substring %q", got, tc.wantHas)
			}
		})
	}
}

func TestIsStatus5xx(t *testing.T) {
	if !isStatus5xx("suite request returned 502") {
		t.Error("502 should be 5xx")
	}
	if isStatus5xx("suite request returned 404") {
		t.Error("404 should not be 5xx")
	}
}

func TestIsNetworkError(t *testing.T) {
	if !isNetworkError(fakeNetError{}) {
		t.Error("expected net.Error to be detected")
	}
	if isNetworkError(errors.New("plain")) {
		t.Error("plain error should not be network")
	}
}

// guard against accidentally taking unexpectedly long during error formatting.
func TestFriendlySuiteErrorIsFast(t *testing.T) {
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_ = friendlySuiteError(runtimeFixture(), "测试", errors.New("suite request returned 500"))
	}
	if time.Since(start) > time.Second {
		t.Errorf("friendlySuiteError too slow")
	}
}
