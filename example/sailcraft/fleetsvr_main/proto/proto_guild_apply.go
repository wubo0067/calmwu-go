package proto

type ProtoGuildApplyInfo struct {
	Uin         int    `json:"Uin"`
	Level       int    `json:"Level"`
	Name        string `json:"Name"`
	CountryCode string `json:"ISOCountryCode"`
	Icon        string `json:"Icon"`
	LeagueLevel int    `json:"LeagueLevel"`
	ApplyTime   int    `json:"ApplyTime"`
}

type ProtoApplyGuildRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoGetGuildApplyListRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoGetGuildApplyListResponse struct {
	ApplyInfoList []*ProtoGuildApplyInfo `json:"ApplyInfoList"`
}

type ProtoHandlerGuildApplyRequest struct {
	Uin       int `json:"Uin"`
	Operation int `json:"Operation"`
}

type ProtoHandlerGuildApplyResponse struct {
	Operation    int                     `json:"Operation"`
	GuildInfo    *ProtoGuildInfo         `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList  []*ProtoMessageInfo     `json:"MessageList"`
}
