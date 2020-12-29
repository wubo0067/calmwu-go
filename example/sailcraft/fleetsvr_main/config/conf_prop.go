package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"

	"github.com/mitchellh/mapstructure"
)

const (
	PROP_TYPE_PIECES            = "SalvagePiece"
	PROP_TYPE_NET               = "Net"
	PROP_TYPE_FORSELL           = "Forsell"
	PROP_TYPE_CHEST             = "Chest"
	PROP_TYPE_ALCOHOL           = "Alcohol"
	PROP_TYPE_RAND_CHEST        = "RandChest"
	PROP_TYPE_CARDPACK          = "Cardpack"
	PROP_TYPE_GUILD_FRIGATE_EXP = "GuildFriExp"
)

type PropNetEffect struct {
	PiecesChance int `json:"PiecesChance"`
	TimesLower   int `json:"TimesLower"`
	TimesUpper   int `json:"TimesUpper"`
}

type PropGuildFriExpEffect struct {
	Exp int `json:"Exp"`
}

type PropProtype struct {
	Id       int         `json:"Id"`
	PropType string      `json:"PropType"`
	Effect   interface{} `json:"Effect"`
}

type PropConfig struct {
	AttrArr []*PropProtype
	AttrMap map[int]*PropProtype
}

var (
	GPropConfig = new(PropConfig)
)

func (this *PropConfig) Init(configFile string) error {
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

	protypeList := make([]*PropProtype, 0)
	err = json.Unmarshal(data, &protypeList)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	this.AttrArr = append(this.AttrArr, protypeList...)

	this.AttrMap = make(map[int]*PropProtype)
	for _, protype := range protypeList {
		this.AttrMap[protype.Id] = protype
	}

	base.GLog.Debug("protype config data is [%+v]", *this)

	return nil
}

func (this *PropConfig) DecodeEffect(protype *PropProtype, rawVal interface{}) error {
	if protype == nil || rawVal == nil {
		return fmt.Errorf("null point")
	}

	config := &mapstructure.DecoderConfig{
		TagName:  "json",
		Metadata: nil,
		Result:   rawVal,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(protype.Effect)
}

func (this *PropConfig) GetPropByType(propTypes ...string) []*PropProtype {
	propArr := make([]*PropProtype, 0)
	if len(propTypes) > 0 {
		typeMap := make(map[string]string)
		for _, propType := range propTypes {
			typeMap[propType] = propType
		}

		for _, protype := range this.AttrArr {
			if _, ok := typeMap[protype.PropType]; ok {
				propArr = append(propArr, protype)
			}
		}
	}

	return propArr
}
