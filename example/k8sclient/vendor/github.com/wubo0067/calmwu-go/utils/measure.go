/*
 * @Author: calm.wu
 * @Date: 2019-12-11 11:42:53
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-12-11 11:43:41
 */

package utils

import (
	"runtime"
	"time"
)

// MeasureFunc 函数的执行耗时测量
func MeasureFunc() func() {
	start := time.Now()
	pc, callFile, callLine, _ := runtime.Caller(1)
	callFuncName := runtime.FuncForPC(pc).Name()
	Debugf("Enter function[%s:%d %s]", callFile, callLine, callFuncName)

	return func() {
		Debugf("Exit function[%s:%d %s] after %s", callFile, callLine, callFuncName, time.Since(start))
	}
}
