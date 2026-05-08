package deckrecommenddata

import "testing"

func TestLoadLocalMasterDataWorldBloomSupportDeckBonuses(t *testing.T) {
	for _, key := range []string{
		"worldBloomSupportDeckBonusesWL1",
		"worldBloomSupportDeckBonusesWL2",
		"worldBloomSupportDeckBonusesWL3",
	} {
		t.Run(key, func(t *testing.T) {
			if !IsLocalMasterKey(key) {
				t.Fatalf("%s should be marked as local master data", key)
			}
			data, err := LoadLocalMasterData(key)
			if err != nil {
				t.Fatalf("LoadLocalMasterData(%s): %v", key, err)
			}
			if len(data) == 0 {
				t.Fatalf("LoadLocalMasterData(%s) returned empty data", key)
			}
		})
	}
}
