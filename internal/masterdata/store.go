package masterdata

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// store.go — Thread-safe in-memory masterdata store
//
// All public getters acquire an RLock so they are safe for concurrent reads.
// SetAll acquires a full write Lock and atomically swaps all data + indexes.
// ---------------------------------------------------------------------------

// Store holds all loaded masterdata in memory with thread-safe access.
type Store struct {
	mu sync.RWMutex

	// ---- raw slices (source of truth) ----
	loadedAt                          time.Time
	cards                             []CardInfo
	musics                            []MusicInfo
	musicDifficulties                 []MusicDifficulty
	events                            []EventInfo
	eventDeckBonuses                  []EventDeckBonus
	eventCards                        []EventCard
	eventMusics                       []EventMusic
	worldBlooms                       []WorldBloom
	virtualLives                      []VirtualLive
	gachas                            []GachaInfo
	cardSupplies                      []CardSupplyInfo
	skills                            []SkillInfo
	characterUnits                    []GameCharacterUnit
	honors                            []HonorInfo
	bondsHonors                       []BondsHonorInfo
	bondsHonorWords                   []BondsHonorWordInfo
	musicVocals                       []MusicVocal
	challengeLiveHighScoreRewards     []ChallengeLiveHighScoreReward
	resourceBoxes                     []ResourceBox
	resourceBoxDetails                []ResourceBoxDetail
	characterMissionV2ParameterGroups []CharacterMissionV2ParameterGroup

	// ---- primary-key indexes (ID → *element inside slice) ----
	cardByID           map[int]*CardInfo
	musicByID          map[int]*MusicInfo
	eventByID          map[int]*EventInfo
	gachaByID          map[int]*GachaInfo
	virtualLiveByID    map[int]*VirtualLive
	cardSupplyByID     map[int]*CardSupplyInfo
	skillByID          map[int]*SkillInfo
	characterUnitByID  map[int]*GameCharacterUnit
	honorByID          map[int]*HonorInfo
	bondsHonorByID     map[int]*BondsHonorInfo
	bondsHonorWordByID map[int]*BondsHonorWordInfo
	musicVocalByID     map[int]*MusicVocal
	resourceBoxByKey   map[string]*ResourceBox

	// ---- derived / relation indexes ----
	diffsByMusicID           map[int][]MusicDifficulty
	bonusesByEventID         map[int][]EventDeckBonus
	eventCardsByEventID      map[int][]EventCard
	eventMusicsByEventID     map[int][]EventMusic
	worldBloomsByEventID     map[int][]WorldBloom
	unitsByCharacterID       map[int][]GameCharacterUnit
	vocalsByMusicID          map[int][]MusicVocal
	challengeRewardsByCharID map[int][]ChallengeLiveHighScoreReward
	resourceBoxDetailsByKey  map[string][]ResourceBoxDetail
	missionParamGroupsByID   map[int][]CharacterMissionV2ParameterGroup
}

// NewStore creates an empty Store ready for use.
func NewStore() *Store {
	s := &Store{}
	s.initMaps()
	return s
}

