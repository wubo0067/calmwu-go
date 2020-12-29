/*
 * @Author: calmwu
 * @Date: 2018-10-24 17:06:32
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-24 17:56:45
 */

package routersvr

import (
	base "doyo-server-go/doyo-base-go"
	"encoding/json"
	"io/ioutil"
	"os"
)

type RouterSvrKafkaConf struct {
	Brokers string `json:"Brokers"`
	GroupID string `json:"GroupID"`
}

type RouterSvrConsulConf struct {
	ConsulListenAddr string `json:"ConsulListenAddr"`
}

type RouterSvrConf struct {
	KfaConf              RouterSvrKafkaConf  `json:"Kafka"`
	DispatchRoutineCount int                 `json:"DispatchRoutineCount"`
	ConsulConf           RouterSvrConsulConf `json:"Consul"`
}

var (
	confMgr *RouterSvrConf
)

func loadConfig(configFile string) error {
	// reload也会重新加载
	confMgr = new(RouterSvrConf)

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

	return nil
}
