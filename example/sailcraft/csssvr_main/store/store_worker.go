/*
 * @Author: calmwu
 * @Date: 2018-01-11 10:47:27
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 11:55:32
 */

package store

import (
	"sailcraft/base"
	"sailcraft/csssvr_main/proto"
	"sync"
)

type CassandraWorkerMgr struct {
	cassandraProcDataChan         chan *proto.CassandraProcDataS
	cassandraDailyRevenueDataChan chan *proto.CassandraProcDataS
	exitChan                      chan struct{}
	processWaitGroup              *sync.WaitGroup
}

func (cwm *CassandraWorkerMgr) Start(workerCount int) {
	cwm.cassandraProcDataChan = make(chan *proto.CassandraProcDataS, 1000)
	cwm.cassandraDailyRevenueDataChan = make(chan *proto.CassandraProcDataS, 2000)
	cwm.exitChan = make(chan struct{})
	cwm.processWaitGroup = new(sync.WaitGroup)
	var index int = 0
	for index < workerCount {
		go cassandraWorkerRoutine(cwm)
		cwm.processWaitGroup.Add(1)
		index++
	}
	go cassandraWorkerDailyRevenueRoutine(cwm)
	cwm.processWaitGroup.Add(1)

	return
}

func (cwm *CassandraWorkerMgr) Stop() {
	close(cwm.exitChan)
	cwm.processWaitGroup.Wait()
}

func (cwm *CassandraWorkerMgr) submitRequest(cpd *proto.CassandraProcDataS) {
	cwm.cassandraProcDataChan <- cpd
}

func cassandraWorkerRoutine(cwm *CassandraWorkerMgr) {
	base.GLog.Info("CassandraProcRoutine running")
	defer cwm.processWaitGroup.Done()
L:
	for {
		select {
		case cassandraProcData, ok := <-cwm.cassandraProcDataChan:
			if ok {
				base.GLog.Debug("call Interface[%s] process", cassandraProcData.ReqData.ReqData.InterfaceName)
				switch cassandraProcData.ReqData.ReqData.InterfaceName {
				case proto.APINAMECssSvrTuitionStepReport:
					processClientTuitionStepReport(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrGetBattleVideo:
					processGetBattleVideo(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrUploadBattleVideo:
					processSvrUploadBattleVideo(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrDeleteBattleVideo:
					processSvrDeleteBattleVideo(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrUserLogin:
					processUserLogin(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrUserLogout:
					processUserLogout(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrQueryPlayerGeo:
					processQueryPlayerGeo(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrUserRecharge:
					processUserRecharge(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrOldUserReceiveCompensation:
					oldUserReceiveCompensation(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrClientCDNResourceDownloadReport:
					processClientCDNResourceDownloadReport(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrUploadUserAction:
					processUserActionReport(cwm, cassandraProcData)
				//
				case proto.APINAMECssSvrQueryUserRechargeInfo:
					processGetUserRechargeInfo(cwm, cassandraProcData)
				default:
					base.GLog.Error("cassandraProcData Interface[%s] is not support!", cassandraProcData.ReqData.ReqData.InterfaceName)
				}
			}
		case <-cwm.exitChan:
			base.GLog.Warn("cassandraWorkerRoutine receive exit notify!")
			break L
		}
	}
	base.GLog.Warn("cassandraWorkerRoutine Exit!")
}

func cassandraWorkerDailyRevenueRoutine(cwm *CassandraWorkerMgr) {
	base.GLog.Info("CassandraSerialProcessDailyRevenue running")
	defer cwm.processWaitGroup.Done()

L:
	for {
		select {
		case _, ok := <-cwm.cassandraDailyRevenueDataChan:
			if ok {
				//processDailyRevenue(cwm, cassandraProcData)
			}
		case <-cwm.exitChan:
			base.GLog.Warn("CassandraSerialProcessDailyRevenue Exit")
			break L
		}
	}
}
