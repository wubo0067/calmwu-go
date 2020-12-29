package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

const (
	POOL_TYPE_PROPS = "props"
	POOL_TYPE_CARDS = "cards"
)

type GuildSalvagePoolProtype struct {
	Id          int    `json:"Id"`
	PoolType    string `json:"PoolType"`
	Count       []int  `json:"Count"`
	Content     []int  `json:"Content"`
	Weight      []int  `json:"Weight"`
	TotalWeight int
}

type GuildSalvagePoolConfig struct {
	AttrMap map[int]*GuildSalvagePoolProtype
}

var (
	GGuildSalvagePoolConfig = new(GuildSalvagePoolConfig)
)

func (this *GuildSalvagePoolConfig) Init(configFile string) error {
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

	protypeList := make([]*GuildSalvagePoolProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrMap = make(map[int]*GuildSalvagePoolProtype)
	for _, protype := range protypeList {
		this.AttrMap[protype.Id] = protype

		protype.TotalWeight = 0
		for _, weight := range protype.Weight {
			protype.TotalWeight += weight
		}
	}

	base.GLog.Debug("guild salvage pool data is [%+v]", *this)

	return nil
}
