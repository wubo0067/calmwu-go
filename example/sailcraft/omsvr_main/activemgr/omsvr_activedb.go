/*
 * @Author: calmwu
 * @Date: 2018-05-17 10:57:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 18:02:32
 * @Comment:
 */

package activemgr

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/omsvr_main/common"
	"sailcraft/omsvr_main/db"
	"sailcraft/omsvr_main/proto"

	"github.com/go-xorm/builder"
)

type activeRecordResult struct {
	records []db.TblActiveInstControlS
	err     error
	state   db.ActivePerformState
}

// 按状态查询活动配置
func QueryActiveInstsByState(performState db.ActivePerformState, engine *mysql.DBEngineInfoS) ([]db.TblActiveInstControlS, error) {
	result := make([]db.TblActiveInstControlS, 0)

	cond := builder.Expr(fmt.Sprintf("PerformState=%d", performState))
	err := mysql.FindRecordsByMultiConds(engine, db.TBNAME_ACTIVEINSTCTRL, &cond, 0, 0, &result)
	if err != nil {
		base.GLog.Error("Query %s ActivePerformState[%s] failed! reason[%s]",
			db.TBNAME_ACTIVEINSTCTRL, performState.String(), err.Error())
		return nil, err
	}

	return result, nil
}

// 设置活动状态
func SetActiveInstPerformState(recordID int64, performState db.ActivePerformState, engine *mysql.DBEngineInfoS) error {
	_, err := mysql.UpdateRecordSpecifiedFieldsByCond(engine, db.TBNAME_ACTIVEINSTCTRL,
		fmt.Sprintf("Id=%d", recordID),
		map[string]interface{}{
			"PerformState": performState})
	if err != nil {
		err := fmt.Errorf("recordID[%d] set %s.PerformState[%d] failed! reason[%s]",
			recordID, db.TBNAME_ACTIVEINSTCTRL, performState, err.Error())
		base.GLog.Critical(err.Error())
		return err
	} else {
		base.GLog.Debug("recordID[%d] set %s.PerformState[%d] successed!",
			recordID, db.TBNAME_ACTIVEINSTCTRL, performState)
	}
	return nil
}

//
func StopAllWaitingActiveInstCtrls(engine *mysql.DBEngineInfoS) {
	_, err := mysql.UpdateRecordSpecifiedFieldsByCond(engine, db.TBNAME_ACTIVEINSTCTRL,
		fmt.Sprintf("PerformState=%d", db.E_ACTIVEPERFORMSTATE_WAITING),
		map[string]interface{}{
			"PerformState": db.E_ACTVIEPERFROMSTATE_COMPLETED})
	if err != nil {
		err := fmt.Errorf("set Wait Record %s.PerformState[E_ACTVIEPERFROMSTATE_COMPLETED] failed! reason[%s]",
			db.TBNAME_ACTIVEINSTCTRL, err.Error())
		base.GLog.Critical(err.Error())
	} else {
		base.GLog.Debug("set Wait Record %s.PerformState[E_ACTVIEPERFROMSTATE_COMPLETED] successed!",
			db.TBNAME_ACTIVEINSTCTRL)
	}
}

// 插入活动实例
func AddActiveInstCtrl(activeInst *proto.ProtoActiveInstControlS, engine *mysql.DBEngineInfoS) error {
	tbActiveInst := new(db.TblActiveInstControlS)
	tbActiveInst.ActiveID = activeInst.ActiveID
	tbActiveInst.ActiveType = activeInst.ActiveType
	tbActiveInst.DurationMinutes = activeInst.DurationMinutes
	tbActiveInst.GroupID = activeInst.GroupID
	tbActiveInst.PerformState = db.E_ACTIVEPERFORMSTATE_WAITING
	tbActiveInst.StartTimeName = activeInst.StartTimeName
	tbActiveInst.TimeZone = activeInst.TimeZone
	tbActiveInst.ZoneID = activeInst.ZoneID
	tbActiveInst.ChannelName = activeInst.ChannelName

	affected, err := mysql.InsertRecord(common.GDBEngine, db.TBNAME_ACTIVEINSTCTRL, tbActiveInst)
	if err != nil {
		err = fmt.Errorf("Insert %s %+v failed! reason[%s]",
			db.TBNAME_ACTIVEINSTCTRL, tbActiveInst, err.Error())
		base.GLog.Error(err.Error)
		return err
	}
	base.GLog.Debug("Insert %s %+v successed! affected[%d]",
		db.TBNAME_ACTIVEINSTCTRL, tbActiveInst, affected)
	return nil
}
