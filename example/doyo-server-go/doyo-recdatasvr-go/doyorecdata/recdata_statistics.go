/*
 * @Author: calmwu
 * @Date: 2018-11-06 15:57:14
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-09 14:44:34
 */

package doyorecdata

import (
	base "doyo-server-go/doyo-base-go"
	"sync"
	"sync/atomic"
	"time"
)

type DoyoRecDataStatistics struct {
	currRunningCmdCount  int32
	totalProcessCmdCount uint64
	exitChan             chan struct{}
	exitWait             sync.WaitGroup
}

var (
	recDataStatistics *DoyoRecDataStatistics
	once              sync.Once
)

func InitDoyoRecDataStatistics() *DoyoRecDataStatistics {
	once.Do(func() {
		recDataStatistics = new(DoyoRecDataStatistics)
		recDataStatistics.exitChan = make(chan struct{})
		recDataStatistics.exitWait.Add(1)
		go recDataStatistics.statisticsRoutine()
	})
	return recDataStatistics
}

func (drds *DoyoRecDataStatistics) Stop() {
	close(drds.exitChan)
	drds.exitWait.Wait()
}

func (drds *DoyoRecDataStatistics) incRunningCmdCount() {
	atomic.AddInt32(&drds.currRunningCmdCount, 1)
	atomic.AddUint64(&drds.totalProcessCmdCount, 1)
}

func (drds *DoyoRecDataStatistics) decRunningCmdCount() {
	atomic.AddInt32(&drds.currRunningCmdCount, -1)
}

func (drds *DoyoRecDataStatistics) statisticsRoutine() {
	base.ZLog.Debug("statisticsRoutine running")

	defer func() {
		drds.exitWait.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("statisticsRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	var preTotalProcessCmdCount uint64
	statisticsOutputTicker := time.NewTicker(time.Minute)

L:
	for {
		select {
		case <-drds.exitChan:
			statisticsOutputTicker.Stop()
			base.ZLog.Info("statisticsRoutine receive exit noitfy")
			break L
		case <-statisticsOutputTicker.C:
			gap := drds.totalProcessCmdCount - preTotalProcessCmdCount
			preTotalProcessCmdCount = drds.totalProcessCmdCount
			base.ZLog.Debugf("currRunningCmdCount[%d] totalProcessCmdCount[%d] flowSpeed[%d]",
				drds.currRunningCmdCount, drds.totalProcessCmdCount, gap/60)
		}
	}
	base.ZLog.Debug("statisticsRoutine exit!")
}
