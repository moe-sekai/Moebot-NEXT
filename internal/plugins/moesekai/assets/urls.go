package assets

import (
	"fmt"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// CDN source management
// ---------------------------------------------------------------------------

// CDNSource identifies which CDN mirror to use.
type CDNSource string

const (
	CDNCNMain          CDNSource = "cn_main"
	CDNCNBackup        CDNSource = "cn_backup"
	CDNOverseas        CDNSource = "overseas"
	CDNOverseasBackup  CDNSource = "overseas_backup"
	legacyCDNCN        CDNSource = "cn"
	legacyCDNMainJP    CDNSource = "main-jp"
	legacyCDNBackupJP  CDNSource = "backup-jp"
	legacyCDNMainCN    CDNSource = "main-cn"
	legacyCDNBackupCN  CDNSource = "backup-cn"
	legacyCDNOverseasJ CDNSource = "overseas-jp"
)

// cdnBaseURLs maps each CDN source to its base URL (no trailing slash).
var cdnBaseURLs = map[CDNSource]string{
	CDNCNMain:          "https://storage.exmeaning.com/sekai-jp-assets",
	legacyCDNCN:        "https://storage.exmeaning.com/sekai-jp-assets",
	CDNCNBackup:        "https://storage2.exmeaning.com/sekai-jp-assets",
	CDNOverseas:        "https://storage.pjsk.moe/sekai-jp-assets",
	CDNOverseasBackup:  "https://storage2.pjsk.moe/sekai-jp-assets",
	legacyCDNMainJP:    "https://storage.exmeaning.com/sekai-jp-assets",
	legacyCDNBackupJP:  "https://storage2.exmeaning.com/sekai-jp-assets",
	legacyCDNOverseasJ: "https://storage.pjsk.moe/sekai-jp-assets",
	legacyCDNMainCN:    "https://storage.exmeaning.com/sekai-cn-assets",
	legacyCDNBackupCN:  "https://storage2.exmeaning.com/sekai-cn-assets",
}

var legacyRendererSources = map[CDNSource]string{
	CDNCNMain:          "main-jp",
	legacyCDNCN:        "main-jp",
	CDNCNBackup:        "backup-jp",
	CDNOverseas:        "overseas-jp",
	CDNOverseasBackup:  "overseas-backup-jp",
	legacyCDNMainJP:    "main-jp",
	legacyCDNBackupJP:  "backup-jp",
	legacyCDNOverseasJ: "overseas-jp",
	legacyCDNMainCN:    "main-cn",
	legacyCDNBackupCN:  "backup-cn",
}

// StaticBase is the base URL for static / locally-hosted assets.
const StaticBase = "https://moe.exmeaning.com/assets"

var (
	cdnMu               sync.RWMutex
	cdnSource           CDNSource = CDNCNMain
	cdnBaseURL          string    = cdnBaseURLs[CDNCNMain]
	rendererAssetSource string    = legacyRendererSources[CDNCNMain]
)

// GetCDNSource returns the currently active CDN source.
func GetCDNSource() CDNSource {
	cdnMu.RLock()
	defer cdnMu.RUnlock()
	return cdnSource
}

// GetCDNBaseURL returns the currently active asset base URL.
func GetCDNBaseURL() string {
	return cdnBase()
}

// GetRendererAssetSource returns the renderer-side source key or custom base URL.
func GetRendererAssetSource() string {
	cdnMu.RLock()
	defer cdnMu.RUnlock()
	return rendererAssetSource
}

// SetCDNSource switches the active CDN source.
func SetCDNSource(src CDNSource) {
	cdnMu.Lock()
	defer cdnMu.Unlock()
	if baseURL, ok := cdnBaseURLs[src]; ok {
		cdnSource = src
		cdnBaseURL = baseURL
		rendererAssetSource = legacyRendererSources[src]
	}
}

// cdnBase returns the base URL for the currently selected CDN.
func cdnBase() string {
	cdnMu.RLock()
	defer cdnMu.RUnlock()
	if cdnBaseURL != "" {
		return cdnBaseURL
	}
	return cdnBaseURLs[cdnSource]
}

// ---------------------------------------------------------------------------
// URL Builders
// ---------------------------------------------------------------------------

// trainingSuffix returns the file name suffix for trained / untrained cards.
func trainingSuffix(trained bool) string {
	if trained {
		return "after_training"
	}
	return "normal"
}

func honorRarityNumber(rarity string) int {
	switch rarity {
	case "middle":
		return 2
	case "high":
		return 3
	case "highest":
		return 4
	default:
		return 1
	}
}

func honorFrameSize(mainOrSub string) string {
	if mainOrSub == "sub" || mainOrSub == "s" {
		return "s"
	}
	return "m"
}

// cardThumbnailBase normalizes card thumbnails to JP assets. Card thumbnail
// filenames are shared across regions and JP is the fastest-updating superset,
// so CN/TW/KR/EN commands can reuse one JP thumbnail cache.
func cardThumbnailBase(baseURL string) string {
	if baseURL == "" {
		return baseURL
	}
	for _, regionAssetPath := range []string{
		"sekai-cn-assets",
		"sekai-sc-assets",
		"sekai-tw-assets",
		"sekai-tc-assets",
		"sekai-kr-assets",
		"sekai-en-assets",
		"sekai-ww-assets",
		"sekai-global-assets",
	} {
		if strings.Contains(baseURL, regionAssetPath) {
			return strings.Replace(baseURL, regionAssetPath, "sekai-jp-assets", 1)
		}
	}
	return baseURL
}

// GetCardThumbnailURL returns the thumbnail URL for a card.
//
//	thumbnail/chara/{name}_{normal|after_training}.png
func GetCardThumbnailURL(assetBundleName string, trained bool) string {
	return fmt.Sprintf("%s/thumbnail/chara/%s_%s.png",
		cardThumbnailBase(cdnBase()), assetBundleName, trainingSuffix(trained))
}

// GetCardFullURL returns the full-size card illustration URL.
//
//	character/member/{name}/card_{normal|after_training}.png
func GetCardFullURL(assetBundleName string, trained bool) string {
	return fmt.Sprintf("%s/character/member/%s/card_%s.png",
		cdnBase(), assetBundleName, trainingSuffix(trained))
}

// GetMusicJacketURL returns the jacket (cover art) URL for a music track.
//
//	music/jacket/{name}/{name}.png
func GetMusicJacketURL(assetBundleName string) string {
	return fmt.Sprintf("%s/music/jacket/%s/%s.png",
		cdnBase(), assetBundleName, assetBundleName)
}

// GetEventBannerURL returns the event banner background URL.
//
//	event/{name}/screen/bg.png
func GetEventBannerURL(assetBundleName string) string {
	return fmt.Sprintf("%s/event/%s/screen/bg.png",
		cdnBase(), assetBundleName)
}

// GetEventLogoURL returns the event logo URL.
//
//	event/{name}/logo/logo.png
func GetEventLogoURL(assetBundleName string) string {
	return fmt.Sprintf("%s/event/%s/logo/logo.png",
		cdnBase(), assetBundleName)
}

// GetGachaLogoURL returns the gacha banner logo URL.
//
//	gacha/{name}/logo/logo.png
func GetGachaLogoURL(assetBundleName string) string {
	return fmt.Sprintf("%s/gacha/%s/logo/logo.png",
		cdnBase(), assetBundleName)
}

// GetCharacterIconURL returns the small character icon hosted on the static
// server (not the CDN).
//
//	moe.exmeaning.com/assets/chr_ts_{id}.png
func GetCharacterIconURL(charID int) string {
	return fmt.Sprintf("%s/chr_ts_%d.png", StaticBase, charID)
}

// GetAttrIconURL returns the attribute icon URL.
//
//	thumbnail/common/attribute/{attr}.png
func GetAttrIconURL(attr Attribute) string {
	return fmt.Sprintf("%s/thumbnail/common/attribute/%s.png",
		cdnBase(), string(attr))
}

// GetUnitLogoURL returns the unit logo URL.
//
//	thumbnail/common/unit/{id}.png
func GetUnitLogoURL(unitID UnitID) string {
	return fmt.Sprintf("%s/thumbnail/common/unit/%s.png",
		cdnBase(), string(unitID))
}

// GetHonorBgURL returns the honor (title / degree) background URL.
//
//	honor/{name}/degree_{main|sub}.png
func GetHonorBgURL(assetBundleName string, mainOrSub string) string {
	return fmt.Sprintf("%s/honor/%s/degree_%s.png",
		cdnBase(), assetBundleName, mainOrSub)
}

// GetHonorFrameURL returns the default honor frame URL for a rarity.
func GetHonorFrameURL(rarity string, mainOrSub string) string {
	return fmt.Sprintf("%s/honor/frame/frame_degree_%s_%d.png",
		cdnBase(), honorFrameSize(mainOrSub), honorRarityNumber(rarity))
}

// GetHonorLevelIconURL returns a profile honor level-star icon URL.
func GetHonorLevelIconURL(sixPlus bool) string {
	name := "icon_degreeLv"
	if sixPlus {
		name = "icon_degreeLv6"
	}
	return fmt.Sprintf("%s/honor/frame/%s.png", cdnBase(), name)
}

// GetStampURL returns the stamp / sticker image URL.
//
//	stamp/{name}/{name}.png
func GetStampURL(assetBundleName string) string {
	return fmt.Sprintf("%s/stamp/%s/%s.png",
		cdnBase(), assetBundleName, assetBundleName)
}
