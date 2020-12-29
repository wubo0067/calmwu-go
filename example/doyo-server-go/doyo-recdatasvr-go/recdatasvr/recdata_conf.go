/*
 * @Author: calmwu
 * @Date: 2018-11-01 11:20:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-07 19:33:17
 */

package recdatasvr

import (
	base "doyo-server-go/doyo-base-go"
	"encoding/json"
	"io/ioutil"
	"os"
)

type DoyoRecDataSvrKafkaConf struct {
	Brokers string `json:"Brokers"`
	GroupID string `json:"GroupID"`
}

type DoyoRecDataSvrHealthCheckConf struct {
	ConsulHost string `json:"ConsulHost"`
	CheckHost  string `json:"CheckHost"`
	CheckPort  int    `json:"CheckPort"`
}

type DoyoRecDataSvrRedisConf struct {
	ServerAddrs  string `json:"ServerAddrs"`
	Password     string `json:"Password"`
	SessionCount int    `json:"SessionCount"`
	IsCluster    int    `json:"IsCluster"`
}

type DoyoRecDataSvrConf struct {
	KafkaConf       DoyoRecDataSvrKafkaConf       `json:"Kafka"`
	HealthCheckConf DoyoRecDataSvrHealthCheckConf `json:"HealthCheck"`
	RedisConf       DoyoRecDataSvrRedisConf       `json:"Redis"`
	Countries       string                        `json:"Countries"`
}

var (
	confMgr *DoyoRecDataSvrConf
)

func loadConfig(configFile string) error {
	confMgr = new(DoyoRecDataSvrConf)

	confFile, err := os.Open(configFile)
	if err != nil {
		base.ZLog.Errorf("open [%s] failed! reason[%s]\n", configFile, err.Error())
		return err
	}
	defer confFile.Close()

	confData, err := ioutil.ReadAll(confFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(confData, confMgr)
	if err != nil {
		base.ZLog.Errorf("unmarshal [%s] file failed! reason[%s]\n", configFile, err.Error())
		return err
	}
	base.ZLog.Debugf("%+v", *confMgr)

	return nil
}
