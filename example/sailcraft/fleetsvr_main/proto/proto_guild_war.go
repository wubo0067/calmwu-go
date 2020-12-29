package proto

type ProtoGuildWarInfo struct {
	GuildWardId   int `json:"GuildWardId"`   // 公会战Id
	GuildWarPhase int `json:"GuildWarPhase"` // 公会战阶段
	RestTime      int `json:"RestTime"`      // 剩余时间
	RewardGroupId int `json:"RewardGroupId"` // 奖励GroupId
}

type ProtoGuildMemberBattleRecord struct {
	ScoreIncr    int `json:"ScoreIncr"`
	BattleResult int `json:"BattleResult"`
}

type ProtoGuildMemberBattleInfo struct {
	Uin           int                             `json:"Uin"`
	UserName      string                          `json:"UserName"`
	Icon          string                          `json:"Icon"`
	Level         int                             `json:"Level"`
	CountryCode   string                          `json:"ISOCountryCode"`
	CurHeadFrame  string                          `json:"CurHeadFrame"`
	HeadFrameProtypeId int 						  `json:"HeadFrameProtypeId"`	
	HeadId        string                          `json:"HeadId"`
	HeadType      int                             `json:"HeadType"`
	Score         int                             `json:"Score"`
	BattleRecords []*ProtoGuildMemberBattleRecord `json:"BattleRecords"`
	//LeagueLevel   int                             `json:"LeagueLevel"`
}

type ProtoGetGuildWarInfoResponse struct {
	GuildWarInfo ProtoGuildWarInfo `json:"GuildWarInfo"`
}

type ProtoGuildWarMemberRankRequest struct {
	GuildId string `json:"GuildId"`
	WarId   int    `json:"WarId"`
}

type ProtoGuildWarMemberRankResponse struct {
	MemberBattleInfoList []*ProtoGuildMemberBattleInfo `json:"MemberBattleInfoList"`
}
