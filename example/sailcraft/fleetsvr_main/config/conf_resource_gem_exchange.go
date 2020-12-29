package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type ResourceGemExchangeAttr struct {
	IronRanges  []int `json:"iron_ranges"`
	StoneRanges []int `json:"stone_ranges"`
	WoodRanges  []int `json:"wood_ranges"`
	GoldRanges  []int `json:"gold_ranges"`
	GemRanges   []int `json:"gem_ranges"`
}

type ResourceGemExchangeConfig struct {
	Attr *ResourceGemExchangeAttr
}

var (
	GResourceGemExchangeConfig *ResourceGemExchangeConfig = new(ResourceGemExchangeConfig)
)

func (conf *ResourceGemExchangeConfig) GetGemExchangeAttr() (ResourceGemExchangeAttr, error) {
	if conf.Attr == nil {
		return ResourceGemExchangeAttr{}, fmt.Errorf("ResourceGemExchangeConfig null point")
	}

	return *conf.Attr, nil
}

func (conf *ResourceGemExchangeConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.Attr = new(ResourceGemExchangeAttr)

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

	err = json.Unmarshal(data, conf.Attr)
	if err != nil {
		base.GLog.Error("json.Unmarshal json %s failed err %s \n", configFile, err.Error())
		return err
	}

	base.GLog.Debug("load resource gem exchange config result is [%+v]", *conf.Attr)

	return nil
}
