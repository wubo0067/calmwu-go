/*
 * @Author: calmwu
 * @Date: 2018-02-22 18:50:46
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:24:39
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) GetRefreshShopCommodities(c *gin.Context) {
	var reqData proto.ProtoGetRefreshShopCommoditiesReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetRefreshShopCommodities", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	commodityInfos, err := handler.GetRefreshShopCommodities(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.RetCode = -1
		resFuncParams.Param = failInfo
		return
	} else {
		resFuncParams.Param = commodityInfos
	}
}

// 更新刷新商店的配置
func (fw *FinanceWebModule) UpdateRefreshShopConfig(c *gin.Context) {
	var reqData proto.ProtoUpdateRefreshShopConfigReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "UpdateRefreshShopConfig", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.UpdateRefreshShopConfig(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.RetCode = -1
		resFuncParams.Param = failInfo
	}
}

// GM工具刷新商品池
func (fw *FinanceWebModule) UpdateRefreshShopCommodityPool(c *gin.Context) {
	var reqData proto.ProtoUpdateRefreshShopCommodityPoolReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "UpdateRefreshShopCommodityPool", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.UpdateRefreshShopCommodityPool(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GetRefreshShopCommodityCost(c *gin.Context) {
	var reqData proto.ProtoGetRefreshShopCommodityCostReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetRefreshShopCommodityCost", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetRefreshShopCommodityCost(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) BuyRefreshShopCommodity(c *gin.Context) {
	var reqData proto.ProtoBuyRefreshShopCommodityReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "BuyRefreshShopCommodity", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.BuyRefreshShopCommodity(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) CheckRefreshShopManualRefresh(c *gin.Context) {
	var reqData proto.ProtoCheckManualRefreshReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "CheckRefreshShopManualRefresh", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.CheckRefreshShopManualRefresh(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
