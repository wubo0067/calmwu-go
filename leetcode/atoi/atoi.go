/*
 * @Author: CALM.WU
 * @Date: 2021-03-18 14:55:03
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-18 18:23:38
 */

package main

import (
	"bytes"
	"fmt"
	"math"
	"unsafe"
)

const (
	// 无符号32位数上线限
	MaxUint32 = ^uint32(0)
	MinUint32 = 0
	// 有符号32位数上下限
	MaxInt32 = int(MaxUint32 >> 1)
	MinInt32 = -MaxInt32 - 1
)

var (
	_minNumASCII = 48
	_maxNumASCII = 57
)

func String2Bytes(s string) []byte {
	sh := (*[2]uintptr)(unsafe.Pointer(&s))
	// bh := reflect.SliceHeader{
	// 	Data: sh.Data,
	// 	Len:  sh.Len,
	// 	Cap:  sh.Len,
	// }
	bh := [3]uintptr{sh[0], sh[1], sh[1]}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func myAtoi(s string) int {
	if len(s) == 0 {
		return 0
	}

	result := 0
	sign := 0
	isNum := false
	nums := make([]int, 0, 200)

	bs := String2Bytes(s)
	bs = bytes.TrimSpace(bs)
	//bs = bytes.TrimPrefix(bs, []byte{'0'})

	fmt.Println(bs)

	for _, b := range bs {
		if b == '-' {
			if sign == 0 {
				sign = -1
			} else {
				return 0
			}
		} else if b == '+' {
			if sign == 0 {
				sign = 1
			} else {
				return 0
			}
		} else {
			bNum := int(b)
			if bNum >= _minNumASCII && bNum <= _maxNumASCII {
				nums = append(nums, bNum-_minNumASCII)
				isNum = true
			} else {
				if !isNum {
					return 0
				}
				// 后面的非数字字符都抛弃
				break
			}
		}
	}

	numWide := len(nums) - 1
	for _, n := range nums {
		result = result + n*int(math.Pow10(numWide))
		numWide--
	}

	if sign < 0 {
		result = 0 - result
	}

	if result <= MinInt32 {
		return MinInt32
	}

	if result >= MaxInt32 {
		return MaxInt32
	}

	return result
}

func main() {
	b := '9'
	fmt.Printf("'b' = %d\n", int(b))

	content := "  -42"
	fmt.Printf("{%s} ===> {%d}\n", content, myAtoi(content))

	content = "4193 with words"
	fmt.Printf("{%s} ===> {%d}\n", content, myAtoi(content))

	content = "words and 987"
	fmt.Printf("{%s} ===> {%d}\n", content, myAtoi(content))

	content = "-91283472332"
	fmt.Printf("{%s} ===> {%d}\n", content, myAtoi(content))

	content = "+-12"
	fmt.Printf("{%s} ===> {%d}\n", content, myAtoi(content))
}
