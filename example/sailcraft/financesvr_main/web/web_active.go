/*
 * @Author: calmwu
 * @Date: 2018-03-30 10:22:46
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:14:33
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) OpenActive(c *gin.Context) {
	var reqData proto.ProtoOpenActiveReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "OpenActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	resFuncParams.Param = handler.OpenActive(&reqData)
}

func (fw *FinanceWebModule) CloseActive(c *gin.Context) {
	var reqData proto.ProtoCloseActiveReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "CloseActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}
	resFuncParams.Param = handler.CloseActive(&reqData)
}

func (fw *FinanceWebModule) GMConfigSuperGiftActive(c *gin.Context) {
	var reqData proto.ProtoGMConfigActiveSuperGiftReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMConfigSuperGiftActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMConfigSuperGiftActive(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GMConfigMissionActive(c *gin.Context) {
	var reqData proto.ProtoGMConfigActiveMissionReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMConfigMissionActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMConfigMissionActive(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GMConfigExchangeActive(c *gin.Context) {
	var reqData proto.ProtoGMConfigActiveExchangeReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMConfigExchangeActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMConfigExchangeActive(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GMConfigCDKeyExchangeActive(c *gin.Context) {
	var reqData proto.ProtoGMConfigActiveCDKeyExchangeReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GMConfigCDKeyExchangeActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	err = handler.GMConfigCDKeyExchangeActive(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *FinanceWebModule) GetPlayerActive(c *gin.Context) {
	var reqData proto.ProtoGetPlayerActiveReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetPlayerActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetPlayerActive(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) ActiveAccumulateParameterNtf(c *gin.Context) {
	var reqData proto.ProtoActiveAccumulateParameterNtf
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "ActiveAccumulateParameterNtf", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.ActiveAccumulateParameterNtf(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
		return
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) PlayerActiveReceive(c *gin.Context) {
	var reqData proto.ProtoPlayerActiveReceiveReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "PlayerActiveReceive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.PlayerActiveReceive(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) GetActiveExchangeCost(c *gin.Context) {
	var reqData proto.ProtoGetActiveExchangeCostReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetActiveExchangeCost", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetActiveExchangeCost(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) CheckActiveConfig(c *gin.Context) {
	var reqData proto.ProtoCheckActiveConfigReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "CheckActiveConfig", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	resFuncParams.Param = handler.CheckActiveConfig(&reqData)
}

func (fw *FinanceWebModule) PlayerExchangeCDKey(c *gin.Context) {
	var reqData proto.ProtoPlayerExchangeCDKeyReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "PlayerExchangeCDKey", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.PlayerExchangeCDKey(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}

func (fw *FinanceWebModule) CheckPlayerActiveIsCompleted(c *gin.Context) {
	var reqData proto.ProtoCheckPlayerActiveIsCompletedReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "CheckPlayerActiveIsCompleted", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.CheckPlayerActiveIsCompleted(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
