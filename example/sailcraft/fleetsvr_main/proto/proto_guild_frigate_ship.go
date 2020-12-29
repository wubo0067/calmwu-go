package proto

type ProtoGuildFrigateShipInfo struct {
	GuildId   string `json:"GuildId"`
	ProtypeId int    `json:"ProtypeId"`
	OldLevel  int    `json:"OldLevel"`
	Level     int    `json:"Level"`
	Exp       int    `json:"Exp"`
}

type ProtoGetGuildFrigateShipWithoutMessageRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoGetGuildFrigateShipWithoutMessageResponse struct {
	FrigateShipInfo ProtoGuildFrigateShipInfo `json:"FrigateShipInfo"`
}

type ProtoGetGuildFrigateShipRequest struct {
	GuildId string `json:"GuildId"`
}

type ProtoGetGuildFrigateShipResponse struct {
	FrigateShipInfo ProtoGuildFrigateShipInfo `json:"FrigateShipInfo"`
	MessageList     []*ProtoMessageInfo       `json:"MessageList"`
}

type ProtoFeedGuildFrigateShipRequest struct {
	Props []ProtoPropUseInfo `json:"Props"`
}

type ProtoFeedGuildFrigateShipResponse struct {
	FrigateShipInfo ProtoGuildFrigateShipInfo `json:"FrigateShipInfo"`
	Cost            ProtoResourcesAttr        `json:"Cost"`
	Rewards         ProtoResourcesAttr        `json:"Rewards"`
	MessageList     []*ProtoMessageInfo       `json:"MessageList"`
}
