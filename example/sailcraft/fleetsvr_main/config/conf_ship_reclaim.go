package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type ShipReclaimProtype struct {
	Id                   int           `json:"Id"`
	Quality              string        `json:"Quality"`
	DowngradeCost        ResourcesAttr `json:"DowngradeCost"`
	ResourcesReturnRatio int           `json:"ResourcesReturnRatio"`
	SalePrices           ResourcesAttr `json:"SalePrices"`
}

type ShipReclaimConfig struct {
	AttrMap map[string]*ShipReclaimProtype
}

var (
	GShipReclaimConfig = new(ShipReclaimConfig)
)

func (conf *ShipReclaimConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	conf.AttrMap = make(map[string]*ShipReclaimProtype)

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

	protypeList := make([]*ShipReclaimProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("json unmarshal %s failed! reason[%s]\n", configFile, err.Error())
		return err
	}

	for _, protype := range protypeList {
		conf.AttrMap[protype.Quality] = protype
	}

	base.GLog.Debug("ship reclaim config data is [%+v]", *conf)

	return nil
}

func (conf *ShipReclaimConfig) ReclaimReward(quality string, count int) (*ResourcesAttr, error) {
	if protype, ok := conf.AttrMap[quality]; ok {
		var reward ResourcesAttr
		reward.Add(&protype.SalePrices)
		reward.Scale(float64(count))

		return &reward, nil
	} else {
		return nil, fmt.Errorf("Can not find ShipReclaimProtype[%s]", quality)
	}
}

func (conf *ShipReclaimConfig) GetReclaimProtype(quality string) (*ShipReclaimProtype, error) {
	if protype, ok := conf.AttrMap[quality]; ok {
		return protype, nil
	} else {
		return nil, fmt.Errorf("Can not find ShipReclaimProtype[%s]", quality)
	}
}