// initMaps allocates all index maps so they are never nil.
func (s *Store) initMaps() {
	s.cardByID = make(map[int]*CardInfo)
	s.musicByID = make(map[int]*MusicInfo)
	s.eventByID = make(map[int]*EventInfo)
	s.gachaByID = make(map[int]*GachaInfo)
	s.virtualLiveByID = make(map[int]*VirtualLive)
	s.cardSupplyByID = make(map[int]*CardSupplyInfo)
	s.skillByID = make(map[int]*SkillInfo)
	s.characterUnitByID = make(map[int]*GameCharacterUnit)
	s.honorByID = make(map[int]*HonorInfo)
	s.bondsHonorByID = make(map[int]*BondsHonorInfo)
	s.bondsHonorWordByID = make(map[int]*BondsHonorWordInfo)
	s.musicVocalByID = make(map[int]*MusicVocal)
	s.resourceBoxByKey = make(map[string]*ResourceBox)
	s.diffsByMusicID = make(map[int][]MusicDifficulty)
	s.bonusesByEventID = make(map[int][]EventDeckBonus)
	s.eventCardsByEventID = make(map[int][]EventCard)
	s.eventMusicsByEventID = make(map[int][]EventMusic)
	s.worldBloomsByEventID = make(map[int][]WorldBloom)
	s.unitsByCharacterID = make(map[int][]GameCharacterUnit)
	s.vocalsByMusicID = make(map[int][]MusicVocal)
	s.challengeRewardsByCharID = make(map[int][]ChallengeLiveHighScoreReward)
	s.resourceBoxDetailsByKey = make(map[string][]ResourceBoxDetail)
	s.missionParamGroupsByID = make(map[int][]CharacterMissionV2ParameterGroup)
}

// ---------- Atomic Data Swap -----------------------------------------------

// SetAll replaces every slice and rebuilds all indexes under a write lock.
// It is designed to be called by the Loader after a full refresh.
func (s *Store) SetAll(data *MasterData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Copy slices from the incoming snapshot.
	s.loadedAt = time.Now()
	s.cards = data.Cards
	s.musics = data.Musics
	s.musicDifficulties = data.MusicDifficulties
	s.events = data.Events
	s.eventDeckBonuses = data.EventDeckBonuses
	s.eventCards = data.EventCards
	s.eventMusics = data.EventMusics
	s.worldBlooms = data.WorldBlooms
	s.virtualLives = data.VirtualLives
	s.gachas = data.Gachas
	s.cardSupplies = data.CardSupplies
	s.skills = data.Skills
	s.characterUnits = data.CharacterUnits
	s.honors = data.Honors
	s.bondsHonors = data.BondsHonors
	s.bondsHonorWords = data.BondsHonorWords
	s.musicVocals = data.MusicVocals
	s.challengeLiveHighScoreRewards = data.ChallengeLiveHighScoreRewards
	s.resourceBoxes = data.ResourceBoxes
	s.resourceBoxDetails = data.ResourceBoxDetails
	s.characterMissionV2ParameterGroups = data.CharacterMissionV2ParameterGroups

	s.buildIndexes()
}

