package config

import (
	"io/ioutil"
	"os"
	"sailcraft/base"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/mitchellh/mapstructure"
)

const (
	CAMPAIGN_EVENT_TYPE_FREE     = "free"
	CAMPAIGN_EVENT_TYPE_MISSION  = "finish_mission"
	CAMPAIGN_EVENT_TYPE_EXCHANGE = "exchange"
)

type CampaignEventProtype struct {
	Id          int           `mapstructure:"Id"`
	EventType   string        `mapstructure:"EventType"`
	DescKey     string        `mapstructure:"DescKey"`
	MissionId   int           `mapstructure:"MissionId"`
	NameKey     string        `mapstructure:"NameKey"`
	Requirement ResourcesAttr `mapstructure:"Requirement"`
	Reward      ResourcesAttr `mapstructure:"Reward"`
	Term        int           `mapsturcture:"Term"`
}

type CampaignEventConfig struct {
	AttrMap map[int]*CampaignEventProtype
}

var (
	GCampaignEventConfig *CampaignEventConfig = new(CampaignEventConfig)
)

func (this *CampaignEventConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	this.AttrMap = make(map[int]*CampaignEventProtype)

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

	list := arraylist.New()
	err = list.FromJSON(data)
	if err != nil {
		base.GLog.Error("arraylist parse json %s failed err %s \n", configFile, err.Error())
		return err
	}

	for index := 0; index < list.Size(); index++ {
		value, bRet := list.Get(index)
		if bRet {
			var attr *CampaignEventProtype = new(CampaignEventProtype)
			err = mapstructure.Decode(value, attr)
			if err != nil {
				base.GLog.Error("mapstructure Decode %s failed err %s \n", configFile, err.Error())
				return err
			}

			this.AttrMap[attr.Id] = attr

			base.GLog.Debug("campaign event protype data is [%+v]", *attr)
		}
	}

	base.GLog.Debug("campaign event protype AttrMap size is %d", len(this.AttrMap))

	return nil
}
