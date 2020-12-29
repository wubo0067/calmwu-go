/*
 * @Author: calmwu
 * @Date: 2018-02-11 11:54:52
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-27 19:16:18
 */

package handler

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/financesvr_main/common"
	"sailcraft/financesvr_main/proto"
)

func QueryShopCommoditiesInfo(zoneID int32, shopCommodityType proto.ShopCommodityType) (interface{}, error) {
	var redisVersionKey string
	var redisCommodityKey string
	var shopCommoditiesInfo interface{}

	switch shopCommodityType {
	case proto.E_SHOPCOMMODITY_RECHARGE:
		redisVersionKey = fmt.Sprintf(common.ShopRechargeCommodityVersionKeyFmt, zoneID)
		shopCommoditiesInfo = new(proto.ShopRechargeCommoditiesInfoS)
	case proto.E_SHOPCOMMODITY_RESOURCE:
		redisVersionKey = fmt.Sprintf(common.ShopResourceCommodityVersionKeyFmt, zoneID)
		shopCommoditiesInfo = new(proto.ResourceShopConfigS)
	case proto.E_SHOPCOMMODITY_CARDPACK:
		redisVersionKey = fmt.Sprintf(common.ShopCardPackCommodityVersionKeyFmt, zoneID)
		shopCommoditiesInfo = new(proto.ShopCardPackCommoditiesInfoS)
	default:
		err := fmt.Errorf("shopCommodityType[%s] is invalid!", shopCommodityType.String())
		base.GLog.Error(err.Error())
		return nil, err
	}

	val, err := common.GRedis.StringGet(redisVersionKey)
	if err != nil {
		err := fmt.Errorf("shopCommodityType[%s] versionKey[%s] is invalid! reason[%s]", shopCommodityType.String(),
			redisVersionKey, err.Error())
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 生成key从redis中查询
	redisCommodityKey = string(val.([]byte))
	val, err = common.GRedis.StringGet(redisCommodityKey)
	// 返回查询数据
	if err != nil {
		err := fmt.Errorf("shopCommodityType[%s] versionKey[%s] commodityKey[%s] is invalid! reason[%s]", shopCommodityType.String(),
			redisVersionKey, redisCommodityKey, err.Error())
		base.GLog.Error(err.Error())
		return nil, err
	}

	shopCommiditiesData := val.([]byte)
	err = json.Unmarshal(shopCommiditiesData, shopCommoditiesInfo)
	if err != nil {
		err = fmt.Errorf("Unmarshal redisCommodityKey[%s] data failed! reason[%s]", redisCommodityKey, err.Error())
		base.GLog.Error(err.Error())
		return nil, err
	}

	return shopCommoditiesInfo, nil
}

// 更新
func UpdateShopCommoditiesInfo(zoneID int32, versionID string, commoditiesRedisData []byte, shopCommodityType proto.ShopCommodityType) error {
	var shopCommodityKey string
	var shopCommodityVersionKey string

	switch shopCommodityType {
	case proto.E_SHOPCOMMODITY_RECHARGE:
		shopCommodityKey = fmt.Sprintf(common.ShopRechargeCommoditiesKeyFmt, zoneID, versionID)
		shopCommodityVersionKey = fmt.Sprintf(common.ShopRechargeCommodityVersionKeyFmt, zoneID)
	case proto.E_SHOPCOMMODITY_RESOURCE:
		shopCommodityKey = fmt.Sprintf(common.ShopResourceCommoditiesKeyFmt, zoneID, versionID)
		shopCommodityVersionKey = fmt.Sprintf(common.ShopResourceCommodityVersionKeyFmt, zoneID)
	case proto.E_SHOPCOMMODITY_CARDPACK:
		shopCommodityKey = fmt.Sprintf(common.ShopCardPackCommoditiesKeyFmt, zoneID, versionID)
		shopCommodityVersionKey = fmt.Sprintf(common.ShopCardPackCommodityVersionKeyFmt, zoneID)
	default:
		err := fmt.Errorf("shopCommodityType[%d] is invalid!", shopCommodityType)
		base.GLog.Error(err.Error())
		return err
	}

	base.GLog.Debug("shopCommodityType[%s] shopCommidityKey[%s]", shopCommodityType.String(), shopCommodityKey)

	bExists, err := common.GRedis.Exists(shopCommodityKey)
	if err != nil {
		err := fmt.Errorf("Query shopCommodityKey[%s] exists failed! reason[%s]", shopCommodityKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}

	if bExists {
		// 存在，返回失败
		err := fmt.Errorf("shopCommidityKey:%s already exists! the version number must be incremented!", shopCommodityKey)
		base.GLog.Error(err.Error())
		return err
	}

	err = common.GRedis.StringSet(shopCommodityKey, commoditiesRedisData)
	if err != nil {
		err := fmt.Errorf("Set shopCommodityKey[%s] data failed! reason[%s]", shopCommodityKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}

	// 同时更新当前商品版本号
	common.GRedis.StringSet(shopCommodityVersionKey, []byte(shopCommodityKey))

	return nil
}

func UpdateRefreshShopConfig(req *proto.ProtoUpdateRefreshShopConfigReq) error {
	req.RefreshShopConfig.ShopAutoRefreshIntervalHours =
		common.NormalizeRefreshIntervalHours(req.RefreshShopConfig.ShopAutoRefreshIntervalHours)

	redisData, err := json.Marshal(req.RefreshShopConfig)
	if err != nil {
		base.GLog.Error("Uin[%d] ZoneID[%d] Marshal failed! reason[%s]", req.Uin, req.ZoneID, err.Error())
		return err
	}

	// 检查 商品刷新次数和对应的价格数组长度是否一致
	if int32(len(req.RefreshShopConfig.CommonManualRefreshCosts)) != req.RefreshShopConfig.DailyCommonManualRefreshCount {
		err = fmt.Errorf("CommonManualRefreshCosts len[%d] not equal DailyCommonManualRefreshCount[%d]",
			len(req.RefreshShopConfig.CommonManualRefreshCosts), req.RefreshShopConfig.DailyCommonManualRefreshCount)
		base.GLog.Error(err.Error())
		return err
	}

	if int32(len(req.RefreshShopConfig.BreakoutManualRefreshCosts)) != req.RefreshShopConfig.DailyBreakoutManualRefreshCount {
		err = fmt.Errorf("BreakoutManualRefreshCosts len[%d] not equal DailyBreakoutManualRefreshCount[%d]",
			len(req.RefreshShopConfig.BreakoutManualRefreshCosts), req.RefreshShopConfig.DailyBreakoutManualRefreshCount)
		base.GLog.Error(err.Error())
		return err
	}

	// 检查橱窗商品展示数量是否和pool的绑定数量一致
	commodityPresentCount := req.RefreshShopConfig.CommonShopPresentCommodityCount
	slotCount := 0
	for index, _ := range req.RefreshShopConfig.CommonShopCommodityPools {
		slotCount += len(req.RefreshShopConfig.CommonShopCommodityPools[index].DisplaySlotIndexs)
	}
	if int(commodityPresentCount) != slotCount {
		err = fmt.Errorf("CommonShopConfig is invalid! total displaySlotCount[%d] not equal CommonShopPresentCommodityCount[%d]",
			slotCount, commodityPresentCount)
		base.GLog.Error(err.Error())
		return err
	}

	commodityPresentCount = req.RefreshShopConfig.BreakoutShopPresentCommodityCount
	slotCount = 0
	for index, _ := range req.RefreshShopConfig.BreakoutShopCommodityPools {
		slotCount += len(req.RefreshShopConfig.BreakoutShopCommodityPools[index].DisplaySlotIndexs)
	}
	if int(commodityPresentCount) != slotCount {
		err = fmt.Errorf("BreakoutShopConfig is invalid! total displaySlotCount[%d] not equal BreakoutShopPresentCommodityCount[%d]",
			slotCount, commodityPresentCount)
		base.GLog.Error(err.Error())
		return err
	}

	refreshShopConfigKey := fmt.Sprintf(common.RefreshShopConfigKeyFmt, req.ZoneID)

	base.GLog.Debug("refreshShopConfigKey[%s] ShopAutoRefreshIntervalHours[%d]",
		refreshShopConfigKey, req.RefreshShopConfig.ShopAutoRefreshIntervalHours)

	err = common.GRedis.StringSet(refreshShopConfigKey, redisData)
	if err != nil {
		err := fmt.Errorf("Set refreshShopConfigKey[%s] data failed! reason[%s]", refreshShopConfigKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}
	return nil
}

func getRefreshShopConfig(zoneID int32) *proto.RefreshShopConfigS {
	refreshShopConfigKey := fmt.Sprintf(common.RefreshShopConfigKeyFmt, zoneID)

	base.GLog.Debug("refreshShopConfigKey[%s]", refreshShopConfigKey)

	var configInfo proto.RefreshShopConfigS

	err := common.GetStrDataFromRedis(refreshShopConfigKey, &configInfo)
	if err != nil {
		return nil
	}

	return &configInfo
}

func UpdateRefreshShopCommodityPool(req *proto.ProtoUpdateRefreshShopCommodityPoolReq) error {
	refreshShopCommodityPoolKey := fmt.Sprintf(common.RefreshShopCommodityPoolKeyFmt, req.RefreshShopCommodityPoolConfig.PoolName,
		req.ZoneID, req.RefreshShopCommodityPoolConfig.VersionID)
	refreshShopCommodityPoolVersionKey := fmt.Sprintf(common.RefreshShopCommodityPoolVersionKeyFmt, req.RefreshShopCommodityPoolConfig.PoolName,
		req.ZoneID)

	base.GLog.Debug("refreshShopCommodityPoolKey[%s] refreshShopCommodityPoolVersionKey[%s]",
		refreshShopCommodityPoolKey, refreshShopCommodityPoolVersionKey)

	bExists, err := common.GRedis.Exists(refreshShopCommodityPoolKey)
	if err != nil {
		err := fmt.Errorf("Query refreshShopCommodityPoolKey[%s] exists failed! reason[%s]", refreshShopCommodityPoolKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}

	if bExists {
		// 存在，返回失败
		err := fmt.Errorf("refreshShopCommodityPoolKey:%s already exists! the version number must be incremented!", refreshShopCommodityPoolKey)
		base.GLog.Error(err.Error())
		return err
	}

	// 设置商品的PoolKey
	for index, _ := range req.RefreshShopCommodityPoolConfig.PoolCommodities {
		req.RefreshShopCommodityPoolConfig.PoolCommodities[index].PoolKey = refreshShopCommodityPoolKey
	}

	// json marshal
	redisData, err := json.Marshal(req.RefreshShopCommodityPoolConfig.PoolCommodities)
	if err != nil {
		base.GLog.Error("refreshShopCommodityPoolKey[%s] data Marshal failed! reason[%s]", refreshShopCommodityPoolKey, err.Error())
		return err
	}

	// redis set
	err = common.GRedis.StringSet(refreshShopCommodityPoolKey, redisData)
	if err != nil {
		err := fmt.Errorf("Set refreshShopCommodityPoolKey[%s] data failed! reason[%s]", refreshShopCommodityPoolKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}

	// 设置关联
	common.GRedis.StringSet(refreshShopCommodityPoolVersionKey, []byte(refreshShopCommodityPoolKey))
	return nil
}

func getRefreshShopCommodityPool(zoneID int32, poolName string) ([]proto.RefreshShopCommodityS, error) {
	refreshShopCommodityPoolVersionKey := fmt.Sprintf(common.RefreshShopCommodityPoolVersionKeyFmt,
		poolName, zoneID)

	base.GLog.Debug("refreshShopCommodityPoolVersionKey[%s]", refreshShopCommodityPoolVersionKey)

	val, err := common.GRedis.StringGet(refreshShopCommodityPoolVersionKey)
	if err != nil {
		err := fmt.Errorf("refreshShopCommodityPoolVersionKey[%s] is invalid! reason[%s]",
			refreshShopCommodityPoolVersionKey, err.Error())
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 生成key从redis中查询
	refreshShopCommodityPoolKey := string(val.([]byte))
	return getRefreshShopCommodityPoolByKey(refreshShopCommodityPoolKey)
}

func getRefreshShopCommodityPoolByKey(poolKey string) ([]proto.RefreshShopCommodityS, error) {
	// val, err := common.GRedis.StringGet(poolKey)
	// // 返回查询数据
	// if err != nil {
	// 	err := fmt.Errorf("refreshShopCommodityPoolKey[%s] is invalid! reason[%s]", poolKey, err.Error())
	// 	base.GLog.Error(err.Error())
	// 	return nil, err
	// }

	// poolData := val.([]byte)
	poolCommodities := make([]proto.RefreshShopCommodityS, 0)

	err := common.GetStrDataFromRedis(poolKey, &poolCommodities)
	if err != nil {
		return nil, err
	}

	// err = json.Unmarshal(poolData, &poolCommodities)
	// if err != nil {
	// 	err := fmt.Errorf("Unmarshal refreshShopCommodityPoolKey[%s] data failed! reason[%s]", poolKey, err.Error())
	// 	base.GLog.Error(err.Error())
	// 	return nil, err
	// }

	return poolCommodities, nil
}
