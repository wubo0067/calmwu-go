package config

import (
	"io/ioutil"
	"os"
	"sailcraft/base"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/mitchellh/mapstructure"
)

const (
	CAMPAIGN_MISSION_TYPE_PVP_TIMES     = "league_pvp_times"
	CAMPAIGN_MISSION_TYPE_PVP_WIN_TIMES = "league_pvp_win_times"
	CAMPAIGN_MISSION_TYPE_SINK_SHIP     = "league_pvp_sink_battleships"
)

const (
	CAMPAIGN_MISSION_PARAMETER_TARGET = "target"
)

type CampaignMissionProtype struct {
	Id          int            `json:"Id"`
	Missiontype string         `json:"Missiontype"`
	Parameter   map[string]int `json:"Parameter"`
}

type CampaignMissionConfig struct {
	AttrMap map[int]*CampaignMissionProtype
	TypeMap map[string]([]*CampaignMissionProtype)
}

var (
	GCampaignMissionConfig *CampaignMissionConfig = new(CampaignMissionConfig)
)

func (this *CampaignMissionConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	this.AttrMap = make(map[int]*CampaignMissionProtype)
	this.TypeMap = make(map[string]([]*CampaignMissionProtype))

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
			var attr *CampaignMissionProtype = new(CampaignMissionProtype)
			err = mapstructure.Decode(value, attr)
			if err != nil {
				base.GLog.Error("mapstructure Decode %s failed err %s \n", configFile, err.Error())
				return err
			}

			this.AttrMap[attr.Id] = attr
			this.TypeMap[attr.Missiontype] = append(this.TypeMap[attr.Missiontype], attr)

			base.GLog.Debug("campaign mission protype data is [%+v]", *attr)
		}
	}

	base.GLog.Debug("campaign mission protype AttrMap size is %d", len(this.AttrMap))

	return nil
}
