package proto

type ProtoMessageInfo struct {
	Uin         int    `json:"Uin"`
	Content     string `json:"Content"`
	SendTime    int    `json:"SendTime"`
	MessageType string `json:"MessageType"`
}

type ProtoMessageUserInfo struct {
	Uin      int    `json:"Uin"`
	UserName string `json:"UserName"`
	Icon     string `json:"Icon"`
	Level    int    `json:"Level"`
}

type ProtoGuildMemberCountChangedMessage struct {
	UserInfo  ProtoMessageUserInfo `json:"UserInfo"`
	Operation int                  `json:"Operation"`
}

type ProtoGuildMemberPostChangedMessage struct {
	Operator   ProtoMessageUserInfo `json:"Operator"`
	TargetUser ProtoMessageUserInfo `json:"TargetUser"`
	Operation  int                  `json:"Operation"`
	FinalPost  string               `json:"FinalPost"`
}

type ProtoGuildChairmanTransferMessage struct {
	OldChairman ProtoMessageUserInfo `json:"OldChairman"`
	NewChairman ProtoMessageUserInfo `json:"NewChairman"`
}

type ProtoFeedGuildFrigateShipMessage struct {
	Exp int `json:"Exp"`
}

type ProtoMessageListRequest struct {
	Channel string `json:"Channel"`
}

type ProtoMessageListResponse struct {
	MessageList []*ProtoMessageInfo `json:"MessageList"`
}

type ProtoSendMessageRequest struct {
	Channel string `json:"Channel"`
	Content string `json:"Content"`
}

type ProtoSendMessageResponse struct {
	MessageList []*ProtoMessageInfo `json:"MessageList"`
}
