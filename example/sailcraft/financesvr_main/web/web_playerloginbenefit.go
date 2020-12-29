/*
 * @Author: calmwu
 * @Date: 2018-03-28 10:58:10
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:17:34
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) GMConfigNewPlayerLoginBenefit(c *gin.Context) {
	var reqData proto.ProtoGMConfigNewPlayerLoginBenefitsReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMConfigNewPlayerLoginBenefit", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMConfigNewPlayerLoginBenefit(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GetNewPlayerLoginBenefitInfo(c *gin.Context) {
	var reqData proto.ProtoGetNewPlayerLoginBenefitReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetNewPlayerLoginBenefitInfo", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetNewPlayerLoginBenefitInfo(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) ReceiveLoginBenefit(c *gin.Context) {
	var reqData proto.ProtoReceiveLoginBenefitReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "ReceiveLoginBenefit", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.ReceiveLoginBenefit(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
