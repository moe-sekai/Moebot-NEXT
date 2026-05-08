package assets

import (
	"strings"
	"testing"

	"moebot-next/internal/config"
)

func TestImageAssetURLsUsePNG(t *testing.T) {
	urls := []string{
		GetCardThumbnailURL("res001_no001", false),
		GetCardFullURL("res001_no001", true),
		GetMusicJacketURL("jacket_s_001"),
		GetEventBannerURL("event_test"),
		GetEventLogoURL("event_test"),
		GetGachaLogoURL("gacha_test"),
		GetAttrIconURL(Attribute("cute")),
		GetUnitLogoURL(UnitID("piapro")),
		GetHonorBgURL("honor_test", "main"),
		GetHonorFrameURL("high", "main"),
		GetHonorLevelIconURL(false),
		GetHonorLevelIconURL(true),
	}
	for _, url := range urls {
		if strings.Contains(url, ".webp") {
			t.Fatalf("image URL still uses webp: %s", url)
		}
		if !strings.Contains(url, ".png") {
			t.Fatalf("image URL does not use png: %s", url)
		}
	}
}

func TestResolverImageAssetURLsUsePNG(t *testing.T) {
	resolver := DefaultResolver()
	urls := []string{
		resolver.GetCardThumbnailURL("res001_no001", false),
		resolver.GetCardFullURL("res001_no001", true),
		resolver.GetMusicJacketURL("jacket_s_001"),
		resolver.GetEventBannerURL("event_test"),
		resolver.GetEventLogoURL("event_test"),
		resolver.GetGachaLogoURL("gacha_test"),
		resolver.GetAttrIconURL(Attribute("cute")),
		resolver.GetUnitLogoURL(UnitID("piapro")),
		resolver.GetHonorBgURL("honor_test", "main"),
		resolver.GetHonorFrameURL("high", "main"),
		resolver.GetHonorLevelIconURL(false),
		resolver.GetHonorLevelIconURL(true),
	}
	for _, url := range urls {
		if strings.Contains(url, ".webp") {
			t.Fatalf("resolver image URL still uses webp: %s", url)
		}
		if !strings.Contains(url, ".png") {
			t.Fatalf("resolver image URL does not use png: %s", url)
		}
	}
}

func TestResolverCardThumbnailsUseJPAssetsAcrossRegions(t *testing.T) {
	cases := []struct {
		name   string
		cfg    config.AssetsConfig
		region string
	}{
		{
			name:   "moesekai cn mirror",
			cfg:    config.AssetsConfig{Source: config.AssetSourceMoeSekai, Region: config.RegionCN, Mirror: config.AssetMirrorMain},
			region: config.RegionCN,
		},
		{
			name:   "sekai best tw",
			cfg:    config.AssetsConfig{Source: config.AssetSourceSekaiBest, Region: config.RegionTW},
			region: config.RegionTW,
		},
		{
			name:   "sekai best kr",
			cfg:    config.AssetsConfig{Source: config.AssetSourceSekaiBest, Region: config.RegionKR},
			region: config.RegionKR,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resolver, err := NewResolver(tc.cfg, tc.region)
			if err != nil {
				t.Fatalf("NewResolver failed: %v", err)
			}
			thumbnail := resolver.GetCardThumbnailURL("res001_no001", false)
			if !strings.Contains(thumbnail, "/sekai-jp-assets/thumbnail/chara/") {
				t.Fatalf("thumbnail did not use JP assets: %s", thumbnail)
			}
			full := resolver.GetCardFullURL("res001_no001", false)
			if strings.Contains(full, "/sekai-jp-assets/") && tc.region != config.RegionJP {
				t.Fatalf("non-thumbnail asset was unexpectedly normalized to JP: %s", full)
			}
		})
	}
}
