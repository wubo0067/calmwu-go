/*
 * @Author: calm.wu
 * @Date: 2019-03-15 17:33:37
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-03-15 17:37:53
 */

package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/cpu"
)

type Person struct {
	num1 uint64
	_    cpu.CacheLinePad
}

func testSizeof() {
	var p Person
	fmt.Printf("Person size:%d\n", unsafe.Sizeof(p))
}
