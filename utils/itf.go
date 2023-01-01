/*
 * @Author: CALM.WU
 * @Date: 2021-01-08 11:35:39
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-01 11:57:38
 */

package utils

import "reflect"

// IsNil 检查interface是否为nil
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}

	switch reflect.ValueOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
