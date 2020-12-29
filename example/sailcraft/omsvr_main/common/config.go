package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sync"
)

type ConfigS struct {
	SysConfPath string `json:"SysConfPath"`
}

type ConfigMgr struct {
	config     *ConfigS
	configFile string
	monitor    *sync.RWMutex
}

var (
	GConfig *ConfigMgr
)

func init() {
	GConfig = new(ConfigMgr)
}

func (configMgr *ConfigMgr) Init(configFile string) error {
	configMgr.config = new(ConfigS)
	configMgr.configFile = configFile
	configMgr.monitor = new(sync.RWMutex)

	return configMgr.ParseConfig()
}

func (configMgr *ConfigMgr) ParseConfig() error {
	conf_file, err := os.Open(configMgr.configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open [%s] failed! err[%s]\n", configMgr.configFile, err.Error())
		return err
	}
	defer conf_file.Close()

	data, err := ioutil.ReadAll(conf_file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, configMgr.config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal [%s] file failed! err[%s]\n", configMgr.configFile, err.Error())
		return err
	}

	base.GLog.Debug("config data: [%+v]", *(configMgr.config))

	return nil
}

func (configMgr *ConfigMgr) ReloadConfig() string {
	configMgr.monitor.Lock()
	defer configMgr.monitor.Unlock()

	return "Reload OK!"
}

func (configMgr *ConfigMgr) GetSysConfPath() string {
	configMgr.monitor.RLock()
	defer configMgr.monitor.RUnlock()

	return configMgr.config.SysConfPath
}
