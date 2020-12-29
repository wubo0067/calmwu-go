package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sync"
	"time"
)

const (
	ShopRechargeCommoditiesKeyFmt      = "ShopRecharge-Zone%d-Version%s"
	ShopRechargeCommodityVersionKeyFmt = "ShopRecharge-Zone%d"

	ShopResourceCommoditiesKeyFmt      = "ShopResource-Zone%d-Version%s"
	ShopResourceCommodityVersionKeyFmt = "ShopResource-Zone%d"

	ShopCardPackCommoditiesKeyFmt      = "ShopCardPack-Zone%d-Version%s"
	ShopCardPackCommodityVersionKeyFmt = "ShopCardPack-Zone%d"

	GameShopWindowConfigKeyFmt = "RefreshShopConfig-%s" // 每一种商店在redis中的key

	MonthlyCardDuration = 30 * 24 * time.Hour // 月卡时间
	DayDuration         = 24 * time.Hour
	//WeeklyCardDuration  = 7 * 24 * time.Hour  // 普通月卡有效时间

	RefreshShopConfigKeyFmt = "RefreshShopConfig-Zone%d"

	RefreshShopCommodityPoolVersionKeyFmt = "%s-Zone%d"
	RefreshShopCommodityPoolKeyFmt        = "%s-Zone%d-Versions%s"
	// 每天签到
	MonthlySigninKeyFmt = "MonthlySignin-Zone%d"
	// 月卡配置
	VIPPrivilegeKeyFmt = "VIPPrivilege-Zone%d"
	// 新玩家7天
	NewPlayerLoginBenefitKeyFmt = "NewPlayerLoginBenefit-Zone%d"
	// 首冲
	FirstRechargeKeyFmt = "FirstRecharge-Zone%d"

	C_MAX_MONTHVIP_COLLECTPRIZEDAYS = 30 // 月卡领奖天数
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

func NormalizeRefreshIntervalHours(intervalHours int32) int32 {
	hours := [8]int32{1, 2, 3, 4, 6, 8, 12, 24}

	for _, val := range hours {
		if intervalHours <= val {
			return val
		}
	}
	return 24
}

func CalcRefreshStartHours(intervalHours, dateHours int32) int32 {
	if dateHours > 23 {
		dateHours = 23
	}
	return dateHours / intervalHours * intervalHours
}
