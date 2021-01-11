/*
 * @Author: CALM.WU
 * @Date: 2021-01-08 11:43:44
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-01-08 17:41:14
 */

package utils

import (
	"reflect"
	"testing"
)

type Cat struct{}

func TestReflectTV(t *testing.T) {
	c := &Cat{}
	var ic interface{} = c
	t.Logf("check is is nil, %v", IsNil(ic))

	// typeof: *utils.Cat, valueof: &{}
	t.Logf("typeof: %v, valueof: %v", reflect.TypeOf(ic), reflect.ValueOf(ic))

	// typeof kind: ptr, valueof kind: ptr
	t.Logf("typeof kind: %v, valueof kind: %v", reflect.TypeOf(ic).Kind(), reflect.ValueOf(ic).Kind())
}