// buildIndexes (re)builds all lookup maps from the current slices.
// MUST be called while holding s.mu in write mode.
func (s *Store) buildIndexes() {
	// Reset maps.
	s.initMaps()

	// --- primary-key indexes ---
	for i := range s.cards {
		s.cardByID[s.cards[i].ID] = &s.cards[i]
	}
	for i := range s.musics {
		s.musicByID[s.musics[i].ID] = &s.musics[i]
	}
	for i := range s.events {
		s.eventByID[s.events[i].ID] = &s.events[i]
	}
	for i := range s.gachas {
		s.gachaByID[s.gachas[i].ID] = &s.gachas[i]
	}
	for i := range s.virtualLives {
		s.virtualLiveByID[s.virtualLives[i].ID] = &s.virtualLives[i]
	}
	for i := range s.cardSupplies {
		s.cardSupplyByID[s.cardSupplies[i].ID] = &s.cardSupplies[i]
	}
	for i := range s.skills {
		s.skillByID[s.skills[i].ID] = &s.skills[i]
	}
	for i := range s.characterUnits {
		s.characterUnitByID[s.characterUnits[i].ID] = &s.characterUnits[i]
	}
	for i := range s.honors {
		s.honorByID[s.honors[i].ID] = &s.honors[i]
	}
	for i := range s.bondsHonors {
		s.bondsHonorByID[s.bondsHonors[i].ID] = &s.bondsHonors[i]
	}
	for i := range s.bondsHonorWords {
		s.bondsHonorWordByID[s.bondsHonorWords[i].ID] = &s.bondsHonorWords[i]
	}
	for i := range s.musicVocals {
		s.musicVocalByID[s.musicVocals[i].ID] = &s.musicVocals[i]
	}
	for i := range s.resourceBoxes {
		s.resourceBoxByKey[resourceBoxKey(s.resourceBoxes[i].ResourceBoxPurpose, s.resourceBoxes[i].ID)] = &s.resourceBoxes[i]
	}

	// --- relation indexes ---
	for _, d := range s.musicDifficulties {
		s.diffsByMusicID[d.MusicID] = append(s.diffsByMusicID[d.MusicID], d)
	}
	for _, b := range s.eventDeckBonuses {
		s.bonusesByEventID[b.EventID] = append(s.bonusesByEventID[b.EventID], b)
	}
	for _, c := range s.eventCards {
		s.eventCardsByEventID[c.EventID] = append(s.eventCardsByEventID[c.EventID], c)
	}
	for _, m := range s.eventMusics {
		s.eventMusicsByEventID[m.EventID] = append(s.eventMusicsByEventID[m.EventID], m)
	}
	for _, b := range s.worldBlooms {
		s.worldBloomsByEventID[b.EventID] = append(s.worldBloomsByEventID[b.EventID], b)
	}
	for _, u := range s.characterUnits {
		s.unitsByCharacterID[u.GameCharacterID] = append(s.unitsByCharacterID[u.GameCharacterID], u)
	}
	for _, v := range s.musicVocals {
		s.vocalsByMusicID[v.MusicID] = append(s.vocalsByMusicID[v.MusicID], v)
	}
	for _, reward := range s.challengeLiveHighScoreRewards {
		s.challengeRewardsByCharID[reward.CharacterID] = append(s.challengeRewardsByCharID[reward.CharacterID], reward)
	}
	for _, detail := range s.resourceBoxDetails {
		key := resourceBoxKey(detail.ResourceBoxPurpose, detail.ResourceBoxID)
		s.resourceBoxDetailsByKey[key] = append(s.resourceBoxDetailsByKey[key], detail)
	}
	for _, group := range s.characterMissionV2ParameterGroups {
		s.missionParamGroupsByID[group.ID] = append(s.missionParamGroupsByID[group.ID], group)
	}
	for id := range s.missionParamGroupsByID {
		groups := s.missionParamGroupsByID[id]
		sort.SliceStable(groups, func(i, j int) bool { return groups[i].Seq < groups[j].Seq })
		s.missionParamGroupsByID[id] = groups
	}
}

// ---------- Single-item Getters (by ID) ------------------------------------
// All getters return nil when the ID is not found.
// The returned pointer refers to internal data — callers MUST NOT modify it.

// GetCard returns the card with the given ID, or nil.
func (s *Store) GetCard(id int) *CardInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cardByID[id]
}

// GetMusic returns the music with the given ID, or nil.
func (s *Store) GetMusic(id int) *MusicInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.musicByID[id]
}

// GetEvent returns the event with the given ID, or nil.
func (s *Store) GetEvent(id int) *EventInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.eventByID[id]
}

// GetGacha returns the gacha with the given ID, or nil.
func (s *Store) GetGacha(id int) *GachaInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gachaByID[id]
}

// GetVirtualLive returns the virtual live with the given ID, or nil.
func (s *Store) GetVirtualLive(id int) *VirtualLive {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.virtualLiveByID[id]
}

// GetCardSupply returns the card supply row with the given ID, or nil.
func (s *Store) GetCardSupply(id int) *CardSupplyInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cardSupplyByID[id]
}

// GetSkill returns the skill with the given ID, or nil.
func (s *Store) GetSkill(id int) *SkillInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.skillByID[id]
}

// GetCharacterUnit returns the game-character-unit with the given ID, or nil.
func (s *Store) GetCharacterUnit(id int) *GameCharacterUnit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.characterUnitByID[id]
}

