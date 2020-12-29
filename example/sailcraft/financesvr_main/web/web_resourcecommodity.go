/*
 * @Author: calmwu
 * @Date: 2018-02-07 12:20:23
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:21:25
 */
package web

import (
	"encoding/json"
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) QueryResourceCommdities(c *gin.Context) {
	var reqData proto.ProtoQueryResourceCommoditiesReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "QueryResourceCommdities", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	shopCommoditiesInfo, err := handler.QueryShopCommoditiesInfo(reqData.ZoneID, proto.E_SHOPCOMMODITY_RESOURCE)
	if err != nil {
		resFuncParams.RetCode = -1
	} else {
		shopResourceCommoditiesInfo := shopCommoditiesInfo.(*proto.ResourceShopConfigS)
		var resData proto.ProtoQueryResourceCommoditiesRes
		resData.Uin = reqData.Uin
		resData.ZoneID = reqData.ZoneID
		resData.ResourceShopConfigInfo = shopResourceCommoditiesInfo
		resFuncParams.Param = resData
	}
}

func (fw *FinanceWebModule) RefreshResourceShopConfig(c *gin.Context) {
	// 判断rediskey是否存在，存在就报错，不能更改原有版本，只能递增版本
	var reqData proto.ProtoRefreshResourceShopConfigReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "RefreshResourceShopConfig", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	// 插入新的充值商品信息，同时更新当前商品版本号
	redisData, err := json.Marshal(reqData.ResourceShopConfigInfo)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
		return
	}

	err = handler.UpdateShopCommoditiesInfo(reqData.ZoneID, reqData.ResourceShopConfigInfo.VersionID, redisData, proto.E_SHOPCOMMODITY_RESOURCE)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
		return
	}
}

func (fw *FinanceWebModule) BuyResourceCommodity(c *gin.Context) {
	var reqData proto.ProtoBuyResourceCommodityReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "BuyResourceCommodity", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes := handler.BuyShopResourceCommodity(&reqData)
	if hRes == nil {
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
