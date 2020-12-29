package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type BattleShipStrengthenAttr struct {
	Card          int           `json:"Card"`
	Cost          ResourcesAttr `json:"Cost"`
	ShipLevelMax  int           `json:"ShipLevelMax"`
	ShipLevelNeed int           `json:"ShipLevelNeed"`
	StarLevel     int           `json:"StarLevel"`
	SumCard       int
	SumCost       ResourcesAttr
}

type BattleShipStrengthenRawConfig struct {
	Common    []BattleShipStrengthenAttr `json:"Common"`
	Epic      []BattleShipStrengthenAttr `json:"Epic"`
	Rare      []BattleShipStrengthenAttr `json:"Rare"`
	Legendary []BattleShipStrengthenAttr `json:"Legendary"`
}

type BattleShipStrengthenConfig struct {
	AttrMap map[string]BattleShipStrengthenAttr
}

var (
	GBattleShipStrengthenConfig *BattleShipStrengthenConfig = new(BattleShipStrengthenConfig)
)

func genBattleShipStrengthenKey(quality string, starLevel int) string {
	return fmt.Sprintf("%s_%d", quality, starLevel)
}

func (conf *BattleShipStrengthenConfig) GetUpgradeAttr(quality string, starLevel int) (BattleShipStrengthenAttr, error) {
	key := genBattleShipStrengthenKey(quality, starLevel)
	if value, ok := conf.AttrMap[key]; ok {
		return value, nil
	} else {
		return BattleShipStrengthenAttr{}, fmt.Errorf("config is not exist %s", key)
	}
}

func (conf *BattleShipStrengthenConfig) addUpgradeAttr(quality string, list *[]BattleShipStrengthenAttr) error {
	if list == nil {
		return fmt.Errorf("null point")
	}

	for _, attr := range *list {
		key := genBattleShipStrengthenKey(quality, attr.StarLevel)
		conf.AttrMap[key] = attr
	}

	return nil
}

func (conf *BattleShipStrengthenConfig) calculateSumCost(quality string) {
	starLevel := 0
	sumCard := 0
	var sumCost ResourcesAttr

	for {
		key := genBattleShipStrengthenKey(quality, starLevel)

		if attr, ok := conf.AttrMap[key]; ok {
			sumCard += attr.Card
			sumCost.Add(&attr.Cost)
			attr.SumCard = sumCard
			attr.SumCost.Add(&sumCost)
			conf.AttrMap[key] = attr
		} else {
			break
		}

		starLevel++
	}
}

func (conf *BattleShipStrengthenConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.AttrMap = make(map[string]BattleShipStrengthenAttr)

	hFile, err := os.Open(configFile)
	if err != nil {
		base.GLog.Error("open file %s failed err %s \n", configFile, err.Error())
		return err
	}
	defer hFile.Close()

	data, err := ioutil.ReadAll(hFile)
	if err != nil {
		base.GLog.Error("read file %s failed err %s \n", configFile, err.Error())
		return err
	}

	var rawConfigData BattleShipStrengthenRawConfig
	err = json.Unmarshal(data, &rawConfigData)
	if err != nil {
		base.GLog.Error("json.Unmarshal json %s failed err %s \n", configFile, err.Error())
		return err
	}

	conf.addUpgradeAttr("Common", &rawConfigData.Common)
	conf.addUpgradeAttr("Epic", &rawConfigData.Epic)
	conf.addUpgradeAttr("Rare", &rawConfigData.Rare)
	conf.addUpgradeAttr("Legendary", &rawConfigData.Legendary)

	conf.calculateSumCost("Common")
	conf.calculateSumCost("Epic")
	conf.calculateSumCost("Rare")
	conf.calculateSumCost("Legendary")

	base.GLog.Debug("battle ship strengthen AttrMap size is %d", len(conf.AttrMap))
	for key, ptr := range conf.AttrMap {
		base.GLog.Debug("battle ship strengthen conf key is %s value is [%+v]", key, ptr)
	}

	return nil
}
