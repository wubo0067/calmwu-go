package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/mitchellh/mapstructure"
)

const (
	LEVEL_EXP_UNLOCK_GUILD        = "guild"
	LEVEL_EXP_UNLOCK_BREAK_OUT    = "breakout"
	LEVEL_EXP_UNLOCK_SHIP_RECLAIM = "shipreclaim"
	LEVEL_EXP_UNLOCK_TASK         = "task"
)

type LevelExpAttr struct {
	Level         int           `mapstructure:"Level"`
	Exp           int           `mapstructure:"NeedExp"`
	Sum           int           `mapstructure:"Sum"`
	Reward        ResourcesAttr `mapstructure:"Reward"`
	Unlock        string        `mapstructure:"Unlock"`
	UnlockNameKey string        `mapstructure:"UnlockNameKey"`
}

type LevelExpConfig struct {
	AttrList  []*LevelExpAttr
	UnlockMap map[string]*LevelExpAttr
}

var (
	GLevelExpConfig *LevelExpConfig = new(LevelExpConfig)
)

func (conf *LevelExpConfig) GetLevelExpAttr(level int) (LevelExpAttr, error) {
	for _, attr := range conf.AttrList {
		if attr.Level == level {
			return *attr, nil
		}
	}

	return LevelExpAttr{}, fmt.Errorf("config is not exist level %d", level)
}

func (conf *LevelExpConfig) QueryLevelByExp(exp int) int {
	if len(conf.AttrList) < 1 {
		return 0
	}

	for index := 0; index < len(conf.AttrList)-1; index++ {
		attr := conf.AttrList[index]
		nextAttr := conf.AttrList[index+1]
		if exp >= attr.Sum && exp < nextAttr.Sum {
			return attr.Level
		}
	}

	return conf.AttrList[len(conf.AttrList)-1].Level
}

func (conf *LevelExpConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.AttrList = make([]*LevelExpAttr, 0)
	conf.UnlockMap = make(map[string]*LevelExpAttr)

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
			var attr *LevelExpAttr = new(LevelExpAttr)
			err = mapstructure.Decode(value, attr)
			if err != nil {
				base.GLog.Error("mapstructure Decode %s failed err %s \n", configFile, err.Error())
				return err
			}

			conf.AttrList = append(conf.AttrList, attr)
			if attr.Unlock != "" {
				conf.UnlockMap[attr.Unlock] = attr
			}
		}
	}

	base.GLog.Debug("load user level exp config result length is %d", len(conf.AttrList))

	return nil
}
