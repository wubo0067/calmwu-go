/*
 * @Author: calmwu
 * @Date: 2018-02-05 19:35:58
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 15:02:23
 * @Comment:
 */

package handler

import (
	"fmt"
	"math/rand"
	"sailcraft/base"
	"sailcraft/base/consul_api"
	csssvr_proto "sailcraft/csssvr_main/proto"
	"sailcraft/financesvr_main/common"
	"sailcraft/financesvr_main/proto"
	"time"
)

func DeliveryRechargeCommodity(reqData *proto.ProtoDeliveryRechargeCommodityReq) (*proto.ProtoDeliveryRechargeCommodityRes, error) {
	// 查询玩家
	userFinance, err := QueryFinanceUser(reqData.Uin)
	if userFinance == nil {
		return nil, err
	}

	deliveryRechargeCommodityRes := new(proto.ProtoDeliveryRechargeCommodityRes)
	deliveryRechargeCommodityRes.Uin = reqData.Uin
	deliveryRechargeCommodityRes.ZoneID = reqData.ZoneID

	base.GLog.Debug("Uin[%d] RechargeCommodityType[%s] RechargeCommodityType[%d]",
		reqData.Uin, reqData.RechargeCommodityType.String(), reqData.RechargeCommodityType)

	switch reqData.RechargeCommodityType {
	case proto.E_RECHARGECOMMODITY_DIAMONDS:
		// 从redis根据zone，version获取充值商品列表
		shopCommoditiesInfo, err := QueryShopCommoditiesInfo(reqData.ZoneID, proto.E_SHOPCOMMODITY_RECHARGE)
		if err != nil {
			base.GLog.Error("QueryShopCommoditiesInfo failed, commodiy does not exist in redis!")
			return nil, err
		}

		shopRechargeCommoditiesInfo := shopCommoditiesInfo.(*proto.ShopRechargeCommoditiesInfoS)
		// 在充值商品列表中查询具体的商品
		rechargeCommmodityInfo := shopRechargeCommoditiesInfo.FindRechargeCommodity(reqData.ChannelID, reqData.RechargeCommdityID)
		if rechargeCommmodityInfo == nil {
			return nil, fmt.Errorf("ChannelID[%s] CommodityID[%d] is invalid!", reqData.ChannelID, reqData.RechargeCommdityID)
		}

		// 计算获得钻石数量
		deliveryRechargeCommodityRes.RechargeCommodityType = proto.E_RECHARGECOMMODITY_DIAMONDS
		deliveryRechargeCommodityRes.IsFristPurchase = 0
		deliveryRechargeCommodityRes.RechargeCommdityID = reqData.RechargeCommdityID
		deliveryRechargeCommodityRes.BuyDiamonds = rechargeCommmodityInfo.BuyDiamonds

		// 累计购买的钻石数量
		userFinance.FirstRecharge.CurrBuyDiamonds += rechargeCommmodityInfo.BuyDiamonds
		userFinance.FirstRecharge.SyncDB(common.GDBEngine, reqData.Uin)

		// 判断是否是首次购买该商品
		base.GLog.Debug("Uin[%d] CurrBuyDiamonds[%d]", reqData.Uin, userFinance.FirstRecharge.CurrBuyDiamonds)
		if userFinance.ShopFirstPurchaseInfo.IsFirstPurchase(proto.E_SHOPCOMMODITY_RECHARGE, reqData.RechargeCommdityID) {
			deliveryRechargeCommodityRes.IsFristPurchase = 1
			deliveryRechargeCommodityRes.PresentDiamonds = rechargeCommmodityInfo.FirstRechargePresentDiamonds
			userFinance.ShopFirstPurchaseInfo.FirstPurchase(proto.E_SHOPCOMMODITY_RECHARGE, reqData.RechargeCommdityID, reqData.Uin)
		} else {
			deliveryRechargeCommodityRes.IsFristPurchase = 0
			deliveryRechargeCommodityRes.PresentDiamonds = rechargeCommmodityInfo.PresentDiamonds
		}

		// 转发请求到csssvr进行统计
		var rechargeNtf csssvr_proto.ProtoCssSvrUserRechargeNtf
		rechargeNtf.ChannelID = reqData.ChannelID
		rechargeNtf.PlatForm = reqData.PlatForm
		rechargeNtf.RechargeAmount = rechargeCommmodityInfo.Price
		rechargeNtf.Uin = reqData.Uin
		consul_api.PostRequstByConsulDns(reqData.Uin, "CssSvrUserRecharge", &rechargeNtf, common.ConsulClient, "CassandraSvr")

	case proto.E_RECHARGECOMMODITY_LUXURYMONTHLYCARD:
		// 购买豪华月卡
		vipType := GetFinanceUserVIPType(userFinance)
		if vipType&proto.E_USER_VIP_LUXURYMONTHLY == 0 {
			err = FinanceUserAddVip(userFinance, proto.E_USER_VIP_LUXURYMONTHLY)
			if err == nil {
				base.GLog.Info("Uin[%d] buy LuxuryMonthlyVIP", reqData.Uin)
				deliveryRechargeCommodityRes.RechargeCommodityType = proto.E_RECHARGECOMMODITY_LUXURYMONTHLYCARD
				//
				vipPrivilegeConfig := getVIPPrivilegeConfig(reqData.ZoneID)
				if vipPrivilegeConfig == nil {
					err = fmt.Errorf("Uin[%d] get ZoneID[%d] VIPPrivilegeConfig failed!", reqData.Uin, reqData.ZoneID)
					base.GLog.Error(err.Error())
					return nil, err
				}
				pInfo := vipPrivilegeConfig.FindVIPPrivilege(reqData.ChannelID, reqData.RechargeCommdityID)
				if pInfo == nil {
					err = fmt.Errorf("Uin[%d] ZoneID[%d] LUXURYMONTHLYCARD ChannelID[%s] RechargeCommdityID[%d] is invalid!", reqData.Uin, reqData.ZoneID,
						reqData.ChannelID, reqData.RechargeCommdityID)
					base.GLog.Error(err.Error())
					return nil, err
				}
				// 按运营要求，这里放到赠送钻石
				deliveryRechargeCommodityRes.PresentDiamonds = pInfo.GemCount
				deliveryRechargeCommodityRes.NameKey = pInfo.NameKey

				var rechargeNtf csssvr_proto.ProtoCssSvrUserRechargeNtf
				rechargeNtf.ChannelID = reqData.ChannelID
				rechargeNtf.PlatForm = reqData.PlatForm
				rechargeNtf.RechargeAmount = pInfo.Price
				rechargeNtf.Uin = reqData.Uin
				consul_api.PostRequstByConsulDns(reqData.Uin, "CssSvrUserRecharge", &rechargeNtf, common.ConsulClient, "CassandraSvr")
			}
		} else {
			// 月卡已经存在
			err = fmt.Errorf("Uin[%d] have %s status!", reqData.Uin, proto.E_USER_VIP_LUXURYMONTHLY.String())
			return nil, err
		}
	case proto.E_RECHARGECOMMODITY_NORMALMONTHLYCARD:
		// 购买普通月卡
		vipType := GetFinanceUserVIPType(userFinance)
		if vipType&proto.E_USER_VIP_NORMALMONTHLY == 0 {
			err = FinanceUserAddVip(userFinance, proto.E_USER_VIP_NORMALMONTHLY)
			if err == nil {
				base.GLog.Info("Uin[%d] buy NormalMonthlyVIP", reqData.Uin)
				deliveryRechargeCommodityRes.RechargeCommodityType = proto.E_RECHARGECOMMODITY_NORMALMONTHLYCARD
				//
				vipPrivilegeConfig := getVIPPrivilegeConfig(reqData.ZoneID)
				if vipPrivilegeConfig == nil {
					err = fmt.Errorf("Uin[%d] ZoneID[%d] get VIPPrivilegeConfig failed!", reqData.Uin, reqData.ZoneID)
					base.GLog.Error(err.Error())
					return nil, err
				}
				pInfo := vipPrivilegeConfig.FindVIPPrivilege(reqData.ChannelID, reqData.RechargeCommdityID)
				if pInfo == nil {
					err = fmt.Errorf("Uin[%d] ZoneID[%d] NORMALMONTHLYCARD ChannelID[%s] RechargeCommdityID[%d] is invalid!", reqData.Uin, reqData.ZoneID,
						reqData.ChannelID, reqData.RechargeCommdityID)
					base.GLog.Error(err.Error())
					return nil, err
				}
				deliveryRechargeCommodityRes.PresentDiamonds = pInfo.GemCount
				deliveryRechargeCommodityRes.NameKey = pInfo.NameKey

				var rechargeNtf csssvr_proto.ProtoCssSvrUserRechargeNtf
				rechargeNtf.ChannelID = reqData.ChannelID
				rechargeNtf.PlatForm = reqData.PlatForm
				rechargeNtf.RechargeAmount = pInfo.Price
				rechargeNtf.Uin = reqData.Uin
				consul_api.PostRequstByConsulDns(reqData.Uin, "CssSvrUserRecharge", &rechargeNtf, common.ConsulClient, "CassandraSvr")
			}
		} else {
			// 月卡已经存在
			err = fmt.Errorf("Uin[%d] have %s status!", reqData.Uin, proto.E_USER_VIP_NORMALMONTHLY.String())
			return nil, err
		}
	case proto.E_RECHARGECOMMODITY_SUPERGIFT:
		deliveryRechargeCommodityRes.RechargeCommodityType = proto.E_RECHARGECOMMODITY_SUPERGIFT
		err = receiveSuperGiftActive(reqData, deliveryRechargeCommodityRes)
		if err != nil {
			return nil, err
		}
	default:
		base.GLog.Critical("Uin[%d] RechargeCommodityType[%d] is invalid!", reqData.Uin, reqData.RechargeCommodityType)
	}

	return deliveryRechargeCommodityRes, nil
}

