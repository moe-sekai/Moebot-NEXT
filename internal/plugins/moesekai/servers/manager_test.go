package servers

import (
	"testing"

	"moebot-next/internal/config"
)

func TestManagerGetExactDoesNotFallbackForDisabledRegion(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Server.Region = config.RegionJP
	manager := NewManager(cfg)

	exact := manager.GetExact(config.RegionTW)
	if exact == nil {
		t.Fatal("GetExact(tw) returned nil")
	}
	if exact.Region != config.RegionTW {
		t.Fatalf("exact region = %q, want tw", exact.Region)
	}
	if exact.Enabled {
		t.Fatal("tw should be disabled by default in exact lookup")
	}

	fallback := manager.Get(config.RegionTW)
	if fallback == nil {
		t.Fatal("Get(tw) returned nil")
	}
	if fallback.Region == config.RegionTW {
		t.Fatalf("Get(tw) should fallback when disabled, got %q", fallback.Region)
	}
}

func TestManagerGetExactRejectsInvalidRegion(t *testing.T) {
	manager := NewManager(config.DefaultConfig())
	if runtime := manager.GetExact("unknown"); runtime != nil {
		t.Fatalf("GetExact(unknown) = %#v, want nil", runtime)
	}
}
