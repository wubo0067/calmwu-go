/*
 * @Author: calmwu
 * @Date: 2018-02-28 14:31:54
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-15 11:40:04
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
	CardPackShopConfigFile          = "ShopCardpacks.json"
	UrlRefreshCardPackShopConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/RefreshCardPackShopConfig"
)

type cc_resource struct {
	CardPackDiamondCost int32 `json:"count"`
}

type cc_resources struct {
	Resources []cc_resource `json:"resources"`
}

type cc_prop struct {
	Count     int `json:"count"`
	ProtypeID int `json:"protype_id"`
}

type cc_props struct {
	Props []cc_prop `json:"props"`
}

type rs_cardpack struct {
	CostResources         cc_resources `json:"Cost"`
	DescKey               string       `json:"DescKey"`
	GiftDescKey           string       `json:"GiftDescKey"`
	Highlight             string       `json:"Highlight"`
	Icon                  string       `json:"Icon"`
	IconAtlas             string       `json:"IconAtlas"`
	ResourceCommodityID   int          `json:"Id"`
	InnerGoods            cc_props     `json:"InnerGoods"`
	ResourceCommodityName string       `json:"NameKey"`
	PacksCountKey         string       `json:"PacksCountKey"`
}

func configCardPackShop(configPath string, scMgr *ShopConfigMgr) {
	// 读取配置文件
	fileFullName := configPath + "/" + CardPackShopConfigFile
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

	cardpackShopCommodities := make([]rs_cardpack, 0)
	err = json.Unmarshal(data, &cardpackShopCommodities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}
	// 查询对应的配置
	shopConfig := scMgr.Find("resourcesshop")

	var configReq proto.ProtoRefreshCardPackShopReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	configReq.ShopCardPackCommoditiesInfo = new(proto.ShopCardPackCommoditiesInfoS)
	configReq.ShopCardPackCommoditiesInfo.VersionID = *cmdParamsVersion
	configReq.ShopCardPackCommoditiesInfo.Count = len(cardpackShopCommodities)
	configReq.ShopCardPackCommoditiesInfo.WeeklyCardDiscountRate = shopConfig.WeeklyCardDiscountRate
	configReq.ShopCardPackCommoditiesInfo.MonthlyCardDiscountRate = shopConfig.MonthlyCardDiscountRate

	configReq.ShopCardPackCommoditiesInfo.CardPackCommodities = make([]proto.CardPackCommdityInfoS, configReq.ShopCardPackCommoditiesInfo.Count)

	for index, _ := range cardpackShopCommodities {
		configReq.ShopCardPackCommoditiesInfo.CardPackCommodities[index].CardPackCommodityID =
			cardpackShopCommodities[index].ResourceCommodityID
		configReq.ShopCardPackCommoditiesInfo.CardPackCommodities[index].CardPackDiamondCost =
			cardpackShopCommodities[index].CostResources.Resources[0].CardPackDiamondCost
		configReq.ShopCardPackCommoditiesInfo.CardPackCommodities[index].CardPackRecommend = 1
		if len(cardpackShopCommodities[index].Highlight) == 0 {
			configReq.ShopCardPackCommoditiesInfo.CardPackCommodities[index].CardPackRecommend = 0
		}
		configReq.ShopCardPackCommoditiesInfo.CardPackCommodities[index].CardPackGiftDescKey =
			cardpackShopCommodities[index].GiftDescKey
		configReq.ShopCardPackCommoditiesInfo.CardPackCommodities[index].CardPackCountKey =
			cardpackShopCommodities[index].PacksCountKey

		data, err := json.Marshal(cardpackShopCommodities[index].InnerGoods)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal cardPack InnerGoods failed! reason[%s]", err.Error())
			return
		}
		configReq.ShopCardPackCommoditiesInfo.CardPackCommodities[index].CardPackJsonContent =
			string(data)
	}

	fmt.Printf("ProtoRefreshCardPackShopReq:%+v\n", configReq)

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
			InterfaceName: "RefreshCardPackShopConfig",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlRefreshCardPackShopConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