func BuyShopResourceCommodity(reqData *proto.ProtoBuyResourceCommodityReq) *proto.ProtoBuyResourceCommodityRes {
	// 从redis根据zone，version获取资源商品列表
	shopCommoditiesInfo, err := QueryShopCommoditiesInfo(reqData.ZoneID, proto.E_SHOPCOMMODITY_RESOURCE)
	if err != nil {
		base.GLog.Error("QueryShopCommoditiesInfo failed, commodiy does not exist in redis!")
		return nil
	}

	shopResourceCommoditiesInfo := shopCommoditiesInfo.(*proto.ResourceShopConfigS)

	userFinance, _ := QueryFinanceUser(reqData.Uin)
	if userFinance == nil {
		return nil
	}

	// 查询资源商品
	resourceCommmodityInfo := shopResourceCommoditiesInfo.Find(reqData.ResourceCommdityID)
	if resourceCommmodityInfo == nil {
		return nil
	}

	// 得到消费卡类型
	vipType := GetFinanceUserVIPType(userFinance)

	buyResourceCommodityRes := new(proto.ProtoBuyResourceCommodityRes)
	buyResourceCommodityRes.Uin = reqData.Uin
	buyResourceCommodityRes.ZoneID = reqData.ZoneID
	buyResourceCommodityRes.ResourceCommodityType = resourceCommmodityInfo.ResourceCommodityType
	buyResourceCommodityRes.ResourceCount = resourceCommmodityInfo.ResourceCommodityStackCount
	buyResourceCommodityRes.CostDiamonds = resourceCommmodityInfo.ResourceCommodityDiamondCost

	if vipType&proto.E_USER_VIP_LUXURYMONTHLY != 0 {
		buyResourceCommodityRes.CostDiamonds =
			int32(float32(buyResourceCommodityRes.CostDiamonds) * float32(shopResourceCommoditiesInfo.MonthlyCardDiscountRate) / 100)
	}

	base.GLog.Debug("Uin[%d] ZoneID[%d] buy ResourceCommodity[%d] costDiamonds[%d]", reqData.Uin, reqData.ZoneID,
		reqData.ResourceCommdityID, buyResourceCommodityRes.CostDiamonds)
	return buyResourceCommodityRes
}

