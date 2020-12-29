package config

import (
	"encoding/json"
	"io/ioutil"
	"sailcraft/base"
)

type AchievementProtype struct {
	Id              int               `json:"Id"`
	AchievementType string            `json:"AchievementType"`
	Parameter       MissionParameters `json:"Parameter"`
}

type AchievementConfig struct {
	AttrMap map[int]*AchievementProtype
	TypeMap map[string][]*AchievementProtype
}

var (
	GAchievementConfig *AchievementConfig = new(AchievementConfig)
)

func (this *AchievementConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	this.AttrMap = make(map[int]*AchievementProtype)
	this.TypeMap = make(map[string][]*AchievementProtype)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		base.GLog.Error("read file '%s' failed err[%s] \n", configFile, err)
		return err
	}

	protypeList := make([]AchievementProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("json unmarshal %s failed err [%s]", configFile, err)
		return err
	}

	for i, protype := range protypeList {
		this.AttrMap[protype.Id] = &protypeList[i]
		// 一个类型对应过个成就实例
		this.TypeMap[protype.AchievementType] = append(this.TypeMap[protype.AchievementType], &protypeList[i])
		base.GLog.Debug("achievement protype data is [%+v]", protype)
	}

	base.GLog.Debug("achievement protype AttrMap size is %d, TypeMap size is %d", len(this.AttrMap), len(this.TypeMap))
	return nil
}
