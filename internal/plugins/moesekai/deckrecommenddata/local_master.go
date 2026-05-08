package deckrecommenddata

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed deck_recommend_data/*.json
var localMasterData embed.FS

var localMasterKeys = map[string]string{
	"worldBloomSupportDeckBonusesWL1": "deck_recommend_data/worldBloomSupportDeckBonusesWL1.json",
	"worldBloomSupportDeckBonusesWL2": "deck_recommend_data/worldBloomSupportDeckBonusesWL2.json",
	"worldBloomSupportDeckBonusesWL3": "deck_recommend_data/worldBloomSupportDeckBonusesWL3.json",
}

// IsLocalMasterKey reports whether the master data key is provided by the local
// Snowy Viewer compatibility data bundle instead of the remote JP masterdata endpoint.
func IsLocalMasterKey(key string) bool {
	_, ok := localMasterKeys[key]
	return ok
}

// LoadLocalMasterData loads locally bundled master data that does not exist on
// the JP masterdata endpoint, such as per-turn World Bloom support deck bonuses.
func LoadLocalMasterData(key string) ([]any, error) {
	path, ok := localMasterKeys[key]
	if !ok {
		return nil, fmt.Errorf("local masterdata %s not configured", key)
	}
	body, err := localMasterData.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read local masterdata %s: %w", key, err)
	}
	var data []any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("decode local masterdata %s: %w", key, err)
	}
	return data, nil
}
