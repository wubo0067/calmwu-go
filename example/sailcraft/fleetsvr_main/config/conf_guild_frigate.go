package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sort"
)

type GuildFrigateWeapon struct {
	Id    int `json:"weapon_id"`
	Level int `json:"weapon_level"`
}

type GuildFrigateProtype struct {
	Id            int                `json:"Id"`
	Level         int                `json:"Level"`
	SpellAmmo     int                `json:"SpellAmmo"`
	FleetSeaarea  int                `json:"FleetSeaarea"`
	FleetLuck     int                `json:"FleeteLuck"`
	FeedbackRatio float32            `json:"FeedbackRatio"`
	EnergyMax     int                `json:"EnergyMax"`
	EnergyOutput  int                `json:"EnergyOutput"`
	EnergeBase    int                `json:"EnergyBase"`
	NeedExp       int                `json:"NeedExp"`
	NameKey       string             `json:"NameKey"`
	DescKey       string             `json:"DescKey"`
	Appearance    string             `json:"Appearance"`
	Weapon        GuildFrigateWeapon `json:"Weapon"`
}

type GuildFrigateConfig struct {
	AttrArr []*GuildFrigateProtype
}

var (
	GGuildFrigateConfig = new(GuildFrigateConfig)
)

func (this *GuildFrigateConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

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

	protypeList := make([]*GuildFrigateProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrArr = append(this.AttrArr, protypeList...)

	sort.Slice(this.AttrArr, func(i, j int) bool { return this.AttrArr[i].Level < this.AttrArr[j].Level })

	base.GLog.Debug("guild frigate data is [%+v]", *this)

	return nil
}

func (this *GuildFrigateConfig) LevelExp(currentLevel int, currentExp int) (newLevel, restExp int) {
	newLevel, restExp = currentLevel, currentExp

	maxLevel := 0
	curIndex := 0
	for index := 0; index < len(this.AttrArr); index++ {
		if maxLevel < this.AttrArr[index].Level {
			maxLevel = this.AttrArr[index].Level
		}

		if currentLevel == this.AttrArr[index].Level {
			curIndex = index
		}
	}

	if currentLevel == maxLevel {
		return
	}

	base.GLog.Debug("LevelExp newLevel %d restExp %d curIndex %d maxLevel %d", newLevel, restExp, curIndex, maxLevel)

	for index := curIndex; index < len(this.AttrArr); index++ {
		base.GLog.Debug("index [%d] and attr value[%v]", index, this.AttrArr[index])
		if restExp >= this.AttrArr[index].NeedExp {
			// 达到本次经验，级别需要多加一级
			newLevel = this.AttrArr[index].Level + 1
			restExp -= this.AttrArr[index].NeedExp
		} else {
			break
		}
	}

	return
}
