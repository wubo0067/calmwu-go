package config

import (
	"testing"
)

func TestResources(t *testing.T) {
	var a ResourcesAttr

	var shipCard1 BattleShipCardItem
	shipCard1.ProtypeId = 10
	shipCard1.CountType = COUNT_TYPE_CONST
	shipCard1.Count = float64(10)
	a.AddShipCards(1, &shipCard1)
	if len(a.BattleShipCards) != 1 || ConstantValue(a.BattleShipCards[0].Count) != 10 {
		t.Error("add battle ship cards error")
	}

	a.AddShipCards(-1, &shipCard1)
	if len(a.BattleShipCards) != 1 || ConstantValue(a.BattleShipCards[0].Count) != 0 {
		t.Error("sub battle ship cards error")
	}
}
