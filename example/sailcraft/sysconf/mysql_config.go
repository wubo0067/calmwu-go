package sysconf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type MysqlAttr struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type MysqlRawConfig struct {
	PlatformSet    MysqlAttr `json:"sailcraft_platform_set"`
	UinSet1To100   MysqlAttr `json:"sailcraft_uin_set_1_1000000"`
	UinSet100To200 MysqlAttr `json:"sailcraft_uin_set_1000001_2000000"`
	UinSet200To300 MysqlAttr `json:"sailcraft_uin_set_2000001_3000000"`
	UinSet300To400 MysqlAttr `json:"sailcraft_uin_set_3000001_4000000"`
	UinSet400to500 MysqlAttr `json:"sailcraft_uin_set_4000001_5000000"`
	UserFinance    MysqlAttr `json:"user_finance"`
	OmsDB          MysqlAttr `json:"omsdb"`
}

type MysqlConfig struct {
	ConfigMap map[string]*MysqlAttr
}

var (
	GMysqlConfig *MysqlConfig = new(MysqlConfig)
)

func (conf *MysqlConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	hFile, err := os.Open(configFile)
	if err != nil {
		base.GLog.Error("open [%s] failed! err[%s]\n", configFile, err.Error())
		return err
	}
	defer hFile.Close()

	data, err := ioutil.ReadAll(hFile)
	if err != nil {
		base.GLog.Error("read [%s] failed! err[%s]\n", configFile, err.Error())
		return err
	}

	var rawConfig MysqlRawConfig = MysqlRawConfig{}

	err = json.Unmarshal(data, &rawConfig)
	if err != nil {
		base.GLog.Error("unmarshal [%s] file failed! err[%s]\n", configFile, err.Error())
		return err
	}

	conf.ConfigMap = make(map[string]*MysqlAttr)
	conf.ConfigMap[rawConfig.PlatformSet.Database] = &rawConfig.PlatformSet
	conf.ConfigMap[rawConfig.UinSet1To100.Database] = &rawConfig.UinSet1To100
	conf.ConfigMap[rawConfig.UinSet100To200.Database] = &rawConfig.UinSet100To200
	conf.ConfigMap[rawConfig.UinSet200To300.Database] = &rawConfig.UinSet200To300
	conf.ConfigMap[rawConfig.UinSet300To400.Database] = &rawConfig.UinSet300To400
	conf.ConfigMap[rawConfig.UinSet400to500.Database] = &rawConfig.UinSet400to500
	conf.ConfigMap[rawConfig.UserFinance.Database] = &rawConfig.UserFinance
	conf.ConfigMap[rawConfig.OmsDB.Database] = &rawConfig.OmsDB

	base.GLog.Debug("MysqlConfig Config[%+v]", *conf)

	return nil
}
