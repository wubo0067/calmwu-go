/*
 * @Author: calmwu
 * @Date: 2018-01-10 16:34:32
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-11 12:19:53
 * @Comment:
 */

package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sync"

	geoip2 "github.com/oschwald/geoip2-golang"
)

type CassandraConfS struct {
	ClusterHosts             []string `json:"ClusterHosts"`
	KeySpaces                []string `json:"KeySpaces"`
	WorkerRoutingCount       int      `json:"WorkerRoutingCount"`
	DisableInitialHostLookup int      `json:"DisableInitialHostLookup"`
	Consistency              uint16   `json:"Consistency"`
}

type ConfigS struct {
	CassandraConf CassandraConfS `json:"CassandraConf"`
	GeoIpConf     string         `json:"GeoIpMmdb"`
}

type ConfigMgr struct {
	ConfigData *ConfigS
	ConfigFile string
	monitor    *sync.RWMutex
	geoDB      *geoip2.Reader
}

var (
	GConfig *ConfigMgr
)

func init() {
	GConfig = new(ConfigMgr)
}

func (configMgr *ConfigMgr) Init(configFile string) error {
	configMgr.ConfigData = new(ConfigS)
	configMgr.ConfigFile = configFile
	configMgr.monitor = new(sync.RWMutex)

	return configMgr.ParseConfig()
}

func (configMgr *ConfigMgr) ParseConfig() error {
	conf_file, err := os.Open(configMgr.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open [%s] failed! err[%s]\n", configMgr.ConfigFile, err.Error())
		return err
	}
	defer conf_file.Close()

	data, err := ioutil.ReadAll(conf_file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, configMgr.ConfigData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal [%s] file failed! err[%s]\n", configMgr.ConfigFile, err.Error())
		return err
	}

	base.GLog.Debug("Config:%+v", configMgr.ConfigData)

	// 打开geodb
	configMgr.geoDB, err = geoip2.Open(configMgr.ConfigData.GeoIpConf)
	if err != nil {
		base.GLog.Error("open GeoIPDB[%s] file failed! err[%s]\n", configMgr.ConfigData.GeoIpConf, err.Error())
		return err
	} else {
		base.GLog.Debug("open GeoIPDB[%s] file successed!\n", configMgr.ConfigData.GeoIpConf)
	}
	return nil
}
