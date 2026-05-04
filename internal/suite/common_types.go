package suite

type BaseProfile struct {
	UploadTime  int64  `json:"upload_time"`
	Source      string `json:"source"`
	LocalSource string `json:"local_source"`
}

type UserGamedata struct {
	UserID jsonID `json:"userId"`
	Name   string `json:"name"`
	Deck   int    `json:"deck"`
	Coin   int64  `json:"coin"`
}

type UserDeck struct {
	DeckID  int `json:"deckId"`
	Member1 int `json:"member1"`
	Member2 int `json:"member2"`
	Member3 int `json:"member3"`
	Member4 int `json:"member4"`
	Member5 int `json:"member5"`
}

type UserCard struct {
	CardID                int    `json:"cardId"`
	Level                 int    `json:"level"`
	MasterRank            int    `json:"masterRank"`
	SkillLevel            int    `json:"skillLevel"`
	DefaultImage          string `json:"defaultImage"`
	SpecialTrainingStatus string `json:"specialTrainingStatus"`
	CreatedAt             int64  `json:"createdAt"`
}
