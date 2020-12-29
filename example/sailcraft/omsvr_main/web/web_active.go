/*
 * @Author: calmwu
 * @Date: 2018-05-18 12:20:01
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 14:08:09
 * @Comment:
 */

package web

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	financesvr_proto "sailcraft/financesvr_main/proto"
	"sailcraft/omsvr_main/activemgr"
	"sailcraft/omsvr_main/common"
	"sailcraft/omsvr_main/db"
	"sailcraft/omsvr_main/proto"

	"github.com/gin-gonic/gin"
	"github.com/go-xorm/builder"
)

func (fw *OMSWebModule) AddActiveInsts(c *gin.Context) {
	var reqData proto.ProtoAddActiveInstCtrlsReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "AddActiveInsts", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	for index := range reqData.ActiveInstCtrls {
		err = activemgr.AddActiveInstCtrl(&reqData.ActiveInstCtrls[index], common.GDBEngine)
		if err != nil {
			break
		}
	}

	if err != nil {
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		resFuncParams.Param = failInfo
		resFuncParams.RetCode = -1
	}
}

func (fw *OMSWebModule) ReloadActiveInsts(c *gin.Context) {
	var reqData proto.ProtoLoadWatingActiveInstCtrlsReq
	_, responseFunc, err := base.RequestPretreatment(c, "ReloadActiveInsts", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	base.GLog.Debug("Uin[%d] Reload ActiveInstsCtrl", reqData.Uin)
	activemgr.GActiveInstCtrlMgr.Reload()
}

func (fw *OMSWebModule) CleanAllActiveInst(c *gin.Context) {
	var reqData proto.ProtoCleanAllActiveInstCtrlsReq
	_, responseFunc, err := base.RequestPretreatment(c, "CleanAllActiveInst", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	base.GLog.Debug("Uin[%d] Clean ActiveInstsCtrl", reqData.Uin)
	activemgr.GActiveInstCtrlMgr.Clean()
}

func (fw *OMSWebModule) QueryRunningActiveTypes(c *gin.Context) {
	var reqData proto.ProtoQueryRunningActiveTypesReq
	resFuncParams, responseFunc, err := base.RequestPretreatment(c, "OpenActive", &reqData)
	defer responseFunc()
	if err != nil {
		return
	}

	var resData proto.ProtoQueryRunningActiveTypesRes
	resData.Uin = reqData.Uin
	resData.ZoneID = reqData.ZoneID

	// 去数据库查询开启的活动
	result := make([]db.TblActiveInstControlS, 0)

	cond := builder.Expr(fmt.Sprintf("PerformState=%d", db.E_ACTIVEPERFORMSTATE_RUNNING))
	err = mysql.FindDistinctRecordsByMultiConds(common.GDBEngine, db.TBNAME_ACTIVEINSTCTRL, []string{"ActiveType"}, &cond, 0, 0, &result)
	if err != nil {
		base.GLog.Error("Query %s ActivePerformState[E_ACTIVEPERFORMSTATE_RUNNING] failed! reason[%s]",
			db.TBNAME_ACTIVEINSTCTRL, err.Error())
	} else {
		base.GLog.Debug("query record count:%d", len(result))
		for index := range result {
			recordActiveInstControl := &result[index]
			resData.RunningActiveTypes = append(resData.RunningActiveTypes,
				recordActiveInstControl.ActiveType)
		}
	}

	// 去financesvr查询七日、首冲是否结束
	financeSvrReq := financesvr_proto.ProtoCheckPlayerActiveIsCompletedReq{
		Uin:    reqData.Uin,
		ZoneID: int32(reqData.ZoneID),
	}
	financeSvrRes, err := common.SendReqToFinanceSvr("CheckPlayerActiveIsCompleted", financeSvrReq)
	if err == nil {
		var completedRes financesvr_proto.ProtoCheckPlayerActiveIsCompletedRes
		err := base.MapstructUnPackByJsonTag(financeSvrRes.ResData.Params, &completedRes)
		if err == nil {
			for index := range completedRes.ActiveCompleteLst {
				if completedRes.ActiveCompleteLst[index].IsCompleted == 0 {
					resData.RunningActiveTypes = append(resData.RunningActiveTypes,
						completedRes.ActiveCompleteLst[index].ActiveType)
				}
			}
		}
	}

	resFuncParams.Param = resData
}
