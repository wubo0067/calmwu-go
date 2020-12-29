package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sort"
)

type FrigateWeaponLevelProtype struct {
	AssetBundle         string        `json:"AssetBundle"`
	Available           int           `json:"Available"`
	Cost                ResourcesAttr `json:"Cost"`
	Level               int           `json:"Level"`
	RequireFrigateLevel int           `json:"RequireFrigateLevel"`
	Skill               int           `json:"Skill"`
}

type FrigateWeaponProtype struct {
	Id           int                          `json:"Id"`
	LevelProtype []*FrigateWeaponLevelProtype `json:"Level"`
}

type FrigateWeaponConfig struct {
	AttrMap map[int]*FrigateWeaponProtype
}

var (
	GFrigateWeaponConfig = new(FrigateWeaponConfig)
)

func (this *FrigateWeaponConfig) Init(configFile string) error {
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

	protypeList := make([]*FrigateWeaponProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrMap = make(map[int]*FrigateWeaponProtype)
	for _, protype := range protypeList {
		sort.Slice(protype.LevelProtype, func(i, j int) bool { return protype.LevelProtype[i].Level < protype.LevelProtype[j].Level })
		this.AttrMap[protype.Id] = protype
	}

	base.GLog.Debug("frigate weapon data is [%+v]", *this)

	return nil
}
