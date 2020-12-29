/*
 * @Author: calmwu
 * @Date: 2018-03-13 14:14:56
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-13 18:49:10
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
	RechargeShopConfigFile           = "IAP.json"
	UrlRefreshRechargeCommoditiesFmt = "http://%s/sailcraft/api/v1/FinanceSvr/RefreshRechargeCommodities"
)

type rc_commodity struct {
	BuyDiamonds                       int32   `json:"Count"`
	DescKey                           string  `json:"DescKey"`
	FirstRechargePresentDiamonds      int32   `json:"FirstTimePurchaseReward"`
	FirstTimePurchaseRewardCaptionKey string  `json:"FirstTimePurchaseRewardCaptionKey"`
	Icon                              string  `json:"Icon"`
	IconAtlas                         string  `json:"IconAtlas"`
	RechargeCommodityID               int     `json:"Id"`
	NameKey                           string  `json:"NameKey"`
	Price                             float32 `json:"Price"`
	PriceDesc                         string  `json:"PriceDesc"`
	ProductId                         string  `json:"ProductId"`
	PresentDiamonds                   int32   `json:"PurchaseReward"`
}

type rc_shop struct {
	AreaCode string         `json:"AreaCode"`
	IAP      []rc_commodity `json:"IAP"`
}

func configRechargeShop(configPath string) {
	fileFullName := configPath + "/" + RechargeShopConfigFile
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

	rechargeShops := make([]rc_shop, 0)
	err = json.Unmarshal(data, &rechargeShops)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//fmt.Printf("rechargeShops:%+v\n", rechargeShops)
	var configReq proto.ProtoRefreshRechargeCommoditiesReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	configReq.ShopRechargeCommoditiesInfo.VersionID = *cmdParamsVersion
	configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos = make([]proto.ChannelRechargeCommodityInfoS, len(rechargeShops))
	for index, _ := range rechargeShops {
		configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].ChannelID = rechargeShops[index].AreaCode
		configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities =
			make([]proto.RechargeCommodityInfoS, len(rechargeShops[index].IAP))

		for j, _ := range rechargeShops[index].IAP {
			// configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].DescKey =
			// 	rechargeShops[index].IAP[j].DescKey
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].RechargeCommodityID =
				rechargeShops[index].IAP[j].RechargeCommodityID
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].BuyDiamonds =
				rechargeShops[index].IAP[j].BuyDiamonds
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].FirstRechargePresentDiamonds =
				rechargeShops[index].IAP[j].FirstRechargePresentDiamonds
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].PresentDiamonds =
				rechargeShops[index].IAP[j].PresentDiamonds
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].FirstTimePurchaseRewardCaptionKey =
				rechargeShops[index].IAP[j].FirstTimePurchaseRewardCaptionKey
			// configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].Icon =
			// 	rechargeShops[index].IAP[j].Icon
			// configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].IconAtlas =
			// 	rechargeShops[index].IAP[j].IconAtlas
			// configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].NameKey =
			// 	rechargeShops[index].IAP[j].NameKey
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].Price =
				rechargeShops[index].IAP[j].Price
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].PriceDesc =
				rechargeShops[index].IAP[j].PriceDesc
			configReq.ShopRechargeCommoditiesInfo.ChannelRechargeCommodityInfos[index].RechargeCommodities[j].ProductId =
				rechargeShops[index].IAP[j].ProductId
		}
	}

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
			InterfaceName: "RefreshRechargeCommodities",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlRefreshRechargeCommoditiesFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
