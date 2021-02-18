/*
 * @Author: CALM.WU
 * @Date: 2021-02-18 17:37:32
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-02-18 17:44:43
 */

package utils

import "testing"

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
