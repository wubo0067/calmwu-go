/*
 * @Author: calmwu
 * @Date: 2018-04-16 17:25:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:16:34
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) GMConfigFirstRecharge(c *gin.Context) {
	var reqData proto.ProtoGMConfigFirstRechargeReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMConfigFirstRecharge", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMConfigFirstRecharge(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GetFirstRechargeActive(c *gin.Context) {
	var reqData proto.ProtoGetFirstRechargeActiveReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetFirstRechargeActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetFirstRechargeActive(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) ReceiveFirstRechargeReward(c *gin.Context) {
	var reqData proto.ProtoReceiveFirstRechargeRewardReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "ReceiveFirstRechargeReward", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.ReceiveFirstRechargeReward(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
