/*
 * @Author: calmwu
 * @Date: 2018-02-02 15:09:53
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 19:15:56
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/financesvr_main/handler"
	"sailcraft/financesvr_main/proto"

	"github.com/gin-gonic/gin"
)

func (fw *FinanceWebModule) NewFinanceUser(c *gin.Context) {
	var reqData proto.ProtoNewFinanceUserReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "NewFinanceUser", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	var resData proto.ProtoNewFinanceUserRes
	resData.Uin = reqData.Uin
	resData.ZoneID = reqData.ZoneID
	resData.Result = 0

	err = handler.AddNewFinanceUser(&reqData)
	if err != nil {
		resData.Result = -1
	}

	resFuncParams.Param = resData
}

// 查询用户消费类型，普通用户、月卡用户、普通月卡用户
func (fw *FinanceWebModule) QueryUserVIPType(c *gin.Context) {
	var reqData proto.ProtoUserVIPTypeReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "QueryUserVIPType", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	var resData proto.ProtoUserVIPTypeRes
	resData.Uin = reqData.Uin
	resData.ZoneID = reqData.ZoneID
	resData.Result = 0

	userFinance, err := handler.QueryFinanceUser(reqData.Uin)
	if err != nil {
		resData.Result = -1
	} else {
		resData.UserVIPType = handler.GetFinanceUserVIPType(userFinance)
		resData.TimeZone = userFinance.TimeZone
	}

	resFuncParams.Param = resData
	resFuncParams.RetCode = resData.Result
}

func (fw *FinanceWebModule) GetUserFinanceBusinessRedLights(c *gin.Context) {
	var reqData proto.ProtoGetFinanceBusinessRedLightsReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "GetUserFinanceBusinessRedLights", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	hRes, err := handler.GetUserFinanceBusinessRedLights(&reqData)
	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	} else {
		resFuncParams.Param = hRes
	}
}
