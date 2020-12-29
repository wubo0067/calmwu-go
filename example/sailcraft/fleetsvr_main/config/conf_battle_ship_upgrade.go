package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type BattleShipUpgradeAttr struct {
	Card    int           `json:"Card"`
	Cost    ResourcesAttr `json:"Cost"`
	Level   int           `json:"Level"`
	SumCard int
	SumCost ResourcesAttr
}

type BattleShipUpgradeRawConfig struct {
	Common    []BattleShipUpgradeAttr `json:"Common"`
	Epic      []BattleShipUpgradeAttr `json:"Epic"`
	Rare      []BattleShipUpgradeAttr `json:"Rare"`
	Legendary []BattleShipUpgradeAttr `json:"Legendary"`
}

type BattleShipUpgradeConfig struct {
	AttrMap map[string]BattleShipUpgradeAttr
}

var (
	GBattleShipUpgradeConfig *BattleShipUpgradeConfig = new(BattleShipUpgradeConfig)
)

func genBattleShipUpgradeKey(quality string, level int) string {
	return fmt.Sprintf("%s_%d", quality, level)
}

func (conf *BattleShipUpgradeConfig) GetUpgradeAttr(quality string, level int) (BattleShipUpgradeAttr, error) {
	key := genBattleShipUpgradeKey(quality, level)
	if value, ok := conf.AttrMap[key]; ok {
		return value, nil
	} else {
		return BattleShipUpgradeAttr{}, fmt.Errorf("config is not exist %s", key)
	}
}

func (conf *BattleShipUpgradeConfig) addUpgradeAttr(quality string, list *[]BattleShipUpgradeAttr) error {
	if list == nil {
		return fmt.Errorf("null point")
	}

	for _, attr := range *list {
		key := genBattleShipUpgradeKey(quality, attr.Level)
		conf.AttrMap[key] = attr
	}

	return nil
}

func (conf *BattleShipUpgradeConfig) calculateSumCost(quality string) {
	level := 1
	sumCard := 0
	var sumCost ResourcesAttr

	for {
		key := genBattleShipUpgradeKey(quality, level)

		if attr, ok := conf.AttrMap[key]; ok {
			sumCard += attr.Card
			sumCost.Add(&attr.Cost)
			attr.SumCard = sumCard
			attr.SumCost.Add(&sumCost)
			conf.AttrMap[key] = attr
		} else {
			break
		}

		level++
	}
}

func (conf *BattleShipUpgradeConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.AttrMap = make(map[string]BattleShipUpgradeAttr)

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

	var rawConfigData BattleShipUpgradeRawConfig
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

	base.GLog.Debug("battle ship upgrade AttrMap size is %d", len(conf.AttrMap))
	for key, ptr := range conf.AttrMap {
		base.GLog.Debug("battle ship upgrade conf key is %s value is [%+v]", key, ptr)
	}

	return nil
}
