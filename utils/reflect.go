/*
 * @Author: calmwu
 * @Date: 2018-05-19 10:21:18
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 10:31:08
 */

// https://github.com/logrusorgru/gopb3any/blob/master/lis/lis.go

package utils

import (
	"errors"
	"reflect"
)

func GetTypeName(obj interface{}) (name1, name2, name3 string) {
	name1 = reflect.ValueOf(obj).Type().Name()
	objType := reflect.Indirect(reflect.ValueOf(obj)).Type()
	name2 = objType.String()
	name3 = objType.Name()
	return
}

var ErrNoOne = errors.New("lis.TypeRegister.Get: no one")

// TypeRegister - type register
type TypeRegister map[string]reflect.Type

// Ser registers new type
func (t TypeRegister) Set(i interface{}) {
	if reflect.ValueOf(i).Kind() != reflect.Ptr {
		panic(errors.New("TypeRegister.Set() argument must to be a pointer"))
	}
	t[reflect.TypeOf(i).String()] = reflect.TypeOf(i)
}

// Get element of type, if no one - err will be ErrNoOne
func (t TypeRegister) Get(name string) (interface{}, error) {
	if typ, ok := t[name]; ok {
		return reflect.New(typ.Elem()).Elem().Addr().Interface(), nil
	}
	return nil, ErrNoOne
}

// shared type register
var TypeReg = make(TypeRegister)
