/*
 * @Author: calmwu
 * @Date: 2018-05-19 10:21:18
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 10:31:08
 */

package utils

import (
	"reflect"
)

func GetTypeName(obj interface{}) (name1, name2, name3 string) {
	name1 = reflect.ValueOf(obj).Type().Name()
	objType := reflect.Indirect(reflect.ValueOf(obj)).Type()
	name2 = objType.String()
	name3 = objType.Name()
	return
}
