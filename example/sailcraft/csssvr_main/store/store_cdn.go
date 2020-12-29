/*
 * @Author: calmwu
 * @Date: 2018-05-16 10:32:06
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-16 10:35:26
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

func processClientCDNResourceDownloadReport(cwm *CassandraWorkerMgr, cpd *proto.CassandraProcDataS) {
	isoCountry, _ := common.QueryGeoInfo(cpd.RemoteIP)
	date := base.GetDate()

	var cdnDownloadReport proto.ProtoClientCDNDownloadReportNtf
	err := mapstructure.Decode(cpd.ReqData.ReqData.Params, &cdnDownloadReport)
	if err == nil {
		session := CasMgr.GetSessionByKeyspace("ks_statisticmodule")
		if session == nil {
			return
		}

		base.GLog.Debug("cdnDownloadReport:%+v", cdnDownloadReport)

		cqlUpdateClientCDNDownloadStatis := fmt.Sprintf(`UPDATE tbl_DailyClientCDNDownloadStatis SET TotalElapseTime=TotalElapseTime+%d, 
			TotalDownloadCount=TotalDownloadCount+1, 
			TotalAttemptCount=TotalAttemptCount+%d 
			WHERE date='%s' AND ClientVersion='%s' AND Platform='%s' AND ISOCountryCode='%s' AND ResourceName='%s' AND ResourceID=%d`,
			cdnDownloadReport.ElapseTime, cdnDownloadReport.AttemptCount, date, cdnDownloadReport.ClientVersion,
			cdnDownloadReport.PlatformName, isoCountry, cdnDownloadReport.ResourceName, cdnDownloadReport.ResourceID)

		execCql(session, cqlUpdateClientCDNDownloadStatis)
	} else {
		base.GLog.Error("Decode ProtoClientCDNDownloadReportNtf params ===> cdnDownloadReport failed! reason[%s]",
			err.Error())
	}
	return
}
