package proto

type ProtoGuildWarBattleSettleRequest struct {
	GuildId      string `json:"GuildId"`
	WarId        int    `json:"WarId"`
	BattleResult int    `json:"BattleResult"`
}

type ProtoGuildWarBattleSettleResponse struct {
	ScoreDelta int `json:"ScoreDelta"`
	Score      int `json:"Score"`
	GuildScore int `json:"GuildScore"`
	WinStreak  int `json:"WinStreak"`
}
