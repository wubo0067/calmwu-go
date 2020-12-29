/*
 * @Author: calmwu
 * @Date: 2018-02-01 17:59:21
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:19:01
 * @Comment:
 */

package web

import (
	"encoding/json"
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) QueryRechargeCommodities(c *gin.Context) {
	var reqData proto.ProtoQueryRechargeCommoditiesReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "QueryRechargeCommodities", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	userFinance, err := handler.QueryFinanceUser(reqData.Uin)
	if userFinance == nil {
		base.GLog.Error("Uin[%d] is not exists!", reqData.Uin)
		resFuncParams.RetCode = -1
		return
	}

	shopCommoditiesInfo, err := handler.QueryShopCommoditiesInfo(reqData.ZoneID, proto.E_SHOPCOMMODITY_RECHARGE)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
		return
	} else {
		shopRechargeCommoditiesInfo := shopCommoditiesInfo.(*proto.ShopRechargeCommoditiesInfoS)
		channelRechargeCommodities := shopRechargeCommoditiesInfo.FindChannelRechargeCommodities(reqData.ChannelID)
		if channelRechargeCommodities != nil {
			var resData proto.ProtoQueryRechargeCommoditiesRes
			resData.Uin = reqData.Uin
			resData.ZoneID = reqData.ZoneID
			resData.ChannelID = reqData.ChannelID
			resData.VersionID = shopRechargeCommoditiesInfo.VersionID
			resData.RechargeCommodities = channelRechargeCommodities.RechargeCommodities

			// 这里要设置是否购买过的标志位
			for i, _ := range resData.RechargeCommodities {
				rechargeCommodity := &resData.RechargeCommodities[i]

				if userFinance.ShopFirstPurchaseInfo.IsFirstPurchase(proto.E_SHOPCOMMODITY_RECHARGE, rechargeCommodity.RechargeCommodityID) {
					rechargeCommodity.PurchasedFlag = 0
				} else {
					rechargeCommodity.PurchasedFlag = 1
				}
			}
			resFuncParams.Param = resData
			//base.GLog.Debug("%s", shopRechargeCommoditiesInfo.String())
		} else {
			failInfo := new(base.ProtoFailInfoS)
			failInfo.FailureReason = err.Error()
			resFuncParams.Param = failInfo
			resFuncParams.RetCode = -1
		}
	}
}

// 商店商品发货
func (fw *FinanceWebModule) DeliveryRechargeCommodity(c *gin.Context) {
	var reqData proto.ProtoDeliveryRechargeCommodityReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "DeliveryRechargeCommodity", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	//base.GLog.Debug("reqData:%+v", reqData)

	hRes, err := handler.DeliveryRechargeCommodity(&reqData)
	if hRes == nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) RefreshRechargeCommodities(c *gin.Context) {
	// 判断rediskey是否存在，存在就报错，不能更改原有版本，只能递增版本
	var reqData proto.ProtoRefreshRechargeCommoditiesReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "RefreshRechargeCommodities", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	// 插入新的充值商品信息，同时更新当前商品版本号
	redisData, err := json.Marshal(reqData.ShopRechargeCommoditiesInfo)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
		return
	}

	err = handler.UpdateShopCommoditiesInfo(reqData.ZoneID, reqData.ShopRechargeCommoditiesInfo.VersionID, redisData, proto.E_SHOPCOMMODITY_RECHARGE)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) QueryRechargeCommodityPrices(c *gin.Context) {
	var reqData proto.ProtoQueryRechargeCommodityPricesReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "QueryRechargeCommodityPrices", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.QueryRechargeCommodityPrices(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
