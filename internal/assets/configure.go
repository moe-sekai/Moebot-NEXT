package assets

import "moebot-next/internal/config"

// Configure applies the resource CDN configuration and returns the resolved settings.
func Configure(cfg config.AssetsConfig, defaultRegion string) (config.ResolvedAssets, error) {
	resolved, err := config.ResolveAssets(cfg, defaultRegion)
	if err != nil {
		return config.ResolvedAssets{}, err
	}

	cdnMu.Lock()
	defer cdnMu.Unlock()
	cdnSource = CDNSource(resolved.CDNSource)
	cdnBaseURL = resolved.BaseURL
	rendererAssetSource = resolved.RendererKey
	return resolved, nil
}
