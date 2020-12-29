package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/mitchellh/mapstructure"
)

const (
	BATTLE_SHIP_QUALITY_COMMON    = "Common"
	BATTLE_SHIP_QUALITY_EPIC      = "Epic"
	BATTLE_SHIP_QUALITY_RARE      = "Rare"
	BATTLE_SHIP_QUALITY_LEGENDARY = "Legendary"
)

type BattleShipLevelAttr struct {
	Level   int `mapstructure:"Level"`
	Luck    int `mapstructure:"Luck"`
	Seaarea int `mapstructure:"Seaarea"`
	Energy  int `mapstructure:"Energy"`
	Ammo    int `mapstructure:"Ammo"`
}

type BattleShipStarAttr struct {
	StarLevel     int    `mapstructure:"StarLevel"`
	Energy        int    `mapstructure:"Energy"`
	Ammo          int    `mapstructure:"Ammo"`
	Seaarea       int    `mapstructure:"Seaarea"`
	Race          string `mapstructure:"Race"`
	ResourceID    string `mapstructure:"ResourceId"`
	Rarity        string `mapstructure:"Rarity"`
	Feature       string `mapstructure:"Feature"`
	Shape         int    `mapstructure:"Shape"`
	Hp            int    `mapstructure:"HP"`
	NameKey       string `mapstructure:"NameKey"`
	DescKey       string `mapstructure:"DescKey"`
	Luck          int    `mapstructure:"Luck"`
	SkillActive   int    `mapstructure:"SkillActive"`
	SkillPassive1 int    `mapstructure:"SkillPassive1"`
	SkillPassive2 int    `mapstructure:"SkillPassive2"`
	SkillPassive3 int    `mapstructure:"SkillPassive3"`
}

type BattleShipProtype struct {
	ID        int                   `mapstructure:"Id"`
	LevelList []BattleShipLevelAttr `mapstructure:"Level"`
	StarList  []BattleShipStarAttr  `mapstructure:"StarLevel"`
}

type BattleShipProtypeConfig struct {
	AttrMap map[int]*BattleShipProtype
}

var (
	GBattleShipProtypeConfig *BattleShipProtypeConfig = new(BattleShipProtypeConfig)
)

func (conf *BattleShipProtypeConfig) GetLevelAttr(protypeID int, level int) (BattleShipLevelAttr, error) {
	if battleShipProtype, ok := conf.AttrMap[protypeID]; ok {
		for _, levelAttr := range battleShipProtype.LevelList {
			if levelAttr.Level == level {
				return levelAttr, nil
			}
		}
	}

	return BattleShipLevelAttr{}, fmt.Errorf("can not find [protypeID:%d level:%d] in battle ship protype config", protypeID, level)
}

func (conf *BattleShipProtypeConfig) GetStarAttr(protypeID int, starLevel int) (BattleShipStarAttr, error) {
	if battleShipProtype, ok := conf.AttrMap[protypeID]; ok {
		for _, starAttr := range battleShipProtype.StarList {
			if starAttr.StarLevel == starLevel {
				return starAttr, nil
			}
		}
	}

	return BattleShipStarAttr{}, fmt.Errorf("can not find [protypeID:%d level:%d] in battle ship protype config", protypeID, starLevel)
}

func (conf *BattleShipProtypeConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.AttrMap = make(map[int]*BattleShipProtype)

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
			var attr *BattleShipProtype = new(BattleShipProtype)
			err = mapstructure.Decode(value, attr)
			if err != nil {
				base.GLog.Error("mapstructure Decode %s failed err %s \n", configFile, err.Error())
				return err
			}

			conf.AttrMap[attr.ID] = attr

			base.GLog.Debug("battle ship protype data is [%+v]", *attr)
		}
	}

	base.GLog.Debug("battle ship protype AttrMap size is %d", len(conf.AttrMap))

	return nil
}
