package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/mitchellh/mapstructure"
)

type DisassembleAttrCost struct {
	Gold int `mapstructure:"gold"`
}

type DisassembleAttrGainSkillPatch struct {
	Fail          int `mapstructure:"fail"`
	SkillActive   int `mapstructure:"skillactive"`
	SkillPassive1 int `mapstructure:"skillpassive1"`
	SkillPassive2 int `mapstructure:"skillpassive2"`
}

type BattleShipDisassembleAttr struct {
	Cost            DisassembleAttrCost           `mapstructure:"Cost"`
	GainEnergyStone int                           `mapstructure:"GainEnergyStone"`
	GainOceanDust   int                           `mapstructure:"GainOceanDust"`
	GainSkillPatch  DisassembleAttrGainSkillPatch `mapstructure:"GainSkillPatch"`
	Rarity          string                        `mapstructure:"Legendary"`
}

type BattleShipDisassembleConfig struct {
	AttrList []BattleShipDisassembleAttr
}

var (
	GBattleShipDisassembleConfig = new(BattleShipDisassembleConfig)
)

func (conf *BattleShipDisassembleConfig) GetDisassembleAttr(quality string) (BattleShipDisassembleAttr, error) {
	for _, attr := range conf.AttrList {
		if attr.Rarity == quality {
			return attr, nil
		}
	}

	return BattleShipDisassembleAttr{}, fmt.Errorf("Rarity[%s] not exist in config", quality)
}

func (conf *BattleShipDisassembleConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.AttrList = make([]BattleShipDisassembleAttr, 0)

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

	list := arraylist.New()
	err = list.FromJSON(data)
	if err != nil {
		base.GLog.Error("arraylist parse json %s failed err %s \n", configFile, err.Error())
		return err
	}

	for index := 0; index < list.Size(); index++ {
		value, bRet := list.Get(index)
		if bRet {
			var attr BattleShipDisassembleAttr
			err = mapstructure.Decode(value, &attr)
			if err != nil {
				base.GLog.Error("mapstructure Decode %s failed err %s \n", configFile, err.Error())
				return err
			}

			conf.AttrList = append(conf.AttrList, attr)
		}
	}

	base.GLog.Debug("load battle ship disassemble config result is [%+v]", conf.AttrList)

	return nil
}
