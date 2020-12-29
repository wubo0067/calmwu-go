package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sort"
)

type GuildLevelProtype struct {
	Level   int `json:"Level"`
	NeedExp int `json:"NeedExp"`
	SumExp  int `json:"Sum"`
}

type GuildLevelConfig struct {
	AttrMap map[int]*GuildLevelProtype
	AttrArr []*GuildLevelProtype
}

var (
	GGuildLevelConfig = new(GuildLevelConfig)
)

func (this *GuildLevelConfig) Init(configFile string) error {
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

	protypeList := make([]*GuildLevelProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrArr = append(this.AttrArr, protypeList...)
	sort.Slice(this.AttrArr, func(i, j int) bool { return this.AttrArr[i].Level < this.AttrArr[j].Level })

	this.AttrMap = make(map[int]*GuildLevelProtype)
	for _, protype := range protypeList {
		this.AttrMap[protype.Level] = protype
	}

	base.GLog.Debug("guild level config data is [%+v]", *this)

	return nil
}

func (this *GuildLevelConfig) LevelBetween(minVitality, maxVitality int) (int, int) {
	minIndex := 0
	for minIndex < len(this.AttrArr) && this.AttrArr[minIndex].SumExp <= minVitality {
		minIndex++
	}

	maxIndex := len(this.AttrArr) - 1
	for maxIndex >= 0 && this.AttrArr[maxIndex].SumExp > maxVitality {
		maxIndex--
	}

	if maxIndex < minIndex {
		return -1, -1
	}

	return this.AttrArr[minIndex].Level, this.AttrArr[maxIndex].Level
}

func (this *GuildLevelConfig) LevelBySumExp(exp int) int {
	minIndex := 1
	for minIndex < len(this.AttrArr) && this.AttrArr[minIndex].SumExp <= exp {
		minIndex++
	}

	return this.AttrArr[minIndex-1].Level
}
