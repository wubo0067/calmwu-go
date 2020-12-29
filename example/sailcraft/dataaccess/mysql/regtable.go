/*
 * @Author: calmwu
 * @Date: 2017-10-21 10:15:53
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-21 11:21:40
 */

package mysql

import (
	"fmt"
	"reflect"
	"sync"
)

var tableTypeRegistry sync.Map

func RegisterTableObj(tableObj interface{}) {
	// 得到类型
	tableObjType := reflect.Indirect(reflect.ValueOf(tableObj)).Type()
	tableObjFullName := tableObjType.String()
	tableObjName := tableObjType.Name()

	fmt.Println("tableObjFullName:", tableObjFullName)
	fmt.Println("tableObjName:", tableObjName)

	tableTypeRegistry.Store(tableObjFullName, tableObjType)
	tableTypeRegistry.Store(tableObjName, tableObjType)
}

func NewTableObj(tblStructName string) (interface{}, error) {
	typeV, exist := tableTypeRegistry.Load(tblStructName)
	if exist {
		if tableType, ok := typeV.(reflect.Type); ok {
			newTableObj := reflect.New(tableType)
			return newTableObj.Interface(), nil
		}
	}

	return nil, fmt.Errorf("tblStructName[%s] no registration", tblStructName)
}
