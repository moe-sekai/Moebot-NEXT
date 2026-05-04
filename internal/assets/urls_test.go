package assets

import (
	"strings"
	"testing"
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