func BuyShopCardPackCommodity(reqData *proto.ProtoBuyCardPackCommodityReq) *proto.ProtoBuyCardPackCommodityRes {
	shopCommoditiesInfo, err := QueryShopCommoditiesInfo(reqData.ZoneID, proto.E_SHOPCOMMODITY_CARDPACK)
	if err != nil {
		base.GLog.Error("QueryShopCommoditiesInfo failed, commodiy does not exist in redis!")
		return nil
	}

	shopCardPackCommoditiesInfo := shopCommoditiesInfo.(*proto.ShopCardPackCommoditiesInfoS)

	userFinance, _ := QueryFinanceUser(reqData.Uin)
	if userFinance == nil {
		return nil
	}

	// 查询资源商品
	cardPackCommmodityInfo := shopCardPackCommoditiesInfo.Find(reqData.CardPackCommdityID)
	if cardPackCommmodityInfo == nil {
		return nil
	}

	// 判断玩家是否有月卡，且是否过期
	vipType := GetFinanceUserVIPType(userFinance)

	buyCardPackCommodityRes := new(proto.ProtoBuyCardPackCommodityRes)
	buyCardPackCommodityRes.Uin = reqData.Uin
	buyCardPackCommodityRes.ZoneID = reqData.ZoneID
	buyCardPackCommodityRes.CardPackJsonContent = cardPackCommmodityInfo.CardPackJsonContent
	buyCardPackCommodityRes.CostDiamonds = cardPackCommmodityInfo.CardPackDiamondCost

	if vipType&proto.E_USER_VIP_LUXURYMONTHLY != 0 {
		buyCardPackCommodityRes.CostDiamonds =
			int32(float32(cardPackCommmodityInfo.CardPackDiamondCost) * float32(shopCardPackCommoditiesInfo.MonthlyCardDiscountRate) / 100)
	}

	// 1011卡包做特殊处理
	if reqData.CardPackCommdityID == 1011 {
		if userFinance.ShopFirstPurchaseInfo.IsFirstPurchase(proto.E_SHOPCOMMODITY_CARDPACK, 1011) {
			userFinance.ShopFirstPurchaseInfo.FirstPurchase(proto.E_SHOPCOMMODITY_CARDPACK, 1011, reqData.Uin)
		} else {
			base.GLog.Error("CarPack 1011 can only be purchased once!")
			return nil
		}
	}

	base.GLog.Debug("Uin[%d] ZoneID[%d] buy CardpackCommodity[%d] costDiamonds[%d]", reqData.Uin, reqData.ZoneID,
		reqData.CardPackCommdityID, buyCardPackCommodityRes.CostDiamonds)
	return buyCardPackCommodityRes
}

func getRefreshShopDiscountRate(uin uint64, vipType proto.UserVIPType, shopType proto.RefreshShopType, config *proto.RefreshShopConfigS) int32 {
	var discountRate int32 = 100

	if vipType&proto.E_USER_VIP_LUXURYMONTHLY != 0 {
		if shopType == proto.C_REFRESHSHOPTYPE_NORMAL {
			discountRate = config.CommonShopMonthlyCardDiscountRate
		} else if shopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
			discountRate = config.BreakoutMonthlyCardDiscountRate
		}
	}
	base.GLog.Debug("Uin[%d] vipType[%s] shopType[%s] discountRate[%d]", uin, vipType.String(),
		shopType, discountRate)
	return discountRate
}

func getManualRefreshPrice(uin uint64, manualRefreshCount int32, refreshShopType proto.RefreshShopType, discountRate int32, config *proto.RefreshShopConfigS) (int32, int32) {
	var price int32

	if refreshShopType == proto.C_REFRESHSHOPTYPE_NORMAL {
		if manualRefreshCount < config.DailyCommonManualRefreshCount {
			price = config.CommonManualRefreshCosts[manualRefreshCount]
		} else {
			price = config.CommonManualRefreshCosts[len(config.CommonManualRefreshCosts)-1]
		}
	} else if refreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		if manualRefreshCount < config.DailyBreakoutManualRefreshCount {
			price = config.BreakoutManualRefreshCosts[manualRefreshCount]
		} else {
			price = config.BreakoutManualRefreshCosts[len(config.BreakoutManualRefreshCosts)-1]
		}
	}

	discountPrice := int32(float32(price) * float32(discountRate) / 100)
	base.GLog.Debug("Uin[%d] shopType[%s] manualRefreshCount[%d] price[%d] discountPrice[%d]", uin, refreshShopType, manualRefreshCount, price, discountPrice)
	return price, discountPrice
}

