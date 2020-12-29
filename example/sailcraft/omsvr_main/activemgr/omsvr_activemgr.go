/*
 * @Author: calmwu
 * @Date: 2018-05-17 12:05:58
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 19:14:11
 * @Comment:
 */

package activemgr

import (
	"sailcraft/base"
	"sailcraft/omsvr_main/common"
	"sailcraft/omsvr_main/db"
	"sync"
	"time"

	"github.com/emirpasic/gods/trees/binaryheap"
)

type ActiveInstCtrlMgrCmd int

type ActiveInstCtrlMgrCmdInfo struct {
	Cmd  ActiveInstCtrlMgrCmd
	Data interface{}
}

const (
	E_CMD_START ActiveInstCtrlMgrCmd = iota
	E_CMD_RELOAD
	E_CMD_EXIT
	E_CMD_CLEAN // 清除所有运行和待运行的活动控制
)

type ActiveInstCtrlMgr struct {
	activeInstWaitingHeap      *binaryheap.Heap              // 按startime.Unix()排序的执行最小堆
	activeInstRunningMap       map[int64]*activeInstCtrl     // 运行活动管理
	cmdChan                    chan ActiveInstCtrlMgrCmdInfo // 控制命令管道
	activeRecordsChan          chan *activeRecordResult      // 数据查询结果管道
	activeInstCtrlCloseNtfChan chan int64                    // 运行活动关闭通知通道
	routineWait                *sync.WaitGroup
}

var (
	GActiveInstCtrlMgr *ActiveInstCtrlMgr
)

func CreateActiveInstCtrlMgr() *ActiveInstCtrlMgr {
	if GActiveInstCtrlMgr == nil {
		GActiveInstCtrlMgr = new(ActiveInstCtrlMgr)
		GActiveInstCtrlMgr.activeInstWaitingHeap = binaryheap.NewWith(ActiveInstCtrlCompare)
		GActiveInstCtrlMgr.activeInstRunningMap = make(map[int64]*activeInstCtrl)
		GActiveInstCtrlMgr.cmdChan = make(chan ActiveInstCtrlMgrCmdInfo, 8)
		GActiveInstCtrlMgr.activeRecordsChan = make(chan *activeRecordResult, 2)
		GActiveInstCtrlMgr.activeInstCtrlCloseNtfChan = make(chan int64, 100)
		GActiveInstCtrlMgr.routineWait = new(sync.WaitGroup)

		go activeInstCtrlMgrRoutine(GActiveInstCtrlMgr)
		GActiveInstCtrlMgr.routineWait.Add(1)

		GActiveInstCtrlMgr.cmdChan <- ActiveInstCtrlMgrCmdInfo{
			Cmd:  E_CMD_START,
			Data: nil,
		}
	}

	return GActiveInstCtrlMgr
}

func (aicm *ActiveInstCtrlMgr) Reload() {
	aicm.cmdChan <- ActiveInstCtrlMgrCmdInfo{
		Cmd:  E_CMD_RELOAD,
		Data: nil,
	}
}

func (aicm *ActiveInstCtrlMgr) Clean() {
	aicm.cmdChan <- ActiveInstCtrlMgrCmdInfo{
		Cmd:  E_CMD_CLEAN,
		Data: nil,
	}
}

func activeInstCtrlMgrRoutine(aicm *ActiveInstCtrlMgr) {
	defer aicm.routineWait.Done()
	defer func() {
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			base.GLog.Error("activeInstCtrlMgrRoutine painc! reson:%v, stack:%s", err, base.GetCallStack())
		}
	}()
	base.GLog.Info("activeInstCtrlMgrRoutine running")
	// 定时检查活动是否开启
	ticker := time.NewTicker(3 * time.Second)

