package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sync"

	"github.com/emirpasic/gods/sets/hashset"
)

type VersionInfoS struct {
	CurrVersions []string `json:"CurrVersions"`
	Channel      string   `json:"Channel"`
	UpdateUrl    string   `json:"UpdateUrl"`
	VersionSet   *hashset.Set
}

type VersionConfigS struct {
	VersionInfo []VersionInfoS `json:"VersionInfo"`
}

type NavigationEnvUrlS struct {
	URLLoginCheck string   `json:"UrlLoginCheck"`
	URLProxySvr   string   `json:"UrlProxySvr"`
	URLSdkSvr     string   `json:"UrlSdkSvr"`
	VersionList   []string `json:"VersionList"`
}

func (neu *NavigationEnvUrlS) checkVersion(clientVersion string) bool {
	for _, version := range neu.VersionList {
		if clientVersion == version {
			return true
		}
	}
	return false
}

type NavigationConfigS struct {
	ProductEnv  NavigationEnvUrlS `json:"ProductEnv"`
	AuditEnv    NavigationEnvUrlS `json:"AuditEnv"`
	StageEnv    NavigationEnvUrlS `json:"StageEnv"`
	Test8885Env NavigationEnvUrlS `json:"8885Env"`
	Test8889Env NavigationEnvUrlS `json:"8889Env"`
	DomesticEnv NavigationEnvUrlS `json:"DomesticEnv"`
}

func (nc *NavigationConfigS) CheckVersion(clientVersion string) *NavigationEnvUrlS {
	if nc.ProductEnv.checkVersion(clientVersion) {
		return &nc.ProductEnv
	}

	if nc.AuditEnv.checkVersion(clientVersion) {
		return &nc.AuditEnv
	}

	if nc.StageEnv.checkVersion(clientVersion) {
		return &nc.StageEnv
	}

	if nc.Test8885Env.checkVersion(clientVersion) {
		return &nc.Test8885Env
	}

	if nc.Test8889Env.checkVersion(clientVersion) {
		return &nc.Test8889Env
	}

	if nc.DomesticEnv.checkVersion(clientVersion) {
		return &nc.DomesticEnv
	}

	base.GLog.Error("clientVersion[%s] is not config in NavigationConfig", clientVersion)
	return nil
}

type ConfigS struct {
	Token            string            `json:"Token"`
	VersionConfig    VersionConfigS    `json:"VersionConfig"`
	NavigationConfig NavigationConfigS `json:"NavigationConfig"`
	SysConfPath      string            `json:"SysConfPath"`
	GMKey            string            `json:"GMKey"`
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

	for index, _ := range configMgr.config.VersionConfig.VersionInfo {
		versionInfo := &configMgr.config.VersionConfig.VersionInfo[index]
		if len(versionInfo.CurrVersions) == 0 {
			err := fmt.Errorf("CurrVersions is empty!")
			base.GLog.Error(err.Error())
			return err
		}

		versionInfo.VersionSet = hashset.New()
		for _, version := range versionInfo.CurrVersions {
			versionInfo.VersionSet.Add(version)
		}
		base.GLog.Debug("Channel[%s] VersionSet size[%d]", versionInfo.Channel, versionInfo.VersionSet.Size())
	}

	base.GLog.Debug("config data: [%+v]", *(configMgr.config))

	return nil
}

func (configMgr *ConfigMgr) ReloadConfig() string {
	configMgr.monitor.Lock()
	defer configMgr.monitor.Unlock()

	err := configMgr.ParseConfig()
	if err != nil {
		return err.Error()
	}

	return "Reload OK!"
}

func (configMgr *ConfigMgr) GetToken() string {
	configMgr.monitor.RLock()
	defer configMgr.monitor.RUnlock()

	return configMgr.config.Token
}

func (configMgr *ConfigMgr) GetVersionInfo(channelName string) *VersionInfoS {
	configMgr.monitor.RLock()
	defer configMgr.monitor.RUnlock()

	for index, _ := range configMgr.config.VersionConfig.VersionInfo {
		versionInfo := &configMgr.config.VersionConfig.VersionInfo[index]
		if versionInfo.Channel == channelName {
			base.GLog.Debug("index[%d] versionInfo[%v]", index, versionInfo)
			return versionInfo
		}
	}
	return nil
}

func (configMgr *ConfigMgr) GetNavigationConfig() *NavigationConfigS {
	configMgr.monitor.RLock()
	defer configMgr.monitor.RUnlock()

	return &configMgr.config.NavigationConfig
}

func (configMgr *ConfigMgr) GetSysConfPath() string {
	configMgr.monitor.RLock()
	defer configMgr.monitor.RUnlock()

	return configMgr.config.SysConfPath
}

func (configMgr *ConfigMgr) GetGMKey() string {
	configMgr.monitor.RLock()
	defer configMgr.monitor.RUnlock()

	return configMgr.config.GMKey
}
