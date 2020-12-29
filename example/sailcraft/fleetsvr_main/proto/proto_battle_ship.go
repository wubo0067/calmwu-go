package proto

type ProtoInitBattleShipRequest struct {
	ProtypeIDs []int `json:"ProtypeIDs"`
}

type ProtoInitBattleShipResponse struct {
	ProtypeIDs []int `json:"ProtypeIDs"`
}

type ProtoBattleShipInfo struct {
	Id         int
	Uin        int
	ProtypeID  int
	Level      int
	StarLevel  int
	CardNumber int
	Status     int
}

type ProtoGetBattleShipListResponse struct {
	BattleShips []ProtoBattleShipInfo `json:"BattleShips"`
}

type ProtoUpgradeBattleShipLevelRequest struct {
	ShipID int `json:"ShipID"`
}

type ProtoUpgradeBattleShipLevelResponse struct {
	BattleShip           ProtoBattleShipInfo `json:"BattleShip"`
	ProtoUserInfoChanged `mapstructure:",squash"`
}

type ProtoUpgradeBattleShipStarLevelRequest struct {
	ShipID int `json:"ShipID"`
}

type ProtoUpgradeBattleShipStarLevelResponse struct {
	BattleShip           ProtoBattleShipInfo `json:"BattleShip"`
	ProtoUserInfoChanged `mapstructure:",squash"`
}

type ProtoBattleShipCardNumberParams struct {
	ProtypeID  int `json:"ProtypeID"`
	CardNumber int `json:"CardNumber"`
}

type ProtoAddBattleShipCardNumberRequest struct {
	ShipCards []*ProtoBattleShipCardNumberParams `json:"ShipCards"`
}

type ProtoAddBattleShipCardNumberResponse struct {
	BattleShips []ProtoBattleShipInfo `json:"BattleShips"`
}

type ProtoCheckBattleShipCardNumberRequest struct {
	ShipCards []*ProtoBattleShipCardNumberParams `json:"ShipCards"`
}

type ProtoModifyBattleShipCardNumberRequest struct {
	ShipCards []*ProtoBattleShipCardNumberParams `json:"ShipCards"`
}

type ProtoModifyBattleShipCardNumberResponse struct {
	BattleShips []*ProtoBattleShipInfo `json:"BattleShips"`
}

type ProtoComposeBattleShipRequest struct {
	ShipID int `json:"ShipID"`
}

type ProtoComposeBattleShipResponse struct {
	BattleShip           ProtoBattleShipInfo `json:"BattleShip"`
	ProtoUserInfoChanged `mapstructure:",squash"`
}

type ProtoReclaimBattleShipCardsRequest struct {
	ShipCards []*ProtoBattleShipCardNumberParams `json:"ShipCards"`
}

type ProtoReclaimBattleShipCardsResponse struct {
	Cost    ProtoResourcesAttr `json:"Cost"`
	Rewards ProtoResourcesAttr `json:"Rewards"`
}

type ProtoDisassembleBattleShipRequest struct {
	ProtypeIdList []int `json:"ProtypeIdList"`
}

type ProtoDisassembleBattleShipResponse struct {
	BattleShips []ProtoBattleShipInfo `json:"BattleShips"`
	Rewards     ProtoResourcesAttr    `json:"Rewards"`
	Cost        ProtoResourcesAttr    `json:"Cost"`
}
