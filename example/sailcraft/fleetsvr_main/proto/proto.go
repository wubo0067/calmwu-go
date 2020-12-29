package proto

type ProtoUserInfo struct {
	Level           int
	Exp             int
	Star            int
	Gold            int
	Wood            int
	Gem             int
	PurchaseGem     int
	Stone           int
	Iron            int
	ChangeNameCount int
	GuildId         string
	ShipSoul        int
}

type ProtoUserInfoChanged struct {
	OldUserInfo ProtoUserInfo `json:"OldUserInfo"`
	UserInfo    ProtoUserInfo `json:"UserInfo"`
}
