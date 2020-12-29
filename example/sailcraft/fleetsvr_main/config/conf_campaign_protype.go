package config

import (
	"io/ioutil"
	"os"
	"sailcraft/base"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/mitchellh/mapstructure"
)

type CampaignProtype struct {
	Id            int           `mapstructure:"Id"`
	InArea        int           `mapstructure:"InArea"`
	PreCampaignId int           `mapstructure:"PreCampaignId"`
	NameKey       string        `mapstructure:"NameKey"`
	DescKey       string        `mapstructure:"DescKey"`
	Type          int           `mapstructure:"Type"`
	Enemy         int           `mapstructure:"Enemy"`
	Requirement   []int         `mapstructure:"Requirement"`
	EventIds      []int         `mapstructure:"Event"`
	EventChances  []int         `mapstructure:"EventChance"`
	Tribute       ResourcesAttr `mapstructure:"Tribute"`
	TotalChance   int
}

type CampaignConfig struct {
	AttrMap    map[int]*CampaignProtype
	ChapterMap map[int][]*CampaignProtype
}

var (
	GCampaignConfig *CampaignConfig = new(CampaignConfig)
)

func (conf *CampaignConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.AttrMap = make(map[int]*CampaignProtype)
	conf.ChapterMap = make(map[int][]*CampaignProtype)

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
			var attr *CampaignProtype = new(CampaignProtype)
			err = mapstructure.Decode(value, attr)
			if err != nil {
				base.GLog.Error("mapstructure Decode %s failed err %s \n", configFile, err.Error())
				return err
			}

			attr.TotalChance = 0
			for _, chance := range attr.EventChances {
				attr.TotalChance += chance
			}

			conf.AttrMap[attr.Id] = attr
			conf.ChapterMap[attr.InArea] = append(conf.ChapterMap[attr.InArea], attr)

			base.GLog.Debug("campaign protype data is [%+v]", *attr)
		}
	}

	base.GLog.Debug("campaign protype AttrMap size is %d", len(conf.AttrMap))

	return nil
}
