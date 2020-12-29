package config

import (
	"encoding/json"
	"io/ioutil"
	"sailcraft/base"
)

type GrowUpTaskProtype struct {
	Id          string            `json:"Id"`
	PreTaskId   string            `json:"PreTaskId"`
	NextTaskIds []string          `json:"NextTaskIds"`
	TaskType    string            `json:"TaskType"`
	Parameter   MissionParameters `json:"Parameter"`
	Reward      ResourcesAttr     `json:"Reward"`
	LevelLimit  int               `json:"LevelLimit"`
}

type GrowupTaskConfig struct {
	AttrMap map[string]*GrowUpTaskProtype
	TypeMap map[string][]*GrowUpTaskProtype
}

var (
	GGrowupTaskConfig *GrowupTaskConfig = new(GrowupTaskConfig)
)

func (this *GrowupTaskConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig %s", configFile)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		base.GLog.Error("read file %s failed err %s", configFile, err)
		return err
	}

	this.AttrMap = make(map[string]*GrowUpTaskProtype)
	this.TypeMap = make(map[string][]*GrowUpTaskProtype)

	protypeList := make([]GrowUpTaskProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("json unmarshal %s failed err %s", configFile, err)
		return err
	}

	for i, protype := range protypeList {
		this.AttrMap[protype.Id] = &protypeList[i]
		this.TypeMap[protype.TaskType] = append(this.TypeMap[protype.TaskType], &protypeList[i])
		base.GLog.Debug("protype data is [%+v]", protype)
	}

	for _, protype := range this.AttrMap {
		for _, nextId := range protype.NextTaskIds {
			if nextProtype, ok := this.AttrMap[nextId]; ok {
				nextProtype.PreTaskId = protype.Id
			}
		}
	}

	base.GLog.Debug("GrowUpTaskConfig AttrMap size is %d", len(this.AttrMap))
	return nil
}