func getAutoRefreshShopTime(userFinance *proto.TblFinanceUserS, refreshShopType proto.RefreshShopType) time.Time {
	refreshTime := userFinance.PlayerRefreshShopDailyInfo.CommonShopAutoRefreshTime
	if refreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		refreshTime = userFinance.PlayerRefreshShopDailyInfo.BreakoutShopAutoRefreshTime
	}
	base.GLog.Debug("Uin[%d] shopType[%s] refreshTime[%s]", userFinance.Uin, refreshShopType, base.TimeName(refreshTime))
	return refreshTime
}

func updateAutoRefreshShopTime(userFinance *proto.TblFinanceUserS, refreshShopType proto.RefreshShopType, config *proto.RefreshShopConfigS,
	now *time.Time, location *time.Location) {
	startHours := common.CalcRefreshStartHours(config.ShopAutoRefreshIntervalHours, int32(now.Hour()))
	if refreshShopType == proto.C_REFRESHSHOPTYPE_NORMAL {
		userFinance.PlayerRefreshShopDailyInfo.CommonShopAutoRefreshTime =
			time.Date(now.Year(), now.Month(), now.Day(), int(startHours), 0, 0, 0, location).Add(time.Duration(config.ShopAutoRefreshIntervalHours) * time.Hour)
		base.GLog.Debug("Uin[%d] shopType[%s] CommonShopAutoRefreshTime[%s]", userFinance.Uin, refreshShopType,
			base.TimeName(userFinance.PlayerRefreshShopDailyInfo.CommonShopAutoRefreshTime))
	} else if refreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		userFinance.PlayerRefreshShopDailyInfo.BreakoutShopAutoRefreshTime =
			time.Date(now.Year(), now.Month(), now.Day(), int(startHours), 0, 0, 0, location).Add(time.Duration(config.ShopAutoRefreshIntervalHours) * time.Hour)
		base.GLog.Debug("Uin[%d] shopType[%s] BreakoutShopAutoRefreshTime[%s]", userFinance.Uin, refreshShopType,
			base.TimeName(userFinance.PlayerRefreshShopDailyInfo.BreakoutShopAutoRefreshTime))
	}
	return
}

