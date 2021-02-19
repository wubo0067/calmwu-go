/*
 * @Author: CALM.WU
 * @Date: 2021-02-18 14:53:52
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-02-18 17:37:10
 */

package utils

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNoSliceType = errors.New("not slice type")

	ErrNoFuncType = errors.New("no func type")

	ErrFuncSignatureInvalid = errors.New("func signature invalid")

	ErrFuncInTypeNoMatch = errors.New("func in parameter type not match with slice elem")

	ErrFuncOutTypeNoMatch = errors.New("func out parameter type not match")

	ErrSliceNotPtr = errors.New("slice not ptr")

	boolType = reflect.ValueOf(true).Type()
)

//--------------------Map

// Apply takes a slice of type []T and a function of type func(T) T.
func Apply(slice, function interface{}) (interface{}, error) {
	return apply(slice, function, false)
}

// ApplyInPlace is like Apply, not allocated slice
func ApplyInPlace(slice, function interface{}) error {
	_, err := apply(slice, function, true)
	return err
}

//--------------------Filter

// Choose 选择function返回true的元素, 返回一个新分配的slice
func Choose(slice, function interface{}) (interface{}, error) {
	r, _, err := chooseOrDrop(slice, function, false, true)
	return r, err
}

// Drop 选择function返回false的元素, 返回一个新分配的slice
func Drop(slice, function interface{}) (interface{}, error) {
	r, _, err := chooseOrDrop(slice, function, false, false)
	return r, err
}

// ChooseInPlace 选择function返回true的元素
func ChooseInPlace(slice, function interface{}) error {
	return chooseOrDropInPlace(slice, function, true)
}

// DropInPlace 选择function返回false的元素
func DropInPlace(slice, function interface{}) error {
	return chooseOrDropInPlace(slice, function, false)
}

//--------------------Reduce，汇聚，

//  Reduce ...
//	func multiply(a, b int) int { return a*b }
//	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//	factorial := Reduce(a, multiply, 1).(int)
func Reduce(slice, pairFunction, zero interface{}) (interface{}, error) {
	in := reflect.ValueOf(slice)
	if in.Kind() != reflect.Slice {
		return nil, ErrNoSliceType
	}

	n := in.Len()
	switch n {
	case 0:
		return zero, nil
	case 1:
		return in.Index(0), nil
	}

	// slice元素类型
	elemType := in.Type().Elem()
	fn := reflect.ValueOf(pairFunction)
	// 判断reduce的函数是否符合签名
	if err := verifyFunc(fn, elemType, elemType, elemType); err != nil {
		str := elemType.String()
		return nil, fmt.Errorf("reduce: function must be type func(%s, %s) %s", str, str, str)
	}

	// 函数输入参数
	var ins [2]reflect.Value
	ins[0] = reflect.ValueOf(zero)
	ins[1] = in.Index(0)
	out := fn.Call(ins[:])[0]

	for i := 1; i < n; i++ {
		ins[0] = out
		ins[1] = in.Index(i)
		out = fn.Call(ins[:])[0]
	}
	return out.Interface(), nil
}

// chooseOrDropInPlace 在原有slice上进行操作，由于要修改长度，所以必须传入slice pointer
func chooseOrDropInPlace(slice, function interface{}, truth bool) error {
	inp := reflect.ValueOf(slice)
	if inp.Kind() != reflect.Ptr {
		return ErrSliceNotPtr
	}

	_, n, err := chooseOrDrop(inp.Elem().Interface(), function, true, truth)
	if err != nil {
		return err
	}
	inp.Elem().SetLen(n)
	return nil
}

// 选择 function = truth的元素
func chooseOrDrop(slice, function interface{}, inPlace, truth bool) (interface{}, int, error) {
	if strSlice, ok := slice.([]string); ok {
		if strFn, ok := function.(func(string) bool); ok {
			var r []string
			if inPlace {
				// 保证了容量不变，长度变为0
				r = strSlice[:0]
			}
			for _, s := range strSlice {
				if strFn(s) == truth {
					r = append(r, s)
				}
			}
			return r, len(r), nil
		}
	}

	in := reflect.ValueOf(slice)
	if in.Kind() != reflect.Slice {
		return nil, -1, ErrNoSliceType
	}
	fn := reflect.ValueOf(function)
	// slice元素类型
	elemType := in.Type().Elem()
	if err := verifyFunc(fn, elemType, boolType); err != nil {
		return nil, -1, err
	}

	var which []int
	var ins [1]reflect.Value
	for i := 0; i < in.Len(); i++ {
		ins[0] = in.Index(i)
		if fn.Call(ins[:])[0].Bool() == truth {
			which = append(which, i)
		}
	}

	out := in
	if !inPlace {
		out = reflect.MakeSlice(in.Type(), len(which), len(which))
	}

	for i := range which {
		out.Index(i).Set(in.Index(which[i]))
	}
	return out.Interface(), len(which), nil
}

func apply(slice, function interface{}, inPlace bool) (interface{}, error) {
	// 对[]string, func(string) string，直接处理
	if strSlice, ok := slice.([]string); ok {
		if strFn, ok := function.(func(string) string); ok {
			r := strSlice
			if !inPlace {
				// 生成新的slice
				r = make([]string, len(strSlice))
			}
			for i, s := range strSlice {
				r[i] = strFn(s)
			}
			return r, nil
		}
	}

	// 输入的slice
	in := reflect.ValueOf(slice)
	if in.Kind() != reflect.Slice {
		return nil, ErrNoSliceType
	}

	fn := reflect.ValueOf(function)
	// slice 元素类型
	elemType := in.Type().Elem()

	if err := verifyFunc(fn, elemType, nil); err != nil {
		return nil, err
	}

	out := in
	if !inPlace {
		out = reflect.MakeSlice(reflect.SliceOf(fn.Type().Out(0)), in.Len(), in.Len())
	}

	var ins [1]reflect.Value
	for i := 0; i < in.Len(); i++ {
		// 函数输入参数
		ins[0] = in.Index(i)
		// 设置elem值
		out.Index(i).Set(fn.Call(ins[:])[0])
	}
	return out.Interface(), nil
}

// verifyFunc 判断函数签名是否合法，参数个数是否正确、类型是否匹配
// 最后一个type是函数返回值类型，其余的是输入参数
// 如果最后一个类型是nil，不用严格判断
func verifyFunc(fn reflect.Value, types ...reflect.Type) error {
	// 判断是否是函数
	if fn.Kind() != reflect.Func {
		return ErrNoFuncType
	}

	// 判断入参、出参个数和类型数量是否匹配
	if fn.Type().NumIn() != len(types)-1 || fn.Type().NumOut() != 1 {
		return ErrFuncSignatureInvalid
	}

	// 检查函数入参类型是否匹配
	for i := 0; i < len(types)-1; i++ {
		if fn.Type().In(i) != types[i] {
			return ErrFuncInTypeNoMatch
		}
	}

	outType := types[len(types)-1]
	if outType != nil && fn.Type().Out(0) != outType {
		return ErrFuncOutTypeNoMatch
	}

	return nil
}