// GetHonor returns the honor with the given ID, or nil.
func (s *Store) GetHonor(id int) *HonorInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.honorByID[id]
}

// GetBondsHonor returns the bonds honor with the given ID, or nil.
func (s *Store) GetBondsHonor(id int) *BondsHonorInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bondsHonorByID[id]
}

// GetBondsHonorWord returns the bonds honor word with the given ID, or nil.
func (s *Store) GetBondsHonorWord(id int) *BondsHonorWordInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bondsHonorWordByID[id]
}

// GetMusicVocal returns the music-vocal with the given ID, or nil.
func (s *Store) GetMusicVocal(id int) *MusicVocal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.musicVocalByID[id]
}

// ---------- Relation Getters -----------------------------------------------

// GetMusicDifficulties returns all difficulty entries for a music ID.
func (s *Store) GetMusicDifficulties(musicID int) []MusicDifficulty {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.diffsByMusicID[musicID]
}

// GetEventDeckBonuses returns all deck-bonus entries for an event ID.
func (s *Store) GetEventDeckBonuses(eventID int) []EventDeckBonus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bonusesByEventID[eventID]
}

// GetEventCards returns all card links for an event ID.
func (s *Store) GetEventCards(eventID int) []EventCard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]EventCard(nil), s.eventCardsByEventID[eventID]...)
}

// GetEventMusics returns all music links for an event ID.
func (s *Store) GetEventMusics(eventID int) []EventMusic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]EventMusic(nil), s.eventMusicsByEventID[eventID]...)
}

// GetWorldBlooms returns all World Link chapters for an event ID.
func (s *Store) GetWorldBlooms(eventID int) []WorldBloom {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]WorldBloom(nil), s.worldBloomsByEventID[eventID]...)
}

// GetCharacterUnits returns all unit memberships for a character ID.
func (s *Store) GetCharacterUnits(characterID int) []GameCharacterUnit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.unitsByCharacterID[characterID]
}

// GetMusicVocals returns all vocal versions for a music ID.
func (s *Store) GetMusicVocals(musicID int) []MusicVocal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vocalsByMusicID[musicID]
}

// GetChallengeLiveHighScoreRewards returns all high-score rewards for a character ID.
func (s *Store) GetChallengeLiveHighScoreRewards(characterID int) []ChallengeLiveHighScoreReward {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ChallengeLiveHighScoreReward(nil), s.challengeRewardsByCharID[characterID]...)
}

// GetResourceBox returns a resource box by purpose and ID.
func (s *Store) GetResourceBox(purpose string, id int) *ResourceBox {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resourceBoxByKey[resourceBoxKey(purpose, id)]
}

// GetResourceBoxDetails returns expanded resource box details by purpose and ID.
func (s *Store) GetResourceBoxDetails(purpose string, id int) []ResourceBoxDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ResourceBoxDetail(nil), s.resourceBoxDetailsByKey[resourceBoxKey(purpose, id)]...)
}

// GetCharacterMissionV2ParameterGroups returns all mission parameter rows for a group ID.
func (s *Store) GetCharacterMissionV2ParameterGroups(id int) []CharacterMissionV2ParameterGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]CharacterMissionV2ParameterGroup(nil), s.missionParamGroupsByID[id]...)
}

// ---------- List Getters (all items) ---------------------------------------
// Returned slices are shallow copies — safe to iterate without holding the lock,
// but the element values should not be mutated.

// AllCards returns a copy of the full card list.
func (s *Store) AllCards() []CardInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CardInfo, len(s.cards))
	copy(out, s.cards)
	return out
}

// AllMusics returns a copy of the full music list.
func (s *Store) AllMusics() []MusicInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]MusicInfo, len(s.musics))
	copy(out, s.musics)
	return out
}

// AllEvents returns a copy of the full event list.
func (s *Store) AllEvents() []EventInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EventInfo, len(s.events))
	copy(out, s.events)
	return out
}

