/*
 * @Author: calm.wu
 * @Date: 2019-07-29 10:05:17
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-07-29 10:22:08
 */

package main

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

type Basic struct {
	Vstring     string      `json:"Vstring"`
	Vint        int         `json:"Vint"`
	Vuint       uint        `json:"Vuint"`
	Vbool       bool        `json:"Vbool"`
	Vfloat      float64     `json:"Vfloat"`
	Vextra      string      `json:"Vextra"`
	Vsilent     bool        `json:"vsilent"`
	Vdata       interface{} `json:"Vdata"`
	VjsonInt    int         `json:"VjsonInt"`
	VjsonFloat  float64     `json:"VjsonFloat"`
	VjsonNumber json.Number `json:"VjsonNumber"`
}

func unsafePointerStruct() {
	i := 10

	var fp *float64 = (*float64)(unsafe.Pointer(&i))
	*fp = *fp * 2.399

	fmt.Println("unsafePointer i:", i)

	basic := new(Basic)
	pVstring := (*string)(unsafe.Pointer(basic))
	*pVstring = "Hello"

	pVjsonFloat := (*float64)(unsafe.Pointer(uintptr(unsafe.Pointer(basic)) + unsafe.Offsetof(basic.VjsonFloat)))
	*pVjsonFloat = 9.323
	fmt.Printf("basic:%+v\n", basic)
}

func unsafeRangeArray() {
	array := [...]int{0, 1, 2, 3, 4, 5}
	pointer := &array[0]
	for i := 0; i < len(array); i++ {
		fmt.Print(*pointer, " ")
		// Pointer需要转变为uintptr之后才能和sizeof进行运算
		pointer = (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(pointer)) + unsafe.Sizeof(array[0])))
	}
}

func testNameReturn() (i int) {
	i, j := func() (int, int) {
		return 10, 12
	}()
	fmt.Println(j)

	httpCode := 10
	defer func(code *int) {
		fmt.Printf("httpCode:%d\n", *code)
	}(&httpCode)

	httpCode = 999
	return
}

func main() {
	unsafePointerStruct()
	unsafeRangeArray()
	i := testNameReturn()
	fmt.Printf("i:%d \n", i)
}
