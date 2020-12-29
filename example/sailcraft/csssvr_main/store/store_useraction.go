/*
 * @Author: calmwu
 * @Date: 2018-06-13 11:42:36
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 11:54:08
 * @Comment:
 */
package store

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/csssvr_main/proto"

	"github.com/mitchellh/mapstructure"
)

func processUserActionReport(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	uin := cpd.ReqData.Uin

	var userActionReportNtf proto.ProtoUserActionReportNtf
	err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &userActionReportNtf)
	if err == nil {
		session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
		if session == nil {
			return
		}

		date := base.GetDate()

		base.GLog.Debug("uin[%d] userActionReportNtf:%+v", uin, userActionReportNtf)
		cqlUpdateUserActionStatis := fmt.Sprintf("UPDATE tbl_UserActionStatis set PerformCount=PerformCount+1 WHERE date='%s' AND ActionName='%s'",
			date, userActionReportNtf.ActionName)
		execCql(session, cqlUpdateUserActionStatis)

		// 对钻石消耗进行分类统计
		if userActionReportNtf.DiamondCostCount > 0 {
			cqlUpdateDiamondCostTypeStatis := fmt.Sprintf("UPDATE tbl_DiamondCostTypeStatis set PerformCount=PerformCount+1, TotalDiamondCost=TotalDiamondCost+%d WHERE date='%s' AND ActionName='%s'",
				userActionReportNtf.DiamondCostCount, date, userActionReportNtf.ActionName)
			execCql(session, cqlUpdateDiamondCostTypeStatis)
		}

	} else {
		base.GLog.Error("Decode UserActionReportNtf params ===> userActionReportNtf failed! reason[%s]",
			err.Error())
	}
}
