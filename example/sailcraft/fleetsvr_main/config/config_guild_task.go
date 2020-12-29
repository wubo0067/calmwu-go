package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type GuildTaskProtype struct {
	Id         int               `json:"Id"`
	Score      int               `json:"Score"`
	Parameters MissionParameters `json:"Parameter"`
	TaskType   string            `json:"TaskType"`
}

type GuildTaskConfig struct {
	AttrArr []*GuildTaskProtype
	AttrMap map[string][]*GuildTaskProtype
}

var (
	GGuildTaskConfig = new(GuildTaskConfig)
)

func (this *GuildTaskConfig) Init(configFile string) error {
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

	protypeList := make([]*GuildTaskProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrArr = append(this.AttrArr, protypeList...)
	this.AttrMap = make(map[string][]*GuildTaskProtype)

	for _, protype := range protypeList {
		this.AttrMap[protype.TaskType] = append(this.AttrMap[protype.TaskType], protype)
	}

	base.GLog.Debug("guild level config data is [%+v]", *this)

	return nil
}
