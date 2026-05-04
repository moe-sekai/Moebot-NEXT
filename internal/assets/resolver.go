package assets

import (
	"fmt"

	"moebot-next/internal/config"
)

// Resolver builds asset URLs for one configured game server without mutating
// the package-level legacy CDN state. Commands use one Resolver per region so
// concurrent JP/CN/TW/KR/EN requests cannot overwrite each other's asset source.
type Resolver struct {
	resolved config.ResolvedAssets
}

// NewResolver resolves an AssetsConfig into an immutable URL builder.
func NewResolver(cfg config.AssetsConfig, defaultRegion string) (*Resolver, error) {
	resolved, err := config.ResolveAssets(cfg, defaultRegion)
	if err != nil {
		return nil, err
	}
	return &Resolver{resolved: resolved}, nil
}

// DefaultResolver returns a resolver backed by the current legacy global CDN
// state. It exists for compatibility with older code paths and tests.
func DefaultResolver() *Resolver {
	return &Resolver{resolved: config.ResolvedAssets{
		BaseURL:     cdnBase(),
		RendererKey: GetRendererAssetSource(),
		CDNSource:   string(GetCDNSource()),
	}}
}

// Resolved returns the effective asset settings used by this resolver.
func (r *Resolver) Resolved() config.ResolvedAssets {
	if r == nil {
		return DefaultResolver().Resolved()
	}
	return r.resolved
}

// BaseURL returns the configured asset base URL.
func (r *Resolver) BaseURL() string {
	return r.Resolved().BaseURL
}

// RendererAssetSource returns the renderer-side asset source key or custom URL.
func (r *Resolver) RendererAssetSource() string {
	resolved := r.Resolved()
	if resolved.RendererKey != "" {
		return resolved.RendererKey
	}
	return resolved.BaseURL
}

func (r *Resolver) base() string {
	baseURL := r.BaseURL()
	if baseURL == "" {
		return cdnBase()
	}
	return baseURL
}

// GetCardThumbnailURL returns the thumbnail URL for a card. Thumbnails are
// normalized to JP assets because the filenames are shared across regions and
// JP is the fastest-updating superset for released cards.
func (r *Resolver) GetCardThumbnailURL(assetBundleName string, trained bool) string {
	return fmt.Sprintf("%s/thumbnail/chara/%s_%s.png", cardThumbnailBase(r.base()), assetBundleName, trainingSuffix(trained))
}

// GetCardFullURL returns the full-size card illustration URL.
func (r *Resolver) GetCardFullURL(assetBundleName string, trained bool) string {
	return fmt.Sprintf("%s/character/member/%s/card_%s.png", r.base(), assetBundleName, trainingSuffix(trained))
}

// GetMusicJacketURL returns the jacket (cover art) URL for a music track.
func (r *Resolver) GetMusicJacketURL(assetBundleName string) string {
	return fmt.Sprintf("%s/music/jacket/%s/%s.png", r.base(), assetBundleName, assetBundleName)
}

// GetEventBannerURL returns the event banner background URL.
func (r *Resolver) GetEventBannerURL(assetBundleName string) string {
	return fmt.Sprintf("%s/event/%s/screen/bg.png", r.base(), assetBundleName)
}

// GetEventLogoURL returns the event logo URL.
func (r *Resolver) GetEventLogoURL(assetBundleName string) string {
	return fmt.Sprintf("%s/event/%s/logo/logo.png", r.base(), assetBundleName)
}

// GetGachaLogoURL returns the gacha banner logo URL.
func (r *Resolver) GetGachaLogoURL(assetBundleName string) string {
	return fmt.Sprintf("%s/gacha/%s/logo/logo.png", r.base(), assetBundleName)
}

// GetAttrIconURL returns the attribute icon URL.
func (r *Resolver) GetAttrIconURL(attr Attribute) string {
	return fmt.Sprintf("%s/thumbnail/common/attribute/%s.png", r.base(), string(attr))
}

// GetUnitLogoURL returns the unit logo URL.
func (r *Resolver) GetUnitLogoURL(unitID UnitID) string {
	return fmt.Sprintf("%s/thumbnail/common/unit/%s.png", r.base(), string(unitID))
}

// GetHonorBgURL returns the honor background URL.
func (r *Resolver) GetHonorBgURL(assetBundleName string, mainOrSub string) string {
	return fmt.Sprintf("%s/honor/%s/degree_%s.png", r.base(), assetBundleName, mainOrSub)
}

// GetStampURL returns the stamp / sticker image URL.
func (r *Resolver) GetStampURL(assetBundleName string) string {
	return fmt.Sprintf("%s/stamp/%s/%s.png", r.base(), assetBundleName, assetBundleName)
}
