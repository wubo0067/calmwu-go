/*
 * @Author: calmwu
 * @Date: 2018-02-28 14:30:23
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 17:34:01
 * @Comment:
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sailcraft/base"
	"strings"
)

const (
	C_UPDATETYPE_RESOURCESHOPTYPE = "resource"
	C_UPDATETYPE_CARDPACKSHOPTYPE = "cardpack"
	C_UPDATETYPE_REFRESHSHOPTYPE  = "refresh"
	C_UPDATETYPE_RECHARGESHOPTYPE = "recharge"
	C_UPDATETYPE_MONTHSIGNIN      = "signin"
	C_UDPATETYPE_VIPPRIVILEGE     = "vip"
	C_UPDATETYPE_NEWPLAYERBENEFIT = "newplayerbenefit"
	C_UPDATETYPE_ACTIVESUPERGIFT  = "supergift"
	C_UPDATETYPE_ACTIVEMISSION    = "mission"
	C_UPDATETYPE_ACTIVEEXCHANGE   = "exchange"
	C_UPDATETYPE_ACTIVECDKEY      = "cdkey"
	C_UPDATETYPE_FIRSTRECHARGE    = "firstrecharge"

	// 商店配置
	ShopConfigFile = "RandomShopRefreshCost.json"
)

var (
	cmdParamsUpdateType = flag.String("type", "", "配置的商店类型")
	cmdParamsConfigPath = flag.String("configpath", "", "json配置文件路径")
	cmdParamsVersion    = flag.String("version", "", "版本号")
	cmdParamsSvrIp      = flag.String("svrip", "123.59.40.19:400", "")
	cmdParamsUin        = flag.Uint64("uin", 1, "")
	cmdParamsZoneID     = flag.Int("zoneid", 0, "")
)

type shop_poolcfg struct {
	Name  string  `json:"Name"`
	Slots []int32 `json:Slots`
}

// type shop_pools struct {
// 	Pools []shop_poolcfg `json:"Pools"`
// }

type shop_config struct {
	CommodityDailyBuyCount  int32          `json:"CommodityDailyBuyCount"`
	Cost                    []int32        `json:"Cost"`                    // 刷新费用
	DurationHours           int32          `json:"Duration"`                // 刷新的时间间隔，单位小时
	MonthlyCardDiscountRate int32          `json:"MonthlyCardDiscountRate"` // 月卡折扣率
	WeeklyCardDiscountRate  int32          `json:"WeeklyCardDiscountRate"`  // 普通月卡折扣率
	RefreshCount            int32          `json:"RefreshTimes"`            // 刷新次数
	RefreshPayType          string         `json:"Resource"`                // 刷新花费代币类型
	ShopType                string         `json:"ShopType"`                // 商店类型名
	Pools                   []shop_poolcfg `json:"Pools"`                   // slot对应的pool
	SlotCount               int32          `json:"SlotCount"`               // 商店显示槽位
}

type ShopConfigMgr struct {
	shopConfigs []shop_config
}

func (sc ShopConfigMgr) Find(shopType string) *shop_config {
	for index, _ := range sc.shopConfigs {
		if shopType == sc.shopConfigs[index].ShopType {
			return &sc.shopConfigs[index]
		}
	}
	fmt.Fprintf(os.Stderr, "shopType[%s] is invalid!\n", shopType)
	return nil
}

func (sc *ShopConfigMgr) Load() {
	fileFullName := *cmdParamsConfigPath + "/" + ShopConfigFile
	conf_file, err := os.Open(fileFullName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open [%s] failed! err[%s]\n", fileFullName, err.Error())
		os.Exit(-1)
	}
	defer conf_file.Close()

	data, err := ioutil.ReadAll(conf_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s failed, reason:%s:\n", fileFullName, err.Error())
		os.Exit(-1)
	}

	err = json.Unmarshal(data, &sc.shopConfigs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		os.Exit(-1)
	}

	//fmt.Printf("ShopConfigMgr:%+v\n\n", sc.shopConfigs)
	return
}

func showUsage() {
	fmt.Printf(`shopConfig usage:
	./shopConfig --type=[resource|cardpack|refresh|recharge|signin|vip|newplayerbenefit|supergift|mission|exchange|cdkey|firstrecharge] --configpath=../doc/shop --svrip=123.59.40.19 --version=1 --zoneid=1 --uin=1`)
	return
}

func SendRequest(url string, req *base.ProtoRequestS) {
	// 编码压缩
	serialData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("url[%s] serialData len[%d]\n", url, len(serialData))

	// 发送
	res, err := http.Post(url, "text/plain; charset=utf-8", strings.NewReader(string(serialData)))
	if err != nil {
		fmt.Printf("Post to %s failed! [%s]\n", url, err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Read body failed! reason[%s]\n", err.Error())
		return
	}
	fmt.Printf("%s\n", body)
}

func main() {
	flag.Parse()

	scMgr := new(ShopConfigMgr)
	scMgr.shopConfigs = make([]shop_config, 0)
	scMgr.Load()

	if *cmdParamsZoneID == 0 || len(*cmdParamsVersion) == 0 {
		fmt.Fprintf(os.Stderr, "ZoneID or VersionID must be set!")
		showUsage()
	}

	if *cmdParamsUpdateType == C_UPDATETYPE_RESOURCESHOPTYPE {
		configResourceShop(*cmdParamsConfigPath, scMgr)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_CARDPACKSHOPTYPE {
		configCardPackShop(*cmdParamsConfigPath, scMgr)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_REFRESHSHOPTYPE {
		configRefreshShop(*cmdParamsConfigPath, scMgr)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_RECHARGESHOPTYPE {
		configRechargeShop(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_MONTHSIGNIN {
		configMonthSignInActive(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UDPATETYPE_VIPPRIVILEGE {
		configVIPPrivilege(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_NEWPLAYERBENEFIT {
		configNewPlayerLoginBenefit(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_ACTIVESUPERGIFT {
		configActiveSuperGift(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_ACTIVEMISSION {
		configActiveMission(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_ACTIVEEXCHANGE {
		configActiveExchange(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_ACTIVECDKEY {
		configActiveCDKeyExchange(*cmdParamsConfigPath)
	} else if *cmdParamsUpdateType == C_UPDATETYPE_FIRSTRECHARGE {
		configActiveFirstRecharge(*cmdParamsConfigPath)
	} else {
		showUsage()
	}

	return
}
