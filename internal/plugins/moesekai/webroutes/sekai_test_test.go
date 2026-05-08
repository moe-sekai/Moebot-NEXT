package webroutes

import (
	"testing"

	"moebot-next/internal/config"
)

func TestSekaiSystemURLSupportsRegionPlaceholder(t *testing.T) {
	got, err := sekaiSystemURL("https://seka-api.exmeaning.com/api/{region}", config.RegionCN)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://seka-api.exmeaning.com/api/cn/system" {
		t.Fatalf("system url = %q", got)
	}

	got, err = sekaiSystemURL("https://seka-api.exmeaning.com", config.RegionJP)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://seka-api.exmeaning.com/api/jp/system" {
		t.Fatalf("default system url = %q", got)
	}
}
