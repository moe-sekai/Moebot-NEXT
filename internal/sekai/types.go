package sekai

import (
	"fmt"
	"sort"
)

// Profile is the normalized player profile used by commands and renderer payloads.
type Profile struct {
	UserID           string
	Name             string
	Rank             int
	Signature        string
	TwitterID        string
	TotalPower       int
	Stats            ProfileStats
	MusicClearCounts []MusicClearCount
	CharacterRanks   []CharacterRank
	ChallengeLive    *ChallengeLiveResult
	ProfileHonors    []ProfileHonor
	LeaderCard       *ProfileDeckCard
	DeckCards        []ProfileDeckCard
}

type ProfileStats struct {
	MvpCount       int
	SuperStarCount int
}

type MusicClearCount struct {
	Difficulty string
	LiveClear  int
	FullCombo  int
	AllPerfect int
}

type CharacterRank struct {
	CharacterID int
	Rank        int
}

type ChallengeLiveResult struct {
	CharacterID int
	HighScore   int
}

type ProfileHonor struct {
	Seq       int
	HonorType string
	HonorID   int
	Level     int
}

type ProfileDeckCard struct {
	CardID         int
	Level          int
	Mastery        int
	DefaultImage   string
	SpecialTrained bool
}

type profileResponse struct {
	User struct {
		UserID jsonID `json:"userId"`
		ID     jsonID `json:"id"`
		Name   string `json:"name"`
		Rank   int    `json:"rank"`
	} `json:"user"`
	UserProfile struct {
		UserID    jsonID `json:"userId"`
		Word      string `json:"word"`
		Signature string `json:"signature"`
		TwitterID string `json:"twitterId"`
	} `json:"userProfile"`
	Profile struct {
		UserID     jsonID `json:"userId"`
		ID         jsonID `json:"id"`
		Name       string `json:"name"`
		Rank       int    `json:"rank"`
		Word       string `json:"word"`
		Signature  string `json:"signature"`
		TwitterID  string `json:"twitterId"`
		TotalPower int    `json:"totalPower"`
	} `json:"profile"`
	TotalPower struct {
		TotalPower int `json:"totalPower"`
	} `json:"totalPower"`
	UserDeck struct {
		Leader  int `json:"leader"`
		Member1 int `json:"member1"`
		Member2 int `json:"member2"`
		Member3 int `json:"member3"`
		Member4 int `json:"member4"`
		Member5 int `json:"member5"`
	} `json:"userDeck"`
	UserCards []struct {
		CardID                int    `json:"cardId"`
		Level                 int    `json:"level"`
		MasterRank            int    `json:"masterRank"`
		SpecialTrainingStatus string `json:"specialTrainingStatus"`
		DefaultImage          string `json:"defaultImage"`
	} `json:"userCards"`
	UserCharacters []struct {
		CharacterID   int `json:"characterId"`
		CharacterRank int `json:"characterRank"`
	} `json:"userCharacters"`
	UserChallengeLiveSoloResult struct {
		CharacterID int `json:"characterId"`
		HighScore   int `json:"highScore"`
	} `json:"userChallengeLiveSoloResult"`
	UserProfileHonors []struct {
		Seq              int    `json:"seq"`
		ProfileHonorType string `json:"profileHonorType"`
		HonorID          int    `json:"honorId"`
		HonorLevel       int    `json:"honorLevel"`
	} `json:"userProfileHonors"`
	UserMultiLiveTopScoreCount struct {
		Mvp       int `json:"mvp"`
		SuperStar int `json:"superStar"`
	} `json:"userMultiLiveTopScoreCount"`
	UserMusicDifficultyClearCount []struct {
		MusicDifficultyType string `json:"musicDifficultyType"`
		LiveClear           int    `json:"liveClear"`
		FullCombo           int    `json:"fullCombo"`
		AllPerfect          int    `json:"allPerfect"`
	} `json:"userMusicDifficultyClearCount"`
}

