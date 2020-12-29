package proto

// 道具
type ProtoPropItem struct {
	ProtypeId int    `json:"ProtypeId"`
	CountType string `json:"CountType"`
	Count     int    `json:"Count"`
}

// 武器卡片
type ProtoWeaponCardItem struct {
	ProtypeId int    `json:"ProtypeId"`
	CountType string `json:"CountType"`
	Count     int    `json:"Count"`
}

// 战舰卡片
type ProtoBattleShipCardItem struct {
	ProtypeId int    `json:"ProtypeId"`
	CountType string `json:"CountType"`
	Count     int    `json:"Count"`
}

// 资源
type ProtoResourceItem struct {
	Type      string `json:"Type"`
	CountType string `json:"CountType"`
	Count     int    `json:"Count"`
}

type ProtoResourcesAttr struct {
	BattleShipCards []*ProtoBattleShipCardItem `json:"BattleShipCards"`
	ResourceItems   []*ProtoResourceItem       `json:"Resources"`
	PropItems       []*ProtoPropItem           `json:"Props"`
}
