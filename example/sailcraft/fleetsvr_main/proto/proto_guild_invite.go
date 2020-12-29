package proto

type ProtoGuildInviteInfo struct {
	TargetUin       int             `json:"TargetUin"`
	FromUin         int             `json:"FromUin"`
	FromUserName    string          `json:"FromUserName"`
	FromCountryCode string          `json:"FromISOCountryCode"`
	FromIcon        string          `json:"FromIcon"`
	Invitetime      int             `json:"InviteTime"`
	GuildInfo       *ProtoGuildInfo `json:"GuildInfo"`
}

type ProtoGuildSendInviteRequest struct {
	TargetUin int `json:"TargetUin"`
}

type ProtoGuildGetAllInviteResponse struct {
	InviteList []*ProtoGuildInviteInfo `json:"InviteList"`
}

type ProtoHandleGuildInviteRequest struct {
	GuildId   string `json:"GuildId"`
	Operation int    `json:"Operation"`
}

type ProtoHandleGuildInviteResponse struct {
	Operation              int                     `json:"Operation"`
	GuildInfo              *ProtoGuildInfo         `json:"GuildInfo"`
	GuildMembers           []*ProtoGuildMemberInfo `json:"GuildMembers"`
	MessageList            []*ProtoMessageInfo     `json:"MessageList"`
	EnterGuildCoolDownTime int                     `json:"EnterGuildCoolDownTime"`
}
