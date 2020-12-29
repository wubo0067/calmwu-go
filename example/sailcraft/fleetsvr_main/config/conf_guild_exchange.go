package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type AncientRelicsProtype struct {
	Id       int           `json:"Id"`
	NeedKeys []int         `json:"NeedKeys"`
	Reward   ResourcesAttr `json:"Reward"`
}

type AncientRelicsConfig struct {
	AttrArr []*AncientRelicsProtype
	AttrMap map[int]*AncientRelicsProtype
}

var (
	GAncientRelicsConfig = new(AncientRelicsConfig)
)

func (this *AncientRelicsConfig) Init(configFile string) error {
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

	protypeList := make([]*AncientRelicsProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrArr = append(this.AttrArr, protypeList...)
	this.AttrMap = make(map[int]*AncientRelicsProtype)
	for _, protype := range protypeList {
		this.AttrMap[protype.Id] = protype
	}

	base.GLog.Debug("guild exchange data is [%+v]", *this)

	return nil
}