func (r profileResponse) normalize(fallbackUserID string) Profile {
	cardsByID := make(map[int]ProfileDeckCard, len(r.UserCards))
	for _, card := range r.UserCards {
		cardsByID[card.CardID] = ProfileDeckCard{
			CardID:         card.CardID,
			Level:          card.Level,
			Mastery:        card.MasterRank,
			DefaultImage:   card.DefaultImage,
			SpecialTrained: card.SpecialTrainingStatus == "done",
		}
	}

	deckIDs := []int{r.UserDeck.Member1, r.UserDeck.Member2, r.UserDeck.Member3, r.UserDeck.Member4, r.UserDeck.Member5}
	deckCards := make([]ProfileDeckCard, 0, len(deckIDs))
	for _, cardID := range deckIDs {
		if cardID == 0 {
			continue
		}
		card := cardsByID[cardID]
		if card.CardID == 0 {
			card.CardID = cardID
		}
		deckCards = append(deckCards, card)
	}

	var leader *ProfileDeckCard
	leaderID := r.UserDeck.Leader
	if leaderID == 0 && len(deckCards) > 0 {
		leaderID = deckCards[0].CardID
	}
	if leaderID != 0 {
		leaderCard := cardsByID[leaderID]
		if leaderCard.CardID == 0 {
			leaderCard.CardID = leaderID
		}
		leader = &leaderCard
	}

	stats := ProfileStats{
		MvpCount:       r.UserMultiLiveTopScoreCount.Mvp,
		SuperStarCount: r.UserMultiLiveTopScoreCount.SuperStar,
	}
	musicClearCounts := make([]MusicClearCount, 0, len(r.UserMusicDifficultyClearCount))
	for _, count := range r.UserMusicDifficultyClearCount {
		musicClearCounts = append(musicClearCounts, MusicClearCount{
			Difficulty: count.MusicDifficultyType,
			LiveClear:  count.LiveClear,
			FullCombo:  count.FullCombo,
			AllPerfect: count.AllPerfect,
		})
	}
	characterRanks := make([]CharacterRank, 0, len(r.UserCharacters))
	for _, character := range r.UserCharacters {
		characterRanks = append(characterRanks, CharacterRank{
			CharacterID: character.CharacterID,
			Rank:        character.CharacterRank,
		})
	}
	sort.SliceStable(characterRanks, func(i, j int) bool {
		return characterRanks[i].Rank > characterRanks[j].Rank
	})

	var challengeLive *ChallengeLiveResult
	if r.UserChallengeLiveSoloResult.CharacterID != 0 || r.UserChallengeLiveSoloResult.HighScore != 0 {
		challengeLive = &ChallengeLiveResult{
			CharacterID: r.UserChallengeLiveSoloResult.CharacterID,
			HighScore:   r.UserChallengeLiveSoloResult.HighScore,
		}
	}
	profileHonors := make([]ProfileHonor, 0, len(r.UserProfileHonors))
	for _, honor := range r.UserProfileHonors {
		profileHonors = append(profileHonors, ProfileHonor{
			Seq:       honor.Seq,
			HonorType: honor.ProfileHonorType,
			HonorID:   honor.HonorID,
			Level:     honor.HonorLevel,
		})
	}
	sort.SliceStable(profileHonors, func(i, j int) bool {
		return profileHonors[i].Seq < profileHonors[j].Seq
	})

	return Profile{
		UserID:           firstString(r.User.UserID.String(), r.User.ID.String(), r.UserProfile.UserID.String(), r.Profile.UserID.String(), r.Profile.ID.String(), fallbackUserID),
		Name:             firstString(r.User.Name, r.Profile.Name, "未知玩家"),
		Rank:             firstNonZero(r.User.Rank, r.Profile.Rank),
		Signature:        firstString(r.UserProfile.Word, r.UserProfile.Signature, r.Profile.Word, r.Profile.Signature),
		TwitterID:        firstString(r.UserProfile.TwitterID, r.Profile.TwitterID),
		TotalPower:       firstNonZero(r.TotalPower.TotalPower, r.Profile.TotalPower),
		Stats:            stats,
		MusicClearCounts: musicClearCounts,
		CharacterRanks:   characterRanks,
		ChallengeLive:    challengeLive,
		ProfileHonors:    profileHonors,
		LeaderCard:       leader,
		DeckCards:        deckCards,
	}
}

type jsonID string

func (id *jsonID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*id = ""
		return nil
	}
	if data[0] == '"' {
		var value string
		if _, err := fmt.Sscanf(string(data), "%q", &value); err != nil {
			return err
		}
		*id = jsonID(value)
		return nil
	}
	*id = jsonID(string(data))
	return nil
}

func (id jsonID) String() string { return string(id) }

func firstString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
