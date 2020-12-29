package config

import (
	"io/ioutil"
	"sailcraft/base"
	"sort"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/mitchellh/mapstructure"
)

type ActivityScoreRewardProtype struct {
	Id     string        `json:"Id"`
	Score  int           `json:"Score"`
	Reward ResourcesAttr `json:"Reward"`
}

type ActivityScoreRewardConfig struct {
	AttrMap       map[string]*ActivityScoreRewardProtype
	AttrSortedArr []*ActivityScoreRewardProtype
}

var (
	GActivityScoreRewardConfig *ActivityScoreRewardConfig = new(ActivityScoreRewardConfig)
)

func (this *ActivityScoreRewardConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig %s", configFile)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		base.GLog.Error("read file '%s' failed err [%s]", configFile, err)
		return err
	}

	this.AttrMap = make(map[string]*ActivityScoreRewardProtype)
	this.AttrSortedArr = make([]*ActivityScoreRewardProtype, 0)

	list := arraylist.New()
	err = list.FromJSON(data)
	if err != nil {
		base.GLog.Error("arraylist parse json %s failed err %s \n", configFile, err.Error())
		return err
	}

	for index := 0; index < list.Size(); index++ {
		value, bRet := list.Get(index)
		if bRet {
			attr := new(ActivityScoreRewardProtype)
			err = mapstructure.Decode(value, attr)
			if err != nil {
				base.GLog.Error("mapstructure Decode %s failed err %s \n", configFile, err.Error())
				return err
			}

			this.AttrMap[attr.Id] = attr
			this.AttrSortedArr = append(this.AttrSortedArr, attr)

			base.GLog.Debug("ActivityScoreReward protype data is [%+v]", attr)
		}
	}

	sort.Slice(this.AttrSortedArr, func(i, j int) bool { return this.AttrSortedArr[i].Score < this.AttrSortedArr[j].Score })

	base.GLog.Debug("ActivityScoreRewardConfig AttrMap size is %d", len(this.AttrMap))

	return nil
}

func (this *ActivityScoreRewardConfig) GetScoreReward(oldScore, newScore int) []*ActivityScoreRewardProtype {
	if oldScore >= newScore {
		return nil
	}

	list := make([]*ActivityScoreRewardProtype, 0)
	for _, info := range this.AttrSortedArr {
		if info.Score > oldScore && info.Score <= newScore {
			list = append(list, info)
		}
	}

	return list
}
