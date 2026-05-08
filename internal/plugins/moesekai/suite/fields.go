package suite

const (
	FieldUploadTime   = "upload_time"
	FieldUserGamedata = "userGamedata"
	FieldUserDecks    = "userDecks"
	FieldUserCards    = "userCards"

	FieldUserGachas            = "userGachas"
	FieldUserMaterials         = "userMaterials"
	FieldUserAreas             = "userAreas"
	FieldUserCharacters        = "userCharacters"
	FieldUserBonds             = "userBonds"
	FieldUserEvents            = "userEvents"
	FieldUserWorldBlooms       = "userWorldBlooms"
	FieldUserMusicResults      = "userMusicResults"
	FieldUserMusicAchievements = "userMusicAchievements"
	FieldUserMusicVocals       = "userMusicVocals"
	FieldUserMusics            = "userMusics"

	FieldUserCharacterMissionV2s        = "userCharacterMissionV2s"
	FieldUserCharacterMissionV2Statuses = "userCharacterMissionV2Statuses"

	FieldUserChallengeLiveSoloDecks            = "userChallengeLiveSoloDecks"
	FieldUserChallengeLiveSoloStages           = "userChallengeLiveSoloStages"
	FieldUserChallengeLiveSoloResults          = "userChallengeLiveSoloResults"
	FieldUserChallengeLiveSoloHighScoreRewards = "userChallengeLiveSoloHighScoreRewards"

	FieldUserMysekaiFixtureGameCharacterPerformanceBonuses = "userMysekaiFixtureGameCharacterPerformanceBonuses"
	FieldUserMysekaiGates                                  = "userMysekaiGates"
)

var defaultHarukiPublicFields = []string{
	FieldUserGamedata,
	FieldUserDecks,
	FieldUserCharacters,
	FieldUserChallengeLiveSoloStages,
	FieldUserChallengeLiveSoloResults,
	FieldUserChallengeLiveSoloHighScoreRewards,
	FieldUserBonds,
	FieldUserMaterials,
	FieldUserAreas,
	FieldUserMysekaiFixtureGameCharacterPerformanceBonuses,
	FieldUserMysekaiGates,
	FieldUploadTime,
}

func DefaultHarukiPublicFields() []string {
	return append([]string(nil), defaultHarukiPublicFields...)
}

func Fields(extra ...string) []string {
	fields := []string{
		FieldUploadTime,
		FieldUserGamedata,
		FieldUserDecks,
		FieldUserCards,
	}
	return append(fields, extra...)
}
