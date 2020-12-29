package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type IconFrameProtype struct {
	Id              int    `json:"Id"`
	NameKey         string `json:"NameKey"`
	ResourceId      string `json:"ResourceId"`
	Serial          int    `json:"Serial"`
	Term            int    `json:"Term"`
	UnlockCondition string `json:"UnlockCondition"`
}

type IconFrameConfig struct {
	AttrMap map[int]*IconFrameProtype
}

var (
	GIconFrameConfig = new(IconFrameConfig)
)

func (this *IconFrameConfig) Init(configFile string) error {
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

	protypeList := make([]*IconFrameProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrMap = make(map[int]*IconFrameProtype)
	for _, protype := range protypeList {
		this.AttrMap[protype.Id] = protype
	}

	base.GLog.Debug("head frame config data is [%+v]", *this)

	return nil
}
