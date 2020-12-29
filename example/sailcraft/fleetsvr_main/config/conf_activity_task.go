package config

import (
	"encoding/json"
	"io/ioutil"
	"sailcraft/base"
)

type ActivityTaskProtype struct {
	Id         int               `json:"Id"`
	TaskType   string            `json:"TaskType"`
	Parameters MissionParameters `json:"Parameter"`
	LevelLimit int               `json:"LevelLimit"`
	NameKey    string            `json:"NameKey"`
	DescKey    string            `json:"DescKey"`
	Reward     ResourcesAttr     `json:"Reward"`
}

type ActivityTaskConfig struct {
	AttrMap map[int]*ActivityTaskProtype
	TypeMap map[string][]*ActivityTaskProtype
}

var (
	GActivityTaskConfig *ActivityTaskConfig = new(ActivityTaskConfig)
)

func (this *ActivityTaskConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig %s", configFile)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		base.GLog.Error("read file '%s' failed err[%s]", configFile, err)
		return err
	}

	this.AttrMap = make(map[int]*ActivityTaskProtype)
	this.TypeMap = make(map[string][]*ActivityTaskProtype)

	protypeList := make([]ActivityTaskProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("json unmarshal %s failed err %s", configFile, err)
		return err
	}

	for i, protype := range protypeList {
		this.AttrMap[protype.Id] = &protypeList[i]
		this.TypeMap[protype.TaskType] = append(this.TypeMap[protype.TaskType], &protypeList[i])

		base.GLog.Debug("activity task protype data is [%+v]", protype)
	}

	base.GLog.Debug("activity protype AttrMap size is %d, TypeMap size is %d", len(this.AttrMap), len(this.TypeMap))
	return nil
}