func GetRefreshShopCommodities(req *proto.ProtoGetRefreshShopCommoditiesReq) (*proto.ProtoGetRefreshShopCommoditiesRes, error) {
	var needSyncDB bool = false
	var err error
	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		return nil, fmt.Errorf("Uin[%d] is not exist!", req.Uin)
	}

	// 查询刷新商店配置
	refreshShopConfig := getRefreshShopConfig(req.ZoneID)
	if refreshShopConfig == nil {
		return nil, fmt.Errorf("Uin[%d] getRefreshShopConfig failed!", req.Uin)
	}

	vipType := GetFinanceUserVIPType(userFinance)
	discountRate := getRefreshShopDiscountRate(req.Uin, vipType, req.RefreshShopType, refreshShopConfig)

	var res proto.ProtoGetRefreshShopCommoditiesRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.RefreshShopType = req.RefreshShopType
	res.IsManualRefresh = req.IsManualRefresh

	var remainderManualRefreshCount int32
	if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_NORMAL {
		res.ManualRefreshPayType = refreshShopConfig.CommonManualRefreshPayType
		// 手动刷新的价格
		res.ManualRefreshPrice, res.ManualRefreshDiscountPrice = getManualRefreshPrice(req.Uin,
			userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount,
			req.RefreshShopType, discountRate, refreshShopConfig)
		// 刷新总次数
		res.ManualRefreshCount = refreshShopConfig.DailyCommonManualRefreshCount
		// 自动刷新期间可购买的数量
		res.CommodityBuyCountInPeriod = refreshShopConfig.CommonShopCommodityDailyBuyCount
		// 不同用户的折扣率
		res.WeeklyCardDiscountRate = refreshShopConfig.CommonShopWeeklyCardDiscountRate
		res.MonthlyCardDiscountRate = refreshShopConfig.CommonShopMonthlyCardDiscountRate
		// 剩余刷新次数
		remainderManualRefreshCount = refreshShopConfig.DailyCommonManualRefreshCount - userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount
	} else if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		res.ManualRefreshPayType = refreshShopConfig.BreakoutManualRefreshPayType
		//
		res.ManualRefreshPrice, res.ManualRefreshDiscountPrice = getManualRefreshPrice(req.Uin,
			userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount,
			req.RefreshShopType, discountRate, refreshShopConfig)
		//
		res.ManualRefreshCount = refreshShopConfig.DailyBreakoutManualRefreshCount
		//
		res.CommodityBuyCountInPeriod = refreshShopConfig.BreakoutCommodityDailyBuyCount
		res.WeeklyCardDiscountRate = refreshShopConfig.BreakoutWeeklyCardDiscountRate
		res.MonthlyCardDiscountRate = refreshShopConfig.BreakoutMonthlyCardDiscountRate
		//
		remainderManualRefreshCount = refreshShopConfig.DailyBreakoutManualRefreshCount - userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount
	} else {
		err = fmt.Errorf("req.RefreshShopType[%s] is invalid", req.RefreshShopType)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 判断是否是自动刷新
	location, _ := time.LoadLocation(userFinance.TimeZone)
	now := time.Now().In(location)
	autoRefreshTime := getAutoRefreshShopTime(userFinance, req.RefreshShopType)
	if now.After(autoRefreshTime) {
		// 计算下次自动刷新时间
		base.GLog.Debug("Uin[%d] autoRefreshTime[%s] Now[%s] shopType[%s] will AutoRefresh",
			req.Uin, base.TimeName(autoRefreshTime), base.TimeName(now), req.RefreshShopType)
		updateAutoRefreshShopTime(userFinance, req.RefreshShopType, refreshShopConfig, &now, location)

		// 刷新显示的商品
		err = updateRefreshShopPresentCommodities(userFinance, req.ZoneID, req.RefreshShopType, refreshShopConfig)
		if err != nil {
			return nil, err
		}
		// 清空已经购买的列表
		if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_NORMAL {
			userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.AlreadyPurchasedCommodities =
				make([]proto.AlreadyPurchasedCommodityInfoS, 0)
			// 刷新次数清零
			userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount = 0
		}

		if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
			userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.AlreadyPurchasedCommodities =
				make([]proto.AlreadyPurchasedCommodityInfoS, 0)
			userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount = 0
		}

		// 重新计算刷新价格
		res.ManualRefreshPrice, res.ManualRefreshDiscountPrice = getManualRefreshPrice(req.Uin,
			0, req.RefreshShopType, discountRate, refreshShopConfig)

		needSyncDB = true
	} else {
		// 如果是手动刷新
		if req.IsManualRefresh == 1 {
			if remainderManualRefreshCount > 0 {
				needSyncDB = true

				// 刷新的花费，月卡，普通月卡有折扣率
				// 计算本次刷新的消费
				if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_NORMAL {
					// 本次刷新的花费
					res.ManualRefreshCost =
						int32(float32(refreshShopConfig.CommonManualRefreshCosts[userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount]) * float32(discountRate) / 100)
					userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount++
					userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.AlreadyPurchasedCommodities =
						make([]proto.AlreadyPurchasedCommodityInfoS, 0)

					// 手动刷新后需要重新计算价格，包括折扣后的价格
					res.ManualRefreshPrice, res.ManualRefreshDiscountPrice = getManualRefreshPrice(req.Uin,
						userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount,
						req.RefreshShopType, discountRate, refreshShopConfig)
				} else if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
					res.ManualRefreshCost =
						int32(float32(refreshShopConfig.BreakoutManualRefreshCosts[userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount]) * float32(discountRate) / 100)
					userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount++
					// 清空已经购买的列表
					userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.AlreadyPurchasedCommodities =
						make([]proto.AlreadyPurchasedCommodityInfoS, 0)

					res.ManualRefreshPrice, res.ManualRefreshDiscountPrice = getManualRefreshPrice(req.Uin,
						userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount,
						req.RefreshShopType, discountRate, refreshShopConfig)
				}

				// 刷新显示的商品
				err = updateRefreshShopPresentCommodities(userFinance, req.ZoneID, req.RefreshShopType, refreshShopConfig)
				if err != nil {
					base.GLog.Error("Uin[%d] updateRefreshShopPresentCommodities failed! reason[%s]", req.Uin, err.Error())
					return nil, err
				}

				base.GLog.Debug("Uin[%d] ManualRefreshShop discountRate[%d] vipType[%s] commonShopManualRefreshCount[%d] breakoutShopManualRefreshCount[%d]",
					req.Uin, discountRate, vipType.String(), userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount,
					userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount)
			} else {
				err = fmt.Errorf("Uin[%d] remainderManualRefreshCount[%d] cannot manually refresh!!", req.Uin, remainderManualRefreshCount)
				base.GLog.Warn(err.Error())
				return nil, err
			}
		}
	}

	if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_NORMAL {
		res.AlreadyPurchasedCommodities = userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.AlreadyPurchasedCommodities
		res.RefreshShopDisplayCommodities =
			userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.CurrentDisplayCommodities
		// 计算剩余刷新次数
		res.ManualRefreshRemainderCount = refreshShopConfig.DailyCommonManualRefreshCount -
			userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount
		// 计算刷新剩余时间，单位秒
		res.AutoRefreshRemainderSeconds = int64(userFinance.PlayerRefreshShopDailyInfo.CommonShopAutoRefreshTime.Sub(now).Seconds())
	} else if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		res.AlreadyPurchasedCommodities = userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.AlreadyPurchasedCommodities
		res.RefreshShopDisplayCommodities =
			userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.CurrentDisplayCommodities
		// 计算剩余刷新次数
		res.ManualRefreshRemainderCount = refreshShopConfig.DailyBreakoutManualRefreshCount -
			userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount
		res.AutoRefreshRemainderSeconds = int64(userFinance.PlayerRefreshShopDailyInfo.BreakoutShopAutoRefreshTime.Sub(now).Seconds())
	}

	//base.GLog.Debug("res:%+v", res)

	// 更新玩家数据库
	if needSyncDB {
		userFinance.PlayerRefreshShopDailyInfo.SyncDB(common.GDBEngine, req.Uin)
	}

	return &res, nil
}

