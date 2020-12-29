package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sort"
)

type GuildTaskRewardProtype struct {
	Id     int           `json:"Id"`
	Score  int           `json:"Score"`
	Reward ResourcesAttr `josn:"Reward"`
}

type GuildTaskRewardConfig struct {
	AttrArr []*GuildTaskRewardProtype
	AttrMap map[int]*GuildTaskRewardProtype
}

var (
	GGuildTaskRewardConfig = new(GuildTaskRewardConfig)
)

func (this *GuildTaskRewardConfig) Init(configFile string) error {
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

	protypeList := make([]*GuildTaskRewardProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrArr = append(this.AttrArr, protypeList...)
	this.AttrMap = make(map[int]*GuildTaskRewardProtype)
	for _, protype := range protypeList {
		this.AttrMap[protype.Id] = protype
	}

	sort.Slice(this.AttrArr, func(i, j int) bool { return this.AttrArr[i].Score < this.AttrArr[j].Score })

	base.GLog.Debug("guild level config data is [%+v]", *this)

	return nil
}

func (this *GuildTaskRewardConfig) GetTaskConfig(protypeId int) *GuildTaskRewardProtype {
	if value, ok := this.AttrMap[protypeId]; ok {
		return value
	}
	return nil
}

func (this *GuildTaskRewardConfig) RewardBetween(minScore, maxScore int) []*GuildTaskRewardProtype {
	list := make([]*GuildTaskRewardProtype, 0)
	for _, info := range this.AttrArr {
		if info.Score > minScore && info.Score <= maxScore {
			list = append(list, info)
		}
	}

	return list
}
