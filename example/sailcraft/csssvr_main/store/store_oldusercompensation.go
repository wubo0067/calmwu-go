/*
 * @Author: calmwu
 * @Date: 2018-05-10 13:25:41
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-10 16:59:07
 * @Comment:
 */

package store

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/csssvr_main/proto"

	"github.com/mitchellh/mapstructure"
)

func oldUserReceiveCompensation(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	cassandraProcResult := &proto.CassandraProcResultS{
		Ok:     false,
		Result: nil,
	}

	session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
	if session == nil {
		return
	}

	var reqData proto.ProtoOldUserReceiveCompensationReq
	err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &reqData)
	if err != nil {
		base.GLog.Error("Decode ProtoOldUserReceiveCompensationReq failed! reason[%s]", err.Error())
		return
	}

	var resData proto.ProtoOldUserReceiveCompensationRes
	resData.DeviceID = reqData.DeviceID
	resData.Result = -1
	cassandraProcResult.Result = &resData
	cassandraProcResult.Ok = true

	// 查询领取状态
	var receiveStatus int
	var level int
	cqlQueryStatus := fmt.Sprintf("SELECT compensationlevel, receivestatus FROM tbl_OldUserCompensation WHERE DeviceID='%s'", reqData.DeviceID)
	if err := session.Query(cqlQueryStatus).Scan(&level, &receiveStatus); err != nil {
		base.GLog.Error("Query DeviceID[%s] Status failed! reason[%s]", reqData.DeviceID, err.Error())
	} else {
		// 判断是否领取过
		base.GLog.Debug("DeviceID[%s] compensationlevel[%d] receiveStatus[%d]", reqData.DeviceID, level, receiveStatus)
		if receiveStatus == 0 {
			// 修改状态
			cqlUpdateStatus := fmt.Sprintf("UPDATE tbl_OldUserCompensation SET receivestatus=1 WHERE DeviceID='%s'", reqData.DeviceID)
			execCql(session, cqlUpdateStatus)
			resData.Result = 0
			resData.Level = level
		}
	}

	cpd.ResultChan <- cassandraProcResult
	base.GLog.Debug("DeviceID[%s] return!", reqData.DeviceID)
}
