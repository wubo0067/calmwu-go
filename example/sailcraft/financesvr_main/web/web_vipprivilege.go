/*
 * @Author: calmwu
 * @Date: 2018-03-23 16:57:59
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:21:52
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) GMConfigVIPPrivilege(c *gin.Context) {
	var reqData proto.ProtoGMConfigVIPPrivilegeReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMConfigVIPPrivilege", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMConfigVIPPrivilege(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GetPlayerVIPInfo(c *gin.Context) {
	var reqData proto.ProtoGetPlayerVIPInfoReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetPlayerVIPInfo", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetPlayerVIPInfo(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) VIPPlayerCollectPrize(c *gin.Context) {
	var reqData proto.ProtoVIPPlayerCollectPrizeReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "VIPPlayerCollectPrize", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.VIPPlayerCollectPrize(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
