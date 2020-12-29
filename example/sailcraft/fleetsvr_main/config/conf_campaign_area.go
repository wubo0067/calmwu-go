package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type CampaignAreaProtype struct {
	Id              int    `json:"Id"`
	NameKey         string `json:"NameKey"`
	DescKey         string `json:"DescKey"`
	UnlockLeague    int    `json:"UnlockLeague"`
	ResourceId      string `json:"ResourceId"`
	ResourceBundle  string `json:"ResourceBundle"`
	ResourceIdSmall string `json:"ResourceIdSmall"`
	Icon            string `json:"Icon"`
	PlotHead        int    `json:"PlotHead"`
}

type CampaignAreaConfig struct {
	AttrMap map[int]*CampaignAreaProtype
}

var (
	GCampaignAreaConfig = new(CampaignAreaConfig)
)

func (this *CampaignAreaConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	this.AttrMap = make(map[int]*CampaignAreaProtype)

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

	protypeList := make([]*CampaignAreaProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		return err
	}

	for _, protype := range protypeList {
		this.AttrMap[protype.Id] = protype
	}

	base.GLog.Debug("campaign area config data is [%+v]", *this)

	return nil
}
