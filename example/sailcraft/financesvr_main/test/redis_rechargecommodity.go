/*
 * @Author: calmwu
 * @Date: 2018-02-02 11:31:52
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-06 16:47:42
 */

package main

import (
	"encoding/json"
	"fmt"
	"sailcraft/dataaccess/redistool"
	"sailcraft/financesvr_main/proto"
)

func main() {
	shopRechargeCommodityInfo := new(proto.ShopRechargeCommoditiesInfoS)
	shopRechargeCommodityInfo.RechargeCommodities = make([]*proto.RechargeCommodityInfoS, 6)
	shopRechargeCommodityInfo.Count = 6
	shopRechargeCommodityInfo.VersionID = "1.0.0"

	i := 0
	for i < 6 {
		shopRechargeCommodityInfo.RechargeCommodities[i] = new(proto.RechargeCommodityInfoS)
		shopRechargeCommodityInfo.RechargeCommodities[i].RechargeCommodityID = i
		shopRechargeCommodityInfo.RechargeCommodities[i].ExchangeDiamonds = 100 * i
		shopRechargeCommodityInfo.RechargeCommodities[i].FirstRechargePresentDiamonds = 200 * i
		shopRechargeCommodityInfo.RechargeCommodities[i].PresentDiamonds = 200 * i
		i++
	}

	redisData, err := json.Marshal(shopRechargeCommodityInfo)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(shopRechargeCommodityInfo.String())
	fmt.Println("--------------------")

	redisAddr := "127.0.0.1:6379"
	sessionCount := 5
	redisMgr := redistool.NewRedis(redisAddr, sessionCount)
	err = redisMgr.Start()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 将充值商品写入redis
	err = redisMgr.StringSet("ShopRecharge-Zone1-1.0.0", redisData)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		fmt.Printf("set key[ShopRecharge-Zone1-1.0.0] successed!\n")
	}

	// 读取充值商品
	val, err := redisMgr.StringGet("ShopRecharge-Zone1-1.0.0")
	if err != nil {
		fmt.Println(err.Error())
	}

	redisData = val.([]byte)
	shopRechargeCommodityInfo_out := new(proto.ShopRechargeCommoditiesInfoS)
	err = json.Unmarshal(redisData, shopRechargeCommodityInfo_out)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("%+v\n", shopRechargeCommodityInfo_out)
}
