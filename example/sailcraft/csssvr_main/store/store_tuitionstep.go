/*
 * @Author: calmwu
 * @Date: 2018-01-11 15:20:12
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-16 18:46:06
 * @Comment:
 */

package store

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/csssvr_main/common"
	"sailcraft/csssvr_main/proto"

	"github.com/mitchellh/mapstructure"
)

func processClientTuitionStepReport(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {

	isoCountry, _ := common.QueryGeoInfo(cpd.RemoteIP)

	var tuitionStepReportParams proto.ProtoTuitionStepReportParamsS
	err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &tuitionStepReportParams)
	if err == nil {
		session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
		if session == nil {
			return
		}

		base.GLog.Debug("TuitionStepInfo:%+v", tuitionStepReportParams)

		cqlUpdateTuitionStepStatistics2 := fmt.Sprintf("UPDATE tbl_TuitionStepStatistics SET count=count+1 WHERE clientversion='%s' and stepid=%d and platform='%s' and channelname='%s' and ISOCountryCode='%s'",
			tuitionStepReportParams.ClientVersion, tuitionStepReportParams.StepId, tuitionStepReportParams.PlatformName, tuitionStepReportParams.ChannelName, isoCountry)
		execCql(session, cqlUpdateTuitionStepStatistics2)

	} else {
		base.GLog.Error("Decode TuitionStepReport params ===> tuitionStepReportParams failed! reason[%s]",
			err.Error())
	}
	return
}
