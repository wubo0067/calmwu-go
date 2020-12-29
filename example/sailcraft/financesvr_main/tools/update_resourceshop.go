/*
 * @Author: calmwu
 * @Date: 2018-02-28 14:30:51
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-15 11:41:45
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
	ResourceShopConfigFile          = "ShopResources.json"
	UrlRefreshResourceShopConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/RefreshResourceShopConfig"
)

type rs_costresource struct {
	ResourceCommodityDiamondCost int32 `json:"count"`
}

type rs_costresources struct {
	Resources []rs_costresource `json:"resources"`
}

type rs_commodityresource struct {
	ResourceCommodityStackCount int32  `json:"count"`
	ResourceCommodityType       string `json:"type"`
}

type rs_commodityresources struct {
	Resources []rs_commodityresource `json:"resources"`
}

type rs_commodity struct {
	CostResources         rs_costresources      `json:"Cost"`
	DescKey               string                `json:"DescKey"`
	Icon                  string                `json:"Icon"`
	IconAtlas             string                `json:"IconAtlas"`
	ResourceCommodityID   int                   `json:"Id"`
	CommodityResources    rs_commodityresources `json:"InnerGoods"`
	ResourceCommodityName string                `json:"NameKey"`
}

func configResourceShop(configPath string, scMgr *ShopConfigMgr) {
	// 读取配置文件
	fileFullName := configPath + "/" + ResourceShopConfigFile
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

	resourceShopCommodities := make([]rs_commodity, 0)
	err = json.Unmarshal(data, &resourceShopCommodities)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	// 查询对应的配置
	shopConfig := scMgr.Find("resourcesshop")

	// 将其转换为系统对应的数据结构
	var configReq proto.ProtoRefreshResourceShopConfigReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	configReq.ResourceShopConfigInfo.Count = len(resourceShopCommodities)
	configReq.ResourceShopConfigInfo.VersionID = *cmdParamsVersion
	configReq.ResourceShopConfigInfo.WeeklyCardDiscountRate = shopConfig.WeeklyCardDiscountRate
	configReq.ResourceShopConfigInfo.MonthlyCardDiscountRate = shopConfig.MonthlyCardDiscountRate

	configReq.ResourceShopConfigInfo.ResourceCommodities = make([]proto.ResourceCommodityInfoS, configReq.ResourceShopConfigInfo.Count)
	for index, _ := range resourceShopCommodities {
		configReq.ResourceShopConfigInfo.ResourceCommodities[index].ResourceCommodityID =
			resourceShopCommodities[index].ResourceCommodityID

		if resourceShopCommodities[index].CommodityResources.Resources[0].ResourceCommodityType == "gold" {
			configReq.ResourceShopConfigInfo.ResourceCommodities[index].ResourceCommodityType = proto.E_RESOURCECOMMODITY_GOLD
		} else if resourceShopCommodities[index].CommodityResources.Resources[0].ResourceCommodityType == "wood" {
			configReq.ResourceShopConfigInfo.ResourceCommodities[index].ResourceCommodityType = proto.E_RESOURCECOMMODITY_WOOD
		} else if resourceShopCommodities[index].CommodityResources.Resources[0].ResourceCommodityType == "stone" {
			configReq.ResourceShopConfigInfo.ResourceCommodities[index].ResourceCommodityType = proto.E_RESOURCECOMMODITY_STONE
		} else if resourceShopCommodities[index].CommodityResources.Resources[0].ResourceCommodityType == "iron" {
			configReq.ResourceShopConfigInfo.ResourceCommodities[index].ResourceCommodityType = proto.E_RESOURCECOMMODITY_IRON
		}

		configReq.ResourceShopConfigInfo.ResourceCommodities[index].ResourceCommodityStackCount =
			resourceShopCommodities[index].CommodityResources.Resources[0].ResourceCommodityStackCount
		configReq.ResourceShopConfigInfo.ResourceCommodities[index].ResourceCommodityDiamondCost =
			resourceShopCommodities[index].CostResources.Resources[0].ResourceCommodityDiamondCost
	}

	fmt.Printf("ProtoRefreshResourceShopConfigReq:%+v\n", configReq)

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
			InterfaceName: "RefreshResourceShopConfig",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlRefreshResourceShopConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
