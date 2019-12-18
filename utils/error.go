/*
 * @Author: calmwu
 * @Date: 2018-01-27 16:59:37
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-17 13:45:49
 * @Comment:
 */

// Package utils for golang tools functions
package utils

import (
	"reflect"

	"github.com/pkg/errors"
)

// NewError 构建错误对象
func NewError(args ...interface{}) error {
	var err error
	var rawData []interface{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case error:
			err = arg.(error)
			continue
		default:
			rawData = append(rawData, arg)
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
