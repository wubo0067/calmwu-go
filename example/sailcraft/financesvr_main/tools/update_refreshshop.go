/*
 * @Author: calmwu
 * @Date: 2018-02-28 14:31:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-14 16:53:06
 * @Comment:
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sailcraft/financesvr_main/proto"
	"time"
)

const (
	RefreshShopConfigFile                = "RandomShopCommodity.json"
	UrlUpdateRefreshShopConfigFmt        = "http://%s/sailcraft/api/v1/FinanceSvr/UpdateRefreshShopConfig"
	UrlUpdateRefreshShopCommodityPoolFmt = "http://%s/sailcraft/api/v1/FinanceSvr/UpdateRefreshShopCommodityPool"
)

type us_resource struct {
	CommodityPrice int32  `json:"count"`
	GamePayType    string `json:"type"`
}

type us_resources struct {
	Resources []us_resource `json:"resources"`
}

type us_prop struct {
	CommodityPropStackCount int32 `json:"count"`
	CommodityPropID         int32 `json:"protype_id"`
}

type us_props struct {
	Props []us_prop `json:"props"`
}

type us_commodity struct {
	CommodityChance int          `json:"Chance"`
	Cost            us_resources `json:"Cost"`
	Discount        int          `json:"Discount"`
	CommodityID     int          `json:"Id"`
	InnerGoods      interface{}  `json:"InnerGoods"`
	ShopType        string       `json:"ShopType"`
	SlotPool        string       `json:"SlotPool"`
}

func getPayType(moneyName string) proto.GamePayType {
	payType := proto.E_GAMEPAY_DIAMOND
	if moneyName == "gold" {
		payType = proto.E_GAMEPAY_GOLD
	} else if moneyName == "honor" {
		payType = proto.E_GAMEPAY_HONOR
	} else if moneyName == "shipsoul" {
		payType = proto.E_GAMEPAY_SHIPSOUL
	}
	return payType
}

func configRefreshShopPools(refreshShopCommodities []us_commodity) {
	refreshShopPools := make(map[string][]proto.RefreshShopCommodityS)

	for index, _ := range refreshShopCommodities {
		commodityConfigInfo := &refreshShopCommodities[index]
		poolName := commodityConfigInfo.SlotPool

		commodityJsonContent, err := json.Marshal(commodityConfigInfo.InnerGoods)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Pool[%s] commodityID[%d] marshal InnerGoods failed! reason[%s]",
				poolName, commodityConfigInfo.CommodityID, err.Error())
			os.Exit(-1)
		}

		poolCommodities, exists := refreshShopPools[poolName]
		if !exists {
			// 添加pool
			fmt.Printf("add refreshShop pool[%s]\n", poolName)
			refreshShopPools[poolName] = make([]proto.RefreshShopCommodityS, 0)

			refreshShopPools[poolName] = append(refreshShopPools[poolName], proto.RefreshShopCommodityS{
				CommodityID:          commodityConfigInfo.CommodityID,
				GamePayType:          getPayType(commodityConfigInfo.Cost.Resources[0].GamePayType),
				CommodityPrice:       commodityConfigInfo.Cost.Resources[0].CommodityPrice,
				CommodityJsonContent: string(commodityJsonContent),
				CommodityChance:      commodityConfigInfo.CommodityChance,
			})
		} else {
			//
			refreshShopPools[poolName] = append(poolCommodities, proto.RefreshShopCommodityS{
				CommodityID:          commodityConfigInfo.CommodityID,
				GamePayType:          getPayType(commodityConfigInfo.Cost.Resources[0].GamePayType),
				CommodityPrice:       commodityConfigInfo.Cost.Resources[0].CommodityPrice,
				CommodityJsonContent: string(commodityJsonContent),
				CommodityChance:      commodityConfigInfo.CommodityChance,
			})
		}
	}

	for poolName, commodities := range refreshShopPools {
		var configPoolReq proto.ProtoUpdateRefreshShopCommodityPoolReq
		configPoolReq.Uin = *cmdParamsUin
		configPoolReq.ZoneID = int32(*cmdParamsZoneID)
		configPoolReq.RefreshShopCommodityPoolConfig = new(proto.RefreshShopCommodityPoolConfigS)
		configPoolReq.RefreshShopCommodityPoolConfig.PoolName = poolName
		configPoolReq.RefreshShopCommodityPoolConfig.VersionID = *cmdParamsVersion
		configPoolReq.RefreshShopCommodityPoolConfig.PoolCommodities = commodities

		req := base.ProtoRequestS{
			ProtoRequestHeadS: base.ProtoRequestHeadS{
				Version:    1,
				EventId:    998,
				TimeStamp:  time.Now().Unix(),
				ChannelUID: "21312",
				Uin:        int(*cmdParamsUin),
				CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
			},
			ReqData: base.ProtoData{
				InterfaceName: "UpdateRefreshShopCommodityPool",
				Params:        configPoolReq,
			},
		}

		fmt.Printf("configPoolReq.PoolName[%s] VersionID[%s]\n", poolName, configPoolReq.RefreshShopCommodityPoolConfig.VersionID)
		UrlQuery := fmt.Sprintf(UrlUpdateRefreshShopCommodityPoolFmt, *cmdParamsSvrIp)
		SendRequest(UrlQuery, &req)
	}
}

func configRefreshShop(configPath string, scMgr *ShopConfigMgr) {
	fileFullName := configPath + "/" + RefreshShopConfigFile
	conf_file, err := os.Open(fileFullName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open [%s] failed! err[%s]\n", fileFullName, err.Error())
		return
	}
	defer conf_file.Close()

	data, err := ioutil.ReadAll(conf_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s failed, reason:%s:\n", fileFullName, err.Error())
		return
	}

	refreshShopCommodities := make([]us_commodity, 0)
	err = json.Unmarshal(data, &refreshShopCommodities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//fmt.Printf("refreshShopCommodities:%+v\n", refreshShopCommodities)

	commonShopConfig := scMgr.Find("commonshop")
	breakoutShopConfig := scMgr.Find("breakoutshop")

	var configReq proto.ProtoUpdateRefreshShopConfigReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	configReq.RefreshShopConfig = new(proto.RefreshShopConfigS)

	configReq.RefreshShopConfig.ShopAutoRefreshIntervalHours = commonShopConfig.DurationHours
	configReq.RefreshShopConfig.CommonManualRefreshCosts = commonShopConfig.Cost
	configReq.RefreshShopConfig.CommonManualRefreshPayType = getPayType(commonShopConfig.RefreshPayType)

	configReq.RefreshShopConfig.CommonShopCommodityPools =
		make([]proto.CommodityPoolS, len(commonShopConfig.Pools))
	for index, _ := range commonShopConfig.Pools {
		configReq.RefreshShopConfig.CommonShopCommodityPools[index] = proto.CommodityPoolS{
			PoolName:          commonShopConfig.Pools[index].Name,
			DisplaySlotIndexs: commonShopConfig.Pools[index].Slots,
		}
	}
	configReq.RefreshShopConfig.DailyCommonManualRefreshCount = commonShopConfig.RefreshCount
	configReq.RefreshShopConfig.CommonShopPresentCommodityCount = commonShopConfig.SlotCount
	configReq.RefreshShopConfig.CommonShopWeeklyCardDiscountRate = commonShopConfig.WeeklyCardDiscountRate
	configReq.RefreshShopConfig.CommonShopMonthlyCardDiscountRate = commonShopConfig.MonthlyCardDiscountRate
	configReq.RefreshShopConfig.CommonShopCommodityDailyBuyCount = commonShopConfig.CommodityDailyBuyCount
	//-----------------------------------------------------------------------------------------------

	configReq.RefreshShopConfig.BreakoutManualRefreshCosts = breakoutShopConfig.Cost
	configReq.RefreshShopConfig.BreakoutManualRefreshPayType = getPayType(breakoutShopConfig.RefreshPayType)
	configReq.RefreshShopConfig.BreakoutShopCommodityPools =
		make([]proto.CommodityPoolS, len(breakoutShopConfig.Pools))
	for index, _ := range breakoutShopConfig.Pools {
		configReq.RefreshShopConfig.BreakoutShopCommodityPools[index] = proto.CommodityPoolS{
			PoolName:          breakoutShopConfig.Pools[index].Name,
			DisplaySlotIndexs: breakoutShopConfig.Pools[index].Slots,
		}
	}
	configReq.RefreshShopConfig.DailyBreakoutManualRefreshCount = breakoutShopConfig.RefreshCount
	configReq.RefreshShopConfig.BreakoutShopPresentCommodityCount = breakoutShopConfig.SlotCount
	configReq.RefreshShopConfig.BreakoutWeeklyCardDiscountRate = breakoutShopConfig.WeeklyCardDiscountRate
	configReq.RefreshShopConfig.BreakoutMonthlyCardDiscountRate = breakoutShopConfig.MonthlyCardDiscountRate
	configReq.RefreshShopConfig.BreakoutCommodityDailyBuyCount = breakoutShopConfig.CommodityDailyBuyCount

	req := base.ProtoRequestS{
		ProtoRequestHeadS: base.ProtoRequestHeadS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "21312",
			Uin:        int(*cmdParamsUin),
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: base.ProtoData{
			InterfaceName: "UpdateRefreshShopConfig",
			Params:        configReq,
		},
	}

	fmt.Printf("RefreshShopConfig:%+v\n\n", configReq.RefreshShopConfig)

	UrlQuery := fmt.Sprintf(UrlUpdateRefreshShopConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)

	//---------------------------------------------------------------------------------------------
	configRefreshShopPools(refreshShopCommodities)
}