// 从池中剩余的商品中按权重调出chooseCount个商品来
func chooseRefreshShopCommodities(commodityPickLst []*proto.RefreshShopCommodityS, chooseCount int, poolName string) ([]*proto.RefreshShopCommodityS, error) {
	wantChooseCount := chooseCount
	commodityPickLstCount := len(commodityPickLst)
	base.GLog.Debug("commodityPickLstCount[%d] chooseCount[%d] from Pool[%s]", commodityPickLstCount, chooseCount, poolName)

	//
	if commodityPickLstCount < chooseCount {
		err := fmt.Errorf("Pool[%s] commodityPickLstCount[%d] less than chooseCount[%d]", poolName, commodityPickLstCount, chooseCount)
		base.GLog.Error(err.Error())
		return nil, err
	}

	if commodityPickLstCount == chooseCount {
		base.GLog.Debug("Pool[%s] commodityPickLstCount == chooseCount! so direct return", poolName)
		return commodityPickLst, nil
	}

	// 计算总的几率值
	var totalCommodityChance int
	for _, commodity := range commodityPickLst {
		totalCommodityChance += commodity.CommodityChance
	}

	base.GLog.Debug("Pool[%s] totalCommodityChance[%d]", poolName, totalCommodityChance)

	// for _, commodity := range commodityPickLst {
	// 	chance := float32(commodity.CommodityChance)
	// 	commodity.CommodityChance = int((chance/totalCommodityChance + 0.005) * 100)
	// 	base.GLog.Debug("commodity[%d] chance[%f] CommodityChance[%d]", commodity.CommodityID, chance, commodity.CommodityChance)
	// }

	lastPos := commodityPickLstCount - 1
	var chanceRandomUpperBound = totalCommodityChance

	for chooseCount > 0 {
		r := rand.Intn(chanceRandomUpperBound)
		base.GLog.Debug("chanceRandomUpperBound[%d] r[%d]", chanceRandomUpperBound, r)
		for index, commodity := range commodityPickLst[:lastPos+1] {

			if r <= commodity.CommodityChance {
				// 交换
				chanceRandomUpperBound -= commodity.CommodityChance
				base.GLog.Debug("---------Pool[%s] Select commodityID[%d] CommodityChance[%d] swap(index[%d]<===>lastPos[%d]) chanceRandomUpperBound[%d]-----------",
					poolName, commodity.CommodityID, commodity.CommodityChance, index, lastPos, chanceRandomUpperBound)
				commodityPickLst[index], commodityPickLst[lastPos] = commodityPickLst[lastPos], commodityPickLst[index]
				lastPos--
				break
			} else {
				r -= commodity.CommodityChance
				base.GLog.Debug("surplus-r[%d] commodity[%d] CommodityChance[%d]", r, commodity.CommodityID, commodity.CommodityChance)
			}
			//base.GLog.Debug("r[%d]", r)
		}
		chooseCount--
	}

	// 保护
	selectCount := len(commodityPickLst[lastPos+1:])
	if selectCount < wantChooseCount {
		base.GLog.Warn("*********************selectCount[%d] less wantChooseCount[%d]", selectCount, wantChooseCount)
		lastPos--
	}

	base.GLog.Debug("Pool[%s] lastPos[%d] SelectCount[%d] commodityPickLst:%v", poolName, lastPos, len(commodityPickLst[lastPos+1:]), commodityPickLst[lastPos+1:])

	return commodityPickLst[lastPos+1:], nil
}

func updateRefreshShopPresentCommodities(userFinance *proto.TblFinanceUserS, zoneID int32,
	refreshShopType proto.RefreshShopType, config *proto.RefreshShopConfigS) error {

	// 橱窗绑定的pool列表
	commodityPools := config.CommonShopCommodityPools
	// 橱窗展示商品个数
	presentCommodityCount := config.CommonShopPresentCommodityCount
	// 玩家当前显示的商品集合
	currDisplayCommodityList := &userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.CurrentDisplayCommodities

	if refreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		commodityPools = config.BreakoutShopCommodityPools
		presentCommodityCount = config.BreakoutShopPresentCommodityCount
		currDisplayCommodityList = &userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.CurrentDisplayCommodities
	}

	// 更新的结果集
	newPresentCommodityCount := 0
	newPresentCommodityList := make([]proto.RefreshShopCommodityS, presentCommodityCount)

	base.GLog.Debug("+++++++++++Start Select [%s] [%d]Commodities+++++++++", refreshShopType, presentCommodityCount)
	for index, _ := range commodityPools {
		poolInfo := &commodityPools[index]
		// 池子关联的slot个数
		poolAssocSlotCount := len(poolInfo.DisplaySlotIndexs)
		base.GLog.Debug("Pool[%s] poolAssocSlotCount[%d]", poolInfo.PoolName, poolAssocSlotCount)

		// 得到商品池中的商品
		poolCommodities, err := getRefreshShopCommodityPool(zoneID, poolInfo.PoolName)
		if err != nil {
			return err
		}

		// 要排除当前显示的商品，得到待挑选的集合
		waitSelectedCommodityList := make([]*proto.RefreshShopCommodityS, 0)
		for i, _ := range poolCommodities {
			refreshShopCommodity := &poolCommodities[i]
			inDisplayList := false
			for j, _ := range *currDisplayCommodityList {
				if refreshShopCommodity.CommodityID == (*currDisplayCommodityList)[j].CommodityID {
					inDisplayList = true
				}
			}
			if !inDisplayList {
				waitSelectedCommodityList = append(waitSelectedCommodityList, refreshShopCommodity)
			}
		}

		// 从待挑选的商品列表中调出poolAssocSlotCount个商品来
		selectedCommodityList, err := chooseRefreshShopCommodities(waitSelectedCommodityList, poolAssocSlotCount, poolInfo.PoolName)
		if err != nil {
			return err
		}

		for i, _ := range selectedCommodityList {
			base.GLog.Debug("append newPresentCommodity[%d] %+v", newPresentCommodityCount, *(selectedCommodityList[i]))
			newPresentCommodityList[newPresentCommodityCount] = *(selectedCommodityList[i])
			newPresentCommodityCount++
		}
	}

	// 更新用户的db数据
	*currDisplayCommodityList = newPresentCommodityList
	return nil
}

