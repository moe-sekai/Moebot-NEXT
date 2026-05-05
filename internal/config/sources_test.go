package config

import "testing"

func TestNormalizeRegionAliases(t *testing.T) {
	cases := map[string]string{
		"sc":     RegionCN,
		"jp":     RegionJP,
		"tc":     RegionTW,
		"kr":     RegionKR,
		"global": RegionEN,
		"国服":     RegionCN,
		"日服":     RegionJP,
		"台服":     RegionTW,
		"韩服":     RegionKR,
		"国际服":    RegionEN,
	}
	for input, want := range cases {
		if got := NormalizeRegion(input); got != want {
			t.Fatalf("NormalizeRegion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestResolveMasterdataHarukiAllRegions(t *testing.T) {
	cases := map[string]string{
		RegionJP: "haruki-sekai-master/main/master",
		RegionCN: "haruki-sekai-sc-master/main/master",
		RegionTW: "haruki-sekai-tc-master/main/master",
		RegionKR: "haruki-sekai-kr-master/main/master",
		RegionEN: "haruki-sekai-en-master/main/master",
	}
	for region, wantContains := range cases {
		resolved, err := ResolveMasterdata(MasterdataConfig{Region: region, Source: MasterdataSourceHaruki}, RegionJP)
		if err != nil {
			t.Fatalf("ResolveMasterdata(%s): %v", region, err)
		}
		if resolved.Region != region {
			t.Fatalf("region = %q, want %q", resolved.Region, region)
		}
		if !contains(resolved.URL, wantContains) {
			t.Fatalf("url = %q, want containing %q", resolved.URL, wantContains)
		}
	}
}

func TestResolveMasterdataUnsupportedPresetRegion(t *testing.T) {
	if _, err := ResolveMasterdata(MasterdataConfig{Region: RegionKR, Source: MasterdataSourceMoeSekai}, RegionJP); err == nil {
		t.Fatal("expected MoeSekai KR masterdata to be rejected")
	}
	if _, err := ResolveMasterdata(MasterdataConfig{Region: RegionEN, Source: MasterdataSource8823}, RegionJP); err == nil {
		t.Fatal("expected 8823 EN masterdata to be rejected")
	}
}

func TestResolveMasterdataCustomUsesURLFields(t *testing.T) {
	resolved, err := ResolveMasterdata(MasterdataConfig{
		Region:      "tc",
		Source:      MasterdataSourceCustom,
		CustomURL:   "https://example.test/master/",
		FallbackURL: "https://fallback.test/master/",
	}, RegionJP)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Region != RegionTW {
		t.Fatalf("region = %q, want tw", resolved.Region)
	}
	if resolved.URL != "https://example.test/master" {
		t.Fatalf("url = %q", resolved.URL)
	}
	if resolved.FallbackURL != "https://fallback.test/master" {
		t.Fatalf("fallback = %q", resolved.FallbackURL)
	}
}

func TestResolveAssetsSekaiBestAllRegions(t *testing.T) {
	cases := map[string]string{
		RegionJP: "https://storage.sekai.best/sekai-jp-assets",
		RegionCN: "https://storage.sekai.best/sekai-cn-assets",
		RegionTW: "https://storage.sekai.best/sekai-tc-assets",
		RegionKR: "https://storage.sekai.best/sekai-kr-assets",
		RegionEN: "https://storage.sekai.best/sekai-en-assets",
	}
	for region, want := range cases {
		resolved, err := ResolveAssets(AssetsConfig{Region: region, Source: AssetSourceSekaiBest}, RegionJP)
		if err != nil {
			t.Fatalf("ResolveAssets(%s): %v", region, err)
		}
		if resolved.BaseURL != want {
			t.Fatalf("base = %q, want %q", resolved.BaseURL, want)
		}
	}
}

func TestResolveAssetsMoeSekaiRegionSupportAndRendererKey(t *testing.T) {
	resolved, err := ResolveAssets(AssetsConfig{Region: RegionCN, Source: AssetSourceMoeSekai, Mirror: AssetMirrorBackup}, RegionJP)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.BaseURL != "https://storage2.exmeaning.com/sekai-cn-assets" {
		t.Fatalf("base = %q", resolved.BaseURL)
	}
	if resolved.RendererKey != "backup-cn" {
		t.Fatalf("renderer key = %q", resolved.RendererKey)
	}
	if _, err := ResolveAssets(AssetsConfig{Region: RegionEN, Source: AssetSourceMoeSekai}, RegionJP); err == nil {
		t.Fatal("expected MoeSekai EN assets to be rejected")
	}
}

func TestDefaultGameServerProfiles(t *testing.T) {
	profiles := DefaultGameServerProfiles()
	if !IsEnabled(profiles[RegionJP].Enabled) {
		t.Fatal("default JP profile should be enabled")
	}
	if IsEnabled(profiles[RegionCN].Enabled) {
		t.Fatal("CN profile should default to disabled")
	}
	if profiles[RegionKR].Masterdata.Source != MasterdataSourceHaruki {
		t.Fatalf("KR masterdata source = %q", profiles[RegionKR].Masterdata.Source)
	}
	if profiles[RegionEN].Assets.Source != AssetSourceSekaiBest {
		t.Fatalf("EN assets source = %q", profiles[RegionEN].Assets.Source)
	}
	if profiles[RegionTW].Masterdata.LocalPath != "./data/master/tw" {
		t.Fatalf("TW local path = %q", profiles[RegionTW].Masterdata.LocalPath)
	}
}

func TestGlobalSekaiAPIAppliesToAllGameServerRegions(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Region = RegionJP
	cfg.SekaiAPI.Enabled = true
	cfg.SekaiAPI.BaseURL = "https://sekai.example.test/api/{region}"
	cfg.SekaiAPI.Headers = map[string]string{"x-moe-sekai-token": "secret-token"}
	cfg.GameServers = map[string]GameServerConfig{
		RegionCN: {
			Enabled: EnabledPtr(true),
			SekaiAPI: SekaiAPIConfig{
				BaseURL: DefaultSekaiAPIURL,
				Region:  RegionCN,
				Headers: map[string]string{},
			},
		},
	}
	NormalizeConfig(cfg)

	cn := ResolveGameServerProfile(cfg, RegionCN)
	if !cn.SekaiAPI.Enabled {
		t.Fatal("cn sekai api should inherit global enabled flag")
	}
	if cn.SekaiAPI.BaseURL != "https://sekai.example.test/api/{region}" {
		t.Fatalf("cn sekai api base url = %q", cn.SekaiAPI.BaseURL)
	}
	if cn.SekaiAPI.Headers["x-moe-sekai-token"] != "secret-token" {
		t.Fatalf("cn sekai api headers = %+v", cn.SekaiAPI.Headers)
	}
}

func TestRankingRegionAlwaysFollowsGameServerRegion(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Server.Region = RegionJP
	cfg.RankingAPI.Region = RegionCN
	cfg.GameServers = map[string]GameServerConfig{
		RegionCN: {
			Enabled: EnabledPtr(true),
			RankingAPI: RankingAPIConfig{
				Region: RegionJP,
			},
		},
	}
	NormalizeConfig(cfg)

	jp := ResolveGameServerProfile(cfg, RegionJP)
	if jp.RankingAPI.Region != RegionJP {
		t.Fatalf("jp ranking region = %q, want %q", jp.RankingAPI.Region, RegionJP)
	}
	cn := ResolveGameServerProfile(cfg, RegionCN)
	if cn.RankingAPI.Region != RegionCN {
		t.Fatalf("cn ranking region = %q, want %q", cn.RankingAPI.Region, RegionCN)
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
