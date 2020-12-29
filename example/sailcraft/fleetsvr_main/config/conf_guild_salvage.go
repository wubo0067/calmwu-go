package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type GuildSalvageProtype struct {
	Id           int   `json:"Id"`
	PiecesPool   int   `json:"PiecesPool"`
	PiecesChance int   `json:"Chance"`
	Pools        []int `json:"Pools"`
	TimesLower   []int `json:"TimesLower"`
	TimesUpper   []int `json:"TimesUpper"`
}

type GuildSalvageConfig struct {
	AttrArr []*GuildSalvageProtype
	AttrMap map[int]*GuildSalvageProtype
}

var (
	GGuildSalvageConfig = new(GuildSalvageConfig)
)

func (this *GuildSalvageConfig) Init(configFile string) error {
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

	protypeList := make([]*GuildSalvageProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrArr = append(this.AttrArr, protypeList...)

	this.AttrMap = make(map[int]*GuildSalvageProtype)
	for _, protype := range protypeList {
		this.AttrMap[protype.Id] = protype
	}

	base.GLog.Debug("guild salvage config data is [%+v]", *this)

	return nil
}
