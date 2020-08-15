/*
 * @Author: calmwu
 * @Date: 2018-01-27 16:59:37
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 20:04:10
 * @Comment:
 */

// Package utils for golang tools functions
package utils

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/snwfdhmp/errlog"
)

// NewError 构建错误对象
func NewError(args ...interface{}) error {
	var err error
	var rawData []interface{}
	for _, fromArg := range args {
		switch toArg := fromArg.(type) {
		case error:
			err = toArg
			continue
		default:
			rawData = append(rawData, toArg)
		}
	}
	if err == nil {
		err = errors.Errorf("%v", rawData)
	}
	return errors.Errorf("%v [error => %s]", rawData, err.Error())
}

// IsInterfaceNil 判断接口是否为空接口
func IsInterfaceNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

var (
	// DefaultErrCheck error判断和列出上下文
	DefaultErrCheck = errlog.NewLogger(&errlog.Config{
		// PrintFunc is of type `func (format string, data ...interface{})`
		// so you can easily implement your own logger func.
		// In this example, logrus is used, but any other logger can be used.
		// Beware that you should add '\n' at the end of format string when printing.
		PrintFunc:          Debugf,
		PrintSource:        true,  //Print the failing source code
		LinesBefore:        2,     //Print 2 lines before failing line
		LinesAfter:         1,     //Print 1 line after failing line
		PrintError:         true,  //Print the error
		PrintStack:         false, //Don't print the stack trace
		ExitOnDebugSuccess: false, //Exit if err
	})
)

func NewDefaultErrCheck(printFunc func(format string, data ...interface{}), printStack bool) {
	DefaultErrCheck = nil
	DefaultErrCheck = errlog.NewLogger(&errlog.Config{
		// PrintFunc is of type `func (format string, data ...interface{})`
		// so you can easily implement your own logger func.
		// In this example, logrus is used, but any other logger can be used.
		// Beware that you should add '\n' at the end of format string when printing.
		PrintFunc:          printFunc,
		PrintSource:        true,       //Print the failing source code
		LinesBefore:        2,          //Print 2 lines before failing line
		LinesAfter:         1,          //Print 1 line after failing line
		PrintError:         true,       //Print the error
		PrintStack:         printStack, //Don't print the stack trace
		ExitOnDebugSuccess: false,      //Exit if err
	})
}