func GetRefreshShopCommodityCost(req *proto.ProtoGetRefreshShopCommodityCostReq) (*proto.ProtoGetRefreshShopCommodityCostRes, error) {
	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		return nil, fmt.Errorf("Uin[%d] is not exist!", req.Uin)
	}
	// 得到消费卡类型
	vipType := GetFinanceUserVIPType(userFinance)

	// 查询刷新商店配置
	refreshShopConfig := getRefreshShopConfig(req.ZoneID)
	if refreshShopConfig == nil {
		return nil, fmt.Errorf("Uin[%d] getRefreshShopConfig failed!", req.Uin)
	}

	// 根据poolkey查询对应的商品pool
	commodityList, err := getRefreshShopCommodityPoolByKey(req.PoolKey)
	if err != nil {
		return nil, err
	}

	var refreshShopCommodity *proto.RefreshShopCommodityS = nil
	for index, _ := range commodityList {
		if commodityList[index].CommodityID == req.CommodityID {
			refreshShopCommodity = &commodityList[index]
		}
	}

	if refreshShopCommodity == nil {
		err = fmt.Errorf("BuyRefreshShopCommodity commodityID[%d] is not exist!", req.CommodityID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	res := new(proto.ProtoGetRefreshShopCommodityCostRes)
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.CommodityID = req.CommodityID
	res.GamePayType = refreshShopCommodity.GamePayType
	discountRate := getRefreshShopDiscountRate(req.Uin, vipType, req.RefreshShopType, refreshShopConfig)
	res.CommodityPrice = refreshShopCommodity.CommodityPrice * discountRate / 100
	return res, nil
}

func BuyRefreshShopCommodity(req *proto.ProtoBuyRefreshShopCommodityReq) (*proto.ProtoBuyRefreshShopCommodityRes, error) {
	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		return nil, fmt.Errorf("Uin[%d] is not exist!", req.Uin)
	}
	// 得到消费卡类型
	vipType := GetFinanceUserVIPType(userFinance)

	// 查询刷新商店配置
	refreshShopConfig := getRefreshShopConfig(req.ZoneID)
	if refreshShopConfig == nil {
		return nil, fmt.Errorf("Uin[%d] getRefreshShopConfig failed!", req.Uin)
	}

	// 根据poolkey查询对应的商品pool
	commodityList, err := getRefreshShopCommodityPoolByKey(req.PoolKey)
	if err != nil {
		return nil, err
	}

	//base.GLog.Debug("commodityList:%+v", commodityList)

	// 根据commodityid查找对应的商品
	var refreshShopCommodity *proto.RefreshShopCommodityS = nil
	for index, _ := range commodityList {
		if commodityList[index].CommodityID == req.CommodityID {
			refreshShopCommodity = &commodityList[index]
		}
	}

	if refreshShopCommodity == nil {
		err = fmt.Errorf("BuyRefreshShopCommodity commodityID[%d] is not exist!", req.CommodityID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	commodityBuyCountInPeriod := refreshShopConfig.CommonShopCommodityDailyBuyCount
	alreadyPurchasedCommodities := &userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.AlreadyPurchasedCommodities
	if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		alreadyPurchasedCommodities = &userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.AlreadyPurchasedCommodities
		commodityBuyCountInPeriod = refreshShopConfig.BreakoutCommodityDailyBuyCount
	}

	// 判断是否已经购买过
	var isPurchased bool = false
	for index, _ := range *alreadyPurchasedCommodities {
		if (*alreadyPurchasedCommodities)[index].CommodityID == req.CommodityID {
			isPurchased = true
			if (*alreadyPurchasedCommodities)[index].PurchasedCount >= commodityBuyCountInPeriod {
				err = fmt.Errorf("Uin[%d] commodityID[%d] had purchased count[%d] in refresh period", req.Uin, req.CommodityID, commodityBuyCountInPeriod)
				base.GLog.Error(err.Error())
				return nil, err
			} else {
				// 递增
				(*alreadyPurchasedCommodities)[index].PurchasedCount++
			}
		}
	}

	// 加入已经购买列表
	if !isPurchased {
		*alreadyPurchasedCommodities = append(*alreadyPurchasedCommodities, proto.AlreadyPurchasedCommodityInfoS{
			CommodityID:    req.CommodityID,
			PurchasedCount: 1,
		})
	}

	// 同步到数据库
	userFinance.PlayerRefreshShopDailyInfo.SyncDB(common.GDBEngine, req.Uin)

	res := new(proto.ProtoBuyRefreshShopCommodityRes)
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.CommodityID = req.CommodityID
	res.GamePayType = refreshShopCommodity.GamePayType
	res.CommodityJsonContent = refreshShopCommodity.CommodityJsonContent
	discountRate := getRefreshShopDiscountRate(req.Uin, vipType, req.RefreshShopType, refreshShopConfig)
	res.CommodityPrice = refreshShopCommodity.CommodityPrice * discountRate / 100
	return res, nil
}

func CheckRefreshShopManualRefresh(req *proto.ProtoCheckManualRefreshReq) (*proto.ProtoCheckManualRefreshRes, error) {
	var err error
	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		return nil, fmt.Errorf("Uin[%d] is not exist!", req.Uin)
	}

	// 查询刷新商店配置
	refreshShopConfig := getRefreshShopConfig(req.ZoneID)
	if refreshShopConfig == nil {
		return nil, fmt.Errorf("Uin[%d] getRefreshShopConfig failed!", req.Uin)
	}

	vipType := GetFinanceUserVIPType(userFinance)
	discountRate := getRefreshShopDiscountRate(req.Uin, vipType, req.RefreshShopType, refreshShopConfig)

	var res proto.ProtoCheckManualRefreshRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID

	if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_NORMAL {
		// 手动刷新的价格
		_, res.ManualRefreshCost = getManualRefreshPrice(req.Uin, userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount,
			req.RefreshShopType, discountRate, refreshShopConfig)
		// 剩余次数
		res.ManualRefreshRemainderCount = refreshShopConfig.DailyCommonManualRefreshCount - userFinance.PlayerRefreshShopDailyInfo.PlayerCommonShopDailyInfo.ManualRefreshCount
		res.ManualRefreshPayType = refreshShopConfig.CommonManualRefreshPayType
	} else if req.RefreshShopType == proto.C_REFRESHSHOPTYPE_BREAKOUT {
		_, res.ManualRefreshCost = getManualRefreshPrice(req.Uin, userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount,
			req.RefreshShopType, discountRate, refreshShopConfig)
		res.ManualRefreshRemainderCount = refreshShopConfig.DailyBreakoutManualRefreshCount - userFinance.PlayerRefreshShopDailyInfo.PlayerBreakoutShopDailyInfo.ManualRefreshCount
		res.ManualRefreshPayType = refreshShopConfig.BreakoutManualRefreshPayType
	} else {
		err = fmt.Errorf("req.RefreshShopType[%s] is invalid", req.RefreshShopType)
		base.GLog.Error(err.Error())
		return nil, err
	}
	return &res, nil
}

