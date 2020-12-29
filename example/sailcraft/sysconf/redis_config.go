package sysconf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type RedisConfig struct {
	ClusterRedisAddressList []string `json:"ClusterRedisAddressList"`
	SingletonRedisAddrsss   string   `json:"SingletonRedisAddrsss"`
}

var (
	GRedisConfig *RedisConfig = new(RedisConfig)
)

func (conf *RedisConfig) Init(configFile string) error {
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

	err = json.Unmarshal(data, conf)
	if err != nil {
		base.GLog.Error("unmarshal [%s] file failed! err[%s]\n", configFile, err.Error())
		return err
	}

	base.GLog.Debug("RedisConfig Config[%+v]", *conf)
	return nil
}
