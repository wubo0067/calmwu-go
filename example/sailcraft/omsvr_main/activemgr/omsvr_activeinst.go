/*
 * @Author: calmwu
 * @Date: 2018-05-17 10:59:51
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 20:50:22
 * @Comment:
 */

package activemgr

import (
	"sailcraft/base"
	financesvr_proto "sailcraft/financesvr_main/proto"
	"sailcraft/omsvr_main/common"
	"sailcraft/omsvr_main/db"
	"time"
)

// 活动实例控制类型
type activeInstCtrl struct {
	StartTime time.Time                 // 活动启动的时间
	EndTime   time.Time                 // 活动结束时间
	Record    *db.TblActiveInstControlS // 配置的实例
	ExitChan  chan struct{}
	aicm      *ActiveInstCtrlMgr
}

func ActiveInstCtrlCompare(a, b interface{}) int {
	lActiveInst := a.(*activeInstCtrl)
	rActiveInst := b.(*activeInstCtrl)

	lStartSecs := lActiveInst.StartTime.Unix()
	rStartSecs := rActiveInst.StartTime.Unix()

	switch {
	case lStartSecs > rStartSecs:
		return 1
	case lStartSecs < rStartSecs:
		return -1
	default:
		return 0
	}
}

func InitActiveInstCtrl(record *db.TblActiveInstControlS, aicm *ActiveInstCtrlMgr) *activeInstCtrl {
	if record != nil {
		now, err := base.GetTimeByTz(record.TimeZone)
		if err != nil {
			base.GLog.Error("RecordID[%d] ActiveInst[%d] ZoneID[%d] TimeZone[%s] is invalid!",
				record.Id, record.ActiveID, record.ZoneID, record.TimeZone)
			return nil
		}

		var activeInst activeInstCtrl
		localtion, _ := time.LoadLocation(record.TimeZone)
		activeInst.StartTime, err = time.ParseInLocation("2006-01-02 15:04:05", record.StartTimeName, localtion)
		activeInst.EndTime = activeInst.StartTime.Add(time.Duration(record.DurationMinutes) * time.Minute)
		activeInst.Record = record
		activeInst.ExitChan = make(chan struct{})
		activeInst.aicm = aicm

		if record.PerformState == db.E_ACTIVEPERFORMSTATE_WAITING && now.After(activeInst.EndTime) {
			base.GLog.Error("RecordID[%d] ActiveInst[%d] ZoneID[%d] endTime[%s] has expired! now[%s]",
				record.Id, record.ActiveID, record.ZoneID, base.TimeName(activeInst.EndTime), base.TimeName(*now))
		}

		base.GLog.Debug("RecordID[%d] ActiveInst[%d] ZoneID[%d] startTime[%s] endTime[%s]", record.Id, record.ActiveID,
			record.ZoneID, base.TimeName(activeInst.StartTime), base.TimeName(activeInst.EndTime))
		return &activeInst
	}
	return nil
}

// 活动启动
func (aic *activeInstCtrl) Running() {
	//
	err := SetActiveInstPerformState(aic.Record.Id, db.E_ACTIVEPERFORMSTATE_RUNNING, common.GDBEngine)
	if err != nil {
		base.GLog.Error("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] set performState[E_ACTIVEPERFORMSTATE_RUNNING] failed",
			aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String())
	}

	// 发送开启命令
	aic.financeSvrOpenActive()

	// 定时关闭
	go func() {
		now, _ := base.GetTimeByTz(aic.Record.TimeZone)

		if now.Equal(aic.EndTime) || now.After(aic.EndTime) {
			base.GLog.Warn("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] direct to closed!",
				aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String())
			aic.Close()
			return
		}

		durationTime := aic.EndTime.Sub(*now)
		closeTimer := time.NewTimer(durationTime)

		base.GLog.Debug("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] running, monitor this expiration",
			aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String())
	L:
		for {
			select {
			case <-closeTimer.C:
				base.GLog.Warn("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] due to closed!",
					aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String())
				aic.Close()
			case <-aic.ExitChan:
				base.GLog.Warn("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] force closed!",
					aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String())
				aic.Close()
				closeTimer.Stop()
				break L
			}
		}
	}()

	return
}

// 判断活动是否已经过期
func (aic *activeInstCtrl) isExpired() bool {
	now, err := base.GetTimeByTz(aic.Record.TimeZone)
	if err != nil {
		base.GLog.Error("RecordID[%d] ActiveInst[%d] ZoneID[%d] TimeZone[%s] is invalid!",
			aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.TimeZone)
		return false
	}

	// 当前时间已经超过了活动结束时间
	if now.After(aic.EndTime) {
		base.GLog.Warn("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] endTime[%s] has expired!",
			aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String(), base.TimeName(aic.EndTime))
		return true
	}

	return false
}

// 判断活动是否可以执行
func (aic *activeInstCtrl) CanStart() bool {
	now, _ := base.GetTimeByTz(aic.Record.TimeZone)

	if now.Equal(aic.StartTime) || now.After(aic.StartTime) {
		base.GLog.Warn("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] startTime[%s] can start!",
			aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String(), base.TimeName(aic.StartTime))
		return true
	}
	base.GLog.Debug("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] startTime[%s] can't start!",
		aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String(), base.TimeName(aic.StartTime))
	return false
}

func (aic *activeInstCtrl) Close() {
	err := SetActiveInstPerformState(aic.Record.Id, db.E_ACTVIEPERFROMSTATE_COMPLETED, common.GDBEngine)
	if err != nil {
		base.GLog.Error("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] set performState[E_ACTVIEPERFROMSTATE_COMPLETED] failed",
			aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String())
	}

	// TODO: 发送关闭命令
	aic.financeSvrCloseActive()
	aic.aicm.activeInstCtrlCloseNtfChan <- aic.Record.Id

	base.GLog.Debug("RecordID[%d] ActiveInst[%d] ZoneID[%d] ActiveType[%s] finished!",
		aic.Record.Id, aic.Record.ActiveID, aic.Record.ZoneID, aic.Record.ActiveType.String())
}

func (aic *activeInstCtrl) ForceClose() {
	close(aic.ExitChan)
}

func (aic *activeInstCtrl) financeSvrOpenActive() {
	realReq := financesvr_proto.ProtoOpenActiveReq{
		Uin:    10000000,
		ZoneID: int32(aic.Record.ZoneID),
		ActiveControlConfigs: []financesvr_proto.ProtoActiveControlInfoS{
			financesvr_proto.ProtoActiveControlInfoS{
				ActiveType:   aic.Record.ActiveType,
				ActiveID:     aic.Record.ActiveID,
				ChannelID:    aic.Record.ChannelName,
				StartTime:    aic.StartTime.Unix(),
				DurationSecs: int64(aic.Record.DurationMinutes * 60),
			},
		},
	}

	common.SendReqToFinanceSvr("OpenActive", realReq)
}

func (aic *activeInstCtrl) financeSvrCloseActive() {
	realReq := financesvr_proto.ProtoCloseActiveReq{
		Uin:        10000000,
		ZoneID:     int32(aic.Record.ZoneID),
		ActiveType: aic.Record.ActiveType,
		ActiveIDs:  []int{aic.Record.ActiveID},
	}

	common.SendReqToFinanceSvr("CloseActive", realReq)
}
