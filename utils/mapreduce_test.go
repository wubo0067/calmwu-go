/*
 * @Author: CALM.WU
 * @Date: 2021-02-18 17:37:32
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-02-18 17:44:43
 */

package utils

import (
	"reflect"
	"testing"
)

func mul(a, b int) int {
	return a * b
}

func TestReduce(t *testing.T) {
	a := make([]int, 10)
	for i := range a {
		a[i] = i + 1
	}
	t.Log(a)
	// Compute 10!
	out, _ := Reduce(a, mul, 2)
	outNum := out.(int)
	expect := 2
	for i := range a {
		expect *= a[i]
	}
	if expect != outNum {
		t.Fatalf("expected %d got %d", expect, outNum)
	}
	t.Logf("outNum:%d expect:%d", outNum, expect)
}

func TestApplySliceInt(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	result, _ := Apply(a, func(i int) int {
		return i * 3
	})
	expect := []int{3, 6, 9, 12, 15, 18, 21, 24, 27}
	if !reflect.DeepEqual(result, expect) {
		t.Fatalf("Apply failed: expect %v result %v", expect, result)
	}
	t.Logf("Apply successed: result %v", result)
}

type Person struct {
	Name   string
	Salary int
}

func TestApplySliceStruct(t *testing.T) {
	a := []Person{
		{
			"1", 10,
		},
		{
			"2", 20,
		},
		{
			"3", 30,
		},
	}

	err := ApplyInPlace(a, func(p Person) Person {
		p.Salary = p.Salary * 10
		return p
	})

	if err != nil {
		t.Fatalf("ApplyInPlace failed. err: %s", err.Error())
	}
	t.Logf("ApplyInPlace successed. a: %v", a)

	b := []*Person{
		{
			"1", 10,
		},
		{
			"2", 20,
		},
		{
			"3", 30,
		},
	}

	err = ApplyInPlace(b, func(p *Person) *Person {
		p.Salary = p.Salary * 10
		return p
	})

	if err != nil {
		t.Fatalf("ApplyInPlace failed. err: %s", err.Error())
	}
	for i, p := range b {
		t.Logf("%d person:%v", i, *p)
	}
}
