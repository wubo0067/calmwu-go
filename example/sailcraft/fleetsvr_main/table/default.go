package table

// 默认是碎片状态
func NewDefaultBattleShip() *TblBattleShip {
	battleShip := new(TblBattleShip)
	battleShip.Level = 0
	battleShip.StarLevel = 0
	battleShip.CardNumber = 0
	battleShip.Status = BATTLE_SHIP_STATUS_CARD
	battleShip.Reserved0 = 0
	battleShip.Reserved1 = 0
	battleShip.Reserved2 = ""
	battleShip.Reserved3 = ""
	battleShip.Reserved4 = ""
	battleShip.Reserved5 = ""

	return battleShip
}
