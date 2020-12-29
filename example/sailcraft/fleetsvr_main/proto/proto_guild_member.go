package proto

type ProtoGuildMemberInfo struct {
	Uin            int    `json:"Uin"`
	Level          int    `json:"Level"`
	Name           string `json:"UserName"`
	CountryCode    string `json:"ISOCountryCode"`
	Icon           string `json:"Icon"`
	JoinTime       int    `json:"JoinTime"`
	Post           string `json:"Post"`
	LeagueLevel    int    `json:"LeagueLevel"`
	WeeklyVitality int    `json:"WeeklyVitaliy"`
	CurHeadFrame   string `json:"CurHeadFrame"`
	HeadFrameProtypeId int `json:"HeadFrameProtypeId"`
	HeadId         string `json:"HeadId"`
	HeadType       int    `json:"HeadType"`
}
