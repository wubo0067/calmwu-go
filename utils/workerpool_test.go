/*
 * @Author: calmwu
 * @Date: 2018-12-10 11:06:42
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-12-10 11:45:24
 */

package utils

import (
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

var totalNum int
var callCount int

func processFunc(arg interface{}) error {
	totalNum += arg.(int)
	callCount++
	if callCount == 100 {
		ZLog.Debug("-----------complete!----------------")
	}
	return nil
}

func TestWokerPool(t *testing.T) {
	InitDefaultZapLog("workpool.log", zapcore.DebugLevel, 0)

	wp, err := StartWorkerPool(processFunc, 100, 3*time.Second)
	if err != nil {
		t.Error(err.Error())
		return
	}

	ZLog.Debug("-----------start!----------------")
	i := 0
	for i < 100 {
		if wp.Serve(i) {
			i++
		}
	}

	time.Sleep(10 * time.Second)

	wp.Stop()

	ZLog.Debugf("totalNum:%d callCount:%d", totalNum, callCount)
}