L:
	for {
		select {
		case cmdInfo, ok := <-aicm.cmdChan:
			if ok {
				base.GLog.Debug("cmd[%d] process", cmdInfo.Cmd)
				switch cmdInfo.Cmd {
				case E_CMD_START:
					go loadActiveRecords(aicm, db.E_ACTIVEPERFORMSTATE_WAITING, db.E_ACTIVEPERFORMSTATE_RUNNING)
				case E_CMD_RELOAD:
					go loadActiveRecords(aicm, db.E_ACTIVEPERFORMSTATE_WAITING)
				case E_CMD_CLEAN:
					// 清空待运行管理
					aicm.activeInstWaitingHeap = binaryheap.NewWith(ActiveInstCtrlCompare)
					// 所有待运行的状态改为E_ACTVIEPERFROMSTATE_COMPLETED
					StopAllWaitingActiveInstCtrls(common.GDBEngine)
					// 关闭所有在运行的活动
					for _, activeCtrl := range aicm.activeInstRunningMap {
						activeCtrl.Close()
					}
				case E_CMD_EXIT:
					base.GLog.Warn("Receive Exist Cmd")
					break L
				}
			}
		case <-ticker.C:
			aicm.checkWaitingActiveInstCtrls()
		case activeRecords, ok := <-aicm.activeRecordsChan:
			if ok {
				// 重新构建WAITING、RUNNING任务控制
				switch activeRecords.state {
				case db.E_ACTIVEPERFORMSTATE_WAITING:
					aicm.buildWaitingActiveCtrls(activeRecords)
				case db.E_ACTIVEPERFORMSTATE_RUNNING:
					aicm.buildRunningActiveCtrls(activeRecords)
				default:
					base.GLog.Error("activeRecords.state[%d] is invalid!", activeRecords.state)
				}
			}
		case recordID, ok := <-aicm.activeInstCtrlCloseNtfChan:
			if ok {
				if _, exist := aicm.activeInstRunningMap[recordID]; exist {
					base.GLog.Error("recordID[%d] delete from activeInstRunningMap", recordID)
					delete(aicm.activeInstRunningMap, recordID)
				} else {
					base.GLog.Error("recordID[%d] not in activeInstRunningMap", recordID)
				}
			}
		}
	}

	base.GLog.Info("activeMgrRoutine exit!")
}

// 服务启动，从数据库中加载WAITING、RUNNING的活动
func loadActiveRecords(aicm *ActiveInstCtrlMgr, states ...db.ActivePerformState) {
	for _, state := range states {
		result := new(activeRecordResult)
		result.state = state
		result.records, result.err = QueryActiveInstsByState(state, common.GDBEngine)
		if len(result.records) > 0 {
			base.GLog.Debug("Load ActiveInstCtrlRecord[%s] count[%d]", state.String(), len(result.records))
			aicm.activeRecordsChan <- result
		} else {
			base.GLog.Warn("Load ActiveInstCtrlRecord[%s] from db, result is empty", state.String())
		}
	}
}

// 判断有没有到期的活动
func (aicm *ActiveInstCtrlMgr) checkWaitingActiveInstCtrls() {
	if !aicm.activeInstWaitingHeap.Empty() {
		for {
			activeCtrlI, ok := aicm.activeInstWaitingHeap.Peek()
			if ok {
				activeInstCtrl := activeCtrlI.(*activeInstCtrl)
				if activeInstCtrl.CanStart() {
					// 活动开启
					activeInstCtrl.Running()
					aicm.activeInstWaitingHeap.Pop()
					aicm.activeInstRunningMap[activeInstCtrl.Record.Id] = activeInstCtrl
				} else {
					// 结束判断
					return
				}
			} else {
				base.GLog.Error("aicm.activeInstWaitingHeap.Peek() failed!")
				break
			}
		}
	}
	base.GLog.Debug("Check Waiting ActiveInstCtrl, now running ActiveInstCtrl count[%d]", len(aicm.activeInstRunningMap))
}

// 建立Waiting活动控制
func (aicm *ActiveInstCtrlMgr) buildWaitingActiveCtrls(activeRecords *activeRecordResult) {
	base.GLog.Debug("Wating ActiveInstCtrl record count[%d]", len(activeRecords.records))
	if activeRecords != nil && activeRecords.err == nil {
		activeInstWaitingHeap := binaryheap.NewWith(ActiveInstCtrlCompare)

		for index := range activeRecords.records {
			activeRecord := activeRecords.records[index]
			activeInstCtrl := InitActiveInstCtrl(&activeRecord, aicm)
			if activeInstCtrl != nil {
				base.GLog.Debug("RecordID[%d] ActiveInst[%d] ZoneID[%d]", activeInstCtrl.Record.Id,
					activeInstCtrl.Record.ActiveID, activeInstCtrl.Record.ZoneID)
				activeInstWaitingHeap.Push(activeInstCtrl)
			}
		}

		aicm.activeInstWaitingHeap = activeInstWaitingHeap
	}
}

// 建立Running活动控制
func (aicm *ActiveInstCtrlMgr) buildRunningActiveCtrls(activeRecords *activeRecordResult) {
	base.GLog.Debug("Running ActiveInstCtrl record count[%d]", len(activeRecords.records))
	if activeRecords != nil && activeRecords.err == nil {
		for index := range activeRecords.records {
			activeRecord := activeRecords.records[index]
			activeInstCtrl := InitActiveInstCtrl(&activeRecord, aicm)
			if activeInstCtrl != nil {
				// 重启活动控制
				base.GLog.Debug("RecordID[%d] ActiveInst[%d] ZoneID[%d]", activeInstCtrl.Record.Id,
					activeInstCtrl.Record.ActiveID, activeInstCtrl.Record.ZoneID)
				aicm.activeInstRunningMap[activeInstCtrl.Record.Id] = activeInstCtrl
				activeInstCtrl.Running()
			}
		}
	}
}
