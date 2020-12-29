/*
 * @Author: calmwu
 * @Date: 2019-08-10 11:52:08
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-08-10 11:54:11
 */

package main

import (
	"fmt"
	rd "runtime/debug"

	"github.com/pkg/errors"
	calm_utils "github.com/wubo0067/calmwu-go/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func debug(args ...interface{}) {
	calm_utils.ZLog.Debug(args...)
}

func debugf(template string, args ...interface{}) {
	calm_utils.ZLog.Debugf(template, args...)
}

func errorf(template string, args ...interface{}) {
	calm_utils.ZLog.Errorf(template, args...)
}

func callThree() error {
	debug("call Three")
	//panic("create panic")
	return errors.New("Create error")
}

func callSecond() error {
	debug("call Second")
	return callThree()
}

func callFirst() error {
	debug("call First")
	return callSecond()
}

type I interface {
	M()
}

type T struct {
	S string
}

func (t *T) M() {
	if t == nil {
		fmt.Println("<nil>")
		return
	}
	fmt.Println(t.S)
}

func backNilInterface() I {
	var t *T
	if t == nil {
		debug("t is nil")
	}
	var i I
	// 就算这里赋值一个nil指针，返回值也不是nil interface了
	i = t
	return i
}

//go:noinline
func read(i interface{}) {
	println(i)
	fmt.Println(i)
}

func main() {
	var i8 string = "2"
	println(i8)
	read(i8)

	calm_utils.InitDefaultZapLog("test.log", zapcore.DebugLevel, 1)

	defer func() {
		if err := recover(); err != nil {
			rd.PrintStack()
			debugf("%#s", calm_utils.CallStack(3))
		}
	}()

	err := callFirst()
	//debugf("err:%+v\n", err)
	errorf("----err---:%+v\n", err)
	fmt.Printf("err:%s\n", err)

	zl, _ := zap.NewDevelopment()
	zls := zl.Sugar()

	zls.Warnf("suger development: %s", "call second\n-------\n++++++++")

	// https://www.jianshu.com/p/28dc038d6c6b?utm_campaign=maleskine&utm_content=note&utm_medium=writer_share&utm_source=twitter
	i := backNilInterface()
	if i == nil {
		debug("i is a nil interface")
	} else {
		debug("i is not a nil interface")
		if t, ok := i.(*T); ok {
			if t == nil {
				debug("t is a nil")
			}
		}
	}
	return
}
