/*
 * @Author: calmwu
 * @Date: 2018-02-11 10:23:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:23:29
 */

package web

import (
	"encoding/json"
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) QueryCardPackCommdities(c *gin.Context) {
	var reqData proto.ProtoQueryCardPackCommoditiesReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "QueryCardPackCommdities", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	shopCommoditiesInfo, err := handler.QueryShopCommoditiesInfo(reqData.ZoneID, proto.E_SHOPCOMMODITY_CARDPACK)
	if err != nil {
		resFuncParams.RetCode = -1
	} else {
		userFinance, _ := handler.QueryFinanceUser(reqData.Uin)
		if userFinance == nil {
			base.GLog.Error("Uin[%d] is not exists!", reqData.Uin)
			resFuncParams.RetCode = -1
			return
		}

		shopCardPackCommoditiesInfo := shopCommoditiesInfo.(*proto.ShopCardPackCommoditiesInfoS)
		if !userFinance.ShopFirstPurchaseInfo.IsFirstPurchase(proto.E_SHOPCOMMODITY_CARDPACK, 1011) {
			shopCardPackCommoditiesInfo.Remove(1011)
		}
		var resData proto.ProtoQueryCardPackCommoditiesRes
		resData.Uin = reqData.Uin
		resData.ZoneID = reqData.ZoneID
		resData.ShopCardPackCommoditiesInfo = shopCardPackCommoditiesInfo
		resFuncParams.Param = resData
	}
}

func (fw *FinanceWebModule) RefreshCardPackShopConfig(c *gin.Context) {
	// 判断rediskey是否存在，存在就报错，不能更改原有版本，只能递增版本
	var reqData proto.ProtoRefreshCardPackShopReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "RefreshCardPackShopConfig", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	// 插入新的充值商品信息，同时更新当前商品版本号
	redisData, err := json.Marshal(reqData.ShopCardPackCommoditiesInfo)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
		return
	}

	err = handler.UpdateShopCommoditiesInfo(reqData.ZoneID, reqData.ShopCardPackCommoditiesInfo.VersionID, redisData, proto.E_SHOPCOMMODITY_CARDPACK)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
		return
	}
}

func (fw *FinanceWebModule) BuyCardPackCommodity(c *gin.Context) {
	var reqData proto.ProtoBuyCardPackCommodityReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "BuyCardPackCommodity", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	resData := handler.BuyShopCardPackCommodity(&reqData)

	if resData == nil {
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = resData
	}
}
