/*
 * @Author: calmwu
 * @Date: 2018-03-15 10:47:38
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:18:20
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

// GM工具刷新每日签到配置
func (fw *FinanceWebModule) GMUpdateMonthlySigninConfigInfo(c *gin.Context) {
	var reqData proto.ProtoGMConfigMonthlySignInReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMUpdateMonthlySigninConfigInfo", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMUpdateMonthlySigninConfigInfo(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GetMonthlySigninInfo(c *gin.Context) {
	var reqData proto.ProtoGetMonthlySigninInfoReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetMonthlySigninInfo", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetMonthlySigninInfo(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) PlayerSignIn(c *gin.Context) {
	var reqData proto.ProtoPlayerSignInReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "PlayerSignIn", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.PlayerSignIn(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