// AllWorldBlooms returns a copy of the full World Link chapter list.
func (s *Store) AllWorldBlooms() []WorldBloom {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]WorldBloom, len(s.worldBlooms))
	copy(out, s.worldBlooms)
	return out
}

// AllGachas returns a copy of the full gacha list.
func (s *Store) AllGachas() []GachaInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]GachaInfo, len(s.gachas))
	copy(out, s.gachas)
	return out
}

// AllVirtualLives returns a copy of the full virtual-live list.
func (s *Store) AllVirtualLives() []VirtualLive {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]VirtualLive, len(s.virtualLives))
	copy(out, s.virtualLives)
	return out
}

// AllEventCards returns a copy of the full event-card relation list.
func (s *Store) AllEventCards() []EventCard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EventCard, len(s.eventCards))
	copy(out, s.eventCards)
	return out
}

// AllEventMusics returns a copy of the full event-music relation list.
func (s *Store) AllEventMusics() []EventMusic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EventMusic, len(s.eventMusics))
	copy(out, s.eventMusics)
	return out
}

// AllCardSupplies returns a copy of the full card-supply list.
func (s *Store) AllCardSupplies() []CardSupplyInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CardSupplyInfo, len(s.cardSupplies))
	copy(out, s.cardSupplies)
	return out
}

// AllSkills returns a copy of the full skill list.
func (s *Store) AllSkills() []SkillInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SkillInfo, len(s.skills))
	copy(out, s.skills)
	return out
}

// AllCharacterUnits returns a copy of the full character-unit list.
func (s *Store) AllCharacterUnits() []GameCharacterUnit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]GameCharacterUnit, len(s.characterUnits))
	copy(out, s.characterUnits)
	return out
}

// AllHonors returns a copy of the full honor list.
func (s *Store) AllHonors() []HonorInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]HonorInfo, len(s.honors))
	copy(out, s.honors)
	return out
}

// AllBondsHonors returns a copy of the full bonds honor list.
func (s *Store) AllBondsHonors() []BondsHonorInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]BondsHonorInfo, len(s.bondsHonors))
	copy(out, s.bondsHonors)
	return out
}

// AllMusicVocals returns a copy of the full music-vocal list.
func (s *Store) AllMusicVocals() []MusicVocal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]MusicVocal, len(s.musicVocals))
	copy(out, s.musicVocals)
	return out
}

// AllChallengeLiveHighScoreRewards returns a copy of all challenge-live score rewards.
func (s *Store) AllChallengeLiveHighScoreRewards() []ChallengeLiveHighScoreReward {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ChallengeLiveHighScoreReward, len(s.challengeLiveHighScoreRewards))
	copy(out, s.challengeLiveHighScoreRewards)
	return out
}

// AllResourceBoxDetails returns a copy of all expanded resource box details.
func (s *Store) AllResourceBoxDetails() []ResourceBoxDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ResourceBoxDetail, len(s.resourceBoxDetails))
	copy(out, s.resourceBoxDetails)
	return out
}

// ---------- Count Helpers --------------------------------------------------

// LoadedAt returns the time when masterdata was last loaded into the store.
func (s *Store) LoadedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadedAt
}

// CardCount returns the number of loaded cards.
func (s *Store) CardCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.cards)
}

// MusicCount returns the number of loaded musics.
func (s *Store) MusicCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.musics)
}

// EventCount returns the number of loaded events.
func (s *Store) EventCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

// GachaCount returns the number of loaded gachas.
func (s *Store) GachaCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.gachas)
}

// VirtualLiveCount returns the number of loaded virtual lives.
func (s *Store) VirtualLiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.virtualLives)
}

func resourceBoxKey(purpose string, id int) string {
	return fmt.Sprintf("%s:%d", purpose, id)
}

// IsLoaded reports whether any masterdata has been loaded into the store.
func (s *Store) IsLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.cards) > 0 || len(s.musics) > 0
}