func QueryRechargeCommodityPrices(req *proto.ProtoQueryRechargeCommodityPricesReq) (*proto.ProtoQueryRechargeCommodityPricesRes, error) {
	var res proto.ProtoQueryRechargeCommodityPricesRes
	res.ZoneID = req.ZoneID
	res.RechargeCommodityType = req.RechargeCommodityType
	res.ChannelID = req.ChannelID
	res.RechargeCommodityID = req.RechargeCommodityID

	switch req.RechargeCommodityType {
	case proto.E_RECHARGECOMMODITY_DIAMONDS:
		// 查询钻石商品价格
		shopCommoditiesInfo, err := QueryShopCommoditiesInfo(req.ZoneID, proto.E_SHOPCOMMODITY_RECHARGE)
		if err != nil {
			base.GLog.Error("QueryShopCommoditiesInfo failed, commodiy does not exist in redis!")
			return nil, err
		}

		shopRechargeCommoditiesInfo := shopCommoditiesInfo.(*proto.ShopRechargeCommoditiesInfoS)
		// 在充值商品列表中查询具体的商品
		rechargeCommmodityInfo := shopRechargeCommoditiesInfo.FindRechargeCommodity(req.ChannelID, req.RechargeCommodityID)
		if rechargeCommmodityInfo == nil {
			return nil, fmt.Errorf("ChannelID[%s] CommodityID[%d] is invalid!", req.ChannelID, req.RechargeCommodityID)
		}
		res.Price = rechargeCommmodityInfo.Price

	case proto.E_RECHARGECOMMODITY_SUPERGIFT:
		// 查询首充礼包商品价格
		activeInstConfigKey := fmt.Sprintf(proto.ActiveInstConfigKeyFmt, req.ZoneID, proto.E_ACTIVETYPE_SUPERGIFTPACKAGE.String(),
			req.RechargeCommodityID)

		// 超级礼包的配置
		activeInstConfig := new(proto.ActiveSuperGiftInfoS)
		err := common.GetStrDataFromRedis(activeInstConfigKey, activeInstConfig)
		if err != nil {
			return nil, err
		}

		res.Price = activeInstConfig.Price

	case proto.E_RECHARGECOMMODITY_NORMALMONTHLYCARD:
		fallthrough
	case proto.E_RECHARGECOMMODITY_LUXURYMONTHLYCARD:
		// 查询月卡商品价格
		vipPrivilegeConfig := getVIPPrivilegeConfig(req.ZoneID)
		if vipPrivilegeConfig == nil {
			err := fmt.Errorf("ZoneID[%d] VIPPrivilegeConfig is not configured!", req.ZoneID)
			return nil, err
		}

		vipPrivilegeInfo := vipPrivilegeConfig.FindVIPPrivilege(req.ChannelID, req.RechargeCommodityID)
		if vipPrivilegeInfo == nil {
			err := fmt.Errorf("MonthCard RechargeCommodity cannot find!")
			return nil, err
		}

		res.Price = vipPrivilegeInfo.Price
	}

	return &res, nil
}
