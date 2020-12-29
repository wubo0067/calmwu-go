package proto

type ProtoGuildModifiableInfo struct {
	Symbol          string `json:"Symbol"`
	JoinType        int    `json:"JoinType"`
	Description     string `json:"Description"`
	CondLeagueLevel int    `json:"CondLeagueLevel"`
}

type ProtoGuildInfo struct {
	ProtoGuildModifiableInfo
	GuildId      string `json:"GuildId"`
	PerformId    int    `json:"PerformId"`
	Name         string `json:"Name"`
	Level        int    `json:"Level"`
	Rank         int    `json:"Rank"`
	Vitality     int    `json:"Vitality"`
	Chairman     int    `json:"Chairman"`
	ChairmanName string `json:"ChairmanName"`
	CreateTime   int    `json:"CreateTime"`
	MemberCount  int    `json:"MemberCount"`
}

type ProtoGuildCreateInfo struct {
	ProtoGuildModifiableInfo `json:",squash"`
	Name                     string `json:"Name"`
}

type ProtoCreateGuildRequest struct {
	ServerId  int                  `json:"ServerId"`
	GuildInfo ProtoGuildCreateInfo `json:"GuildInfo"`
}

type ProtoCreateGuildResponse struct {
	GuildInfo    ProtoGuildInfo          `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	Cost         ProtoResourcesAttr      `json:"Cost"`
}

type ProtoModifyGuildInfoRequest struct {
	ProtoGuildModifiableInfo `json:",squash"`
}

type ProtoModifyGuildInfoResponse struct {
	ProtoGuildInfo `json:"GuildInfo"`
}

type ProtoGetSimpleGuildInfoByUinRequest struct {
	Uin int `json:"Uin"`
}

type ProtoGetSimpleGuildInfoByUinResponse struct {
	GuildInfo *ProtoGuildInfo `json:"GuildInfo"`
	GuildMemberInfo *ProtoGuildMemberInfo `json:"GuildMemberInfo"`
}

type ProtoQuerySingleGuildInfoRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoQuerySingleGuildInfoResponse struct {
	GuildInfo *ProtoGuildInfo `json:"GuildInfo"`
}

type ProtoQueryMultiGuildInfoRequest struct {
	GuildIdList []string `json:"GuildIdList"`
}

type ProtoQueryMultiGuildInfoResponse struct {
	GuildInfoList []*ProtoGuildInfo `json:"GuildInfoList"`
}

type ProtoJoinGuildRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoJoinGuildWaitForLeaveTimeResponse struct {
	RestWaitTime int `json:"RestWaitTime"`
}

type ProtoJoinGuildSuccessResponse struct {
	GuildInfo    ProtoGuildInfo          `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList  []*ProtoMessageInfo     `json:"MessageList"`
}

type ProtoLeaveGuildResponse struct {
	DeleteGuild  int                     `json:"DeleteGuild"`
	GuildInfo    *ProtoGuildInfo         `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList  []*ProtoMessageInfo     `json:"MessageList"`
}

type ProtoKickOutOfGuildRequest struct {
	MemberUin int `json:"MemberUin"`
}

type ProtoKickOutOfGuildResponse struct {
	GuildInfo    *ProtoGuildInfo         `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList  []*ProtoMessageInfo     `json:"MessageList"`
}

type ProtoPromoteMemberRequest struct {
	Uin int `json:"Uin"`
}

type ProtoPromoteMemberResponse struct {
	GuildInfo    ProtoGuildInfo          `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList  []*ProtoMessageInfo     `json:"MessageList"`
}

type ProtoDemoteMemberRequest struct {
	Uin int `json:"Uin"`
}

type ProtoDemoteMemberResponse struct {
	GuildInfo    ProtoGuildInfo          `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList  []*ProtoMessageInfo     `json:"MessageList"`
}

type ProtoGetGuildInfoRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoGetGuildInfoResponse struct {
	GuildInfo    ProtoGuildInfo          `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList  []*ProtoMessageInfo     `json:"MessageList"`
}

type ProtoGetSimpleGuildInfoWithMembersRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoGetSimpleGuildInfoWithMembersResponse struct {
	GuildInfo    ProtoGuildInfo          `json:"GuildInfo"`
	GuildMembers []*ProtoGuildMemberInfo `json:"GuildMembers"`
}

type ProtoGetGuildMemberUinListRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoGetGuildMemeberUinListResponse struct {
	UinList []int `json:"UinList"`
}

type ProtoGetGuildMemberInfoByUinRequest struct {
	GuildId string `json:"GuildId"`
	Uin     int    `json:"Uin"`
}

type ProtoGetGuildMemberInfoByUinResponse struct {
	GuildMemberInfo *ProtoGuildMemberInfo `json:"GuildMemberInfo"`
}
